package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"scaffold-api/internal/db/query"
	"scaffold-api/internal/errs"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

const (
	defaultPageSize = 20
	maxPageSize     = 100
	dateLayout      = "2006-01-02"
)

// CreateUserInput defines the accepted request body for creating a user.
type CreateUserInput struct {
	UID      string  `json:"uid" validate:"required,max=64" example:"user-001"`
	Name     string  `json:"name" validate:"required,max=255" example:"Alice"`
	Email    *string `json:"email" validate:"omitempty,email,max=255" example:"alice@example.com"`
	UsedName string  `json:"used_name" validate:"omitempty,max=255" example:"Ali"`
	Company  string  `json:"company" validate:"omitempty,max=255" example:"ACME"`
	Birth    *string `json:"birth" validate:"omitempty,datetime=2006-01-02" example:"1990-01-01"`
}

// UpdateUserInput defines the accepted request body for updating a user.
type UpdateUserInput struct {
	Name     string  `json:"name" validate:"required,max=255" example:"Alice"`
	Email    *string `json:"email" validate:"omitempty,email,max=255" example:"alice@example.com"`
	UsedName string  `json:"used_name" validate:"omitempty,max=255" example:"Ali"`
	Company  string  `json:"company" validate:"omitempty,max=255" example:"ACME"`
	Birth    *string `json:"birth" validate:"omitempty,datetime=2006-01-02" example:"1990-01-01"`
}

type ListUsersInput struct {
	Email    *string
	NameLike *string
	Page     int
	PageSize int
}

type UserPage struct {
	Items    []query.User
	Total    int64
	Page     int
	PageSize int
}

type UserStore interface {
	CreateUser(ctx context.Context, arg query.CreateUserParams) (query.User, error)
	GetUserByID(ctx context.Context, id int64) (query.User, error)
	ListUsers(ctx context.Context, arg query.ListUsersParams) ([]query.User, error)
	CountUsers(ctx context.Context, arg query.CountUsersParams) (int64, error)
	UpdateUser(ctx context.Context, arg query.UpdateUserParams) (query.User, error)
	DeleteUser(ctx context.Context, id int64) (int64, error)
}

type UserService struct {
	store UserStore
}

func NewUserService(store UserStore) *UserService {
	return &UserService{store: store}
}

func (s *UserService) Create(ctx context.Context, input CreateUserInput) (*query.User, error) {
	uid := strings.TrimSpace(input.UID)
	name := strings.TrimSpace(input.Name)
	if uid == "" || name == "" {
		return nil, errs.NewInvalidArgument("uid and name are required", map[string]string{
			"uid":  "uid is required",
			"name": "name is required",
		})
	}

	birth, err := parseBirth(input.Birth)
	if err != nil {
		return nil, err
	}

	user, err := s.store.CreateUser(ctx, query.CreateUserParams{
		Uid:      uid,
		Name:     name,
		Email:    nullableText(normalizeOptionalString(input.Email)),
		UsedName: strings.TrimSpace(input.UsedName),
		Company:  strings.TrimSpace(input.Company),
		Birth:    nullableDate(birth),
	})
	if err != nil {
		return nil, translateStoreError(err)
	}

	return &user, nil
}

func (s *UserService) GetByID(ctx context.Context, id int64) (*query.User, error) {
	user, err := s.store.GetUserByID(ctx, id)
	if err != nil {
		return nil, translateStoreError(err)
	}
	return &user, nil
}

func (s *UserService) List(ctx context.Context, input ListUsersInput) (*UserPage, error) {
	page := input.Page
	if page <= 0 {
		page = 1
	}

	pageSize := input.PageSize
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	filters := query.CountUsersParams{
		Email:    nullableText(normalizeOptionalString(input.Email)),
		NameLike: nullableText(normalizeOptionalString(input.NameLike)),
	}

	items, err := s.store.ListUsers(ctx, query.ListUsersParams{
		Email:    filters.Email,
		NameLike: filters.NameLike,
		Offset:   int32((page - 1) * pageSize),
		Limit:    int32(pageSize),
	})
	if err != nil {
		return nil, translateStoreError(err)
	}

	total, err := s.store.CountUsers(ctx, filters)
	if err != nil {
		return nil, translateStoreError(err)
	}

	return &UserPage{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *UserService) Update(ctx context.Context, id int64, input UpdateUserInput) (*query.User, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, errs.NewInvalidArgument("name is required", map[string]string{
			"name": "name is required",
		})
	}

	birth, err := parseBirth(input.Birth)
	if err != nil {
		return nil, err
	}

	user, err := s.store.UpdateUser(ctx, query.UpdateUserParams{
		ID:       id,
		Name:     name,
		Email:    nullableText(normalizeOptionalString(input.Email)),
		UsedName: strings.TrimSpace(input.UsedName),
		Company:  strings.TrimSpace(input.Company),
		Birth:    nullableDate(birth),
	})
	if err != nil {
		return nil, translateStoreError(err)
	}

	return &user, nil
}

func (s *UserService) Delete(ctx context.Context, id int64) error {
	rows, err := s.store.DeleteUser(ctx, id)
	if err != nil {
		return translateStoreError(err)
	}
	if rows == 0 {
		return errs.NewNotFound("user not found")
	}
	return nil
}

func parseBirth(value *string) (*time.Time, error) {
	if value == nil {
		return nil, nil
	}

	raw := strings.TrimSpace(*value)
	if raw == "" {
		return nil, nil
	}

	parsed, err := time.Parse(dateLayout, raw)
	if err != nil {
		return nil, errs.NewInvalidArgument("birth must be in YYYY-MM-DD format", map[string]string{
			"birth": "birth must use YYYY-MM-DD",
		})
	}

	return &parsed, nil
}

func normalizeOptionalString(value *string) *string {
	if value == nil {
		return nil
	}

	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}

	return &trimmed
}

func nullableText(value *string) pgtype.Text {
	if value == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *value, Valid: true}
}

func nullableDate(value *time.Time) pgtype.Date {
	if value == nil {
		return pgtype.Date{}
	}
	return pgtype.Date{Time: *value, Valid: true}
}

func translateStoreError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return errs.NewNotFound("user not found")
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == "23505" {
			return errs.NewConflict("user uid or email already exists")
		}
	}

	var appErr *errs.AppError
	if errors.As(err, &appErr) {
		return appErr
	}

	return errs.NewInternal(err)
}
