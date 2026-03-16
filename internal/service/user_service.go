package service

import (
	"context"
	"strings"
	"time"

	"scaffold-api/internal/errs"
	"scaffold-api/internal/model"
	"scaffold-api/internal/repository"
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
	Items    []model.User
	Total    int64
	Page     int
	PageSize int
}

type UserService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) Create(ctx context.Context, input CreateUserInput) (*model.User, error) {
	uid := strings.TrimSpace(input.UID)
	name := strings.TrimSpace(input.Name)
	if uid == "" || name == "" {
		return nil, errs.NewValidation("uid and name are required", map[string]string{
			"uid":  "uid is required",
			"name": "name is required",
		})
	}

	birth, err := parseBirth(input.Birth)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		UID:      uid,
		Name:     name,
		Email:    normalizeOptionalString(input.Email),
		UsedName: strings.TrimSpace(input.UsedName),
		Company:  strings.TrimSpace(input.Company),
		Birth:    birth,
	}

	return s.repo.Create(ctx, user)
}

func (s *UserService) GetByID(ctx context.Context, id uint64) (*model.User, error) {
	return s.repo.GetByID(ctx, id)
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

	items, total, err := s.repo.List(ctx, repository.UserListFilter{
		Email:    normalizeOptionalString(input.Email),
		NameLike: normalizeOptionalString(input.NameLike),
		Limit:    pageSize,
		Offset:   (page - 1) * pageSize,
	})
	if err != nil {
		return nil, err
	}

	return &UserPage{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *UserService) Update(ctx context.Context, id uint64, input UpdateUserInput) (*model.User, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, errs.NewValidation("name is required", map[string]string{
			"name": "name is required",
		})
	}

	birth, err := parseBirth(input.Birth)
	if err != nil {
		return nil, err
	}

	return s.repo.Update(ctx, id, repository.UserUpdatePatch{
		Name:     name,
		Email:    normalizeOptionalString(input.Email),
		UsedName: strings.TrimSpace(input.UsedName),
		Company:  strings.TrimSpace(input.Company),
		Birth:    birth,
	})
}

func (s *UserService) Delete(ctx context.Context, id uint64) error {
	return s.repo.Delete(ctx, id)
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
		return nil, errs.NewValidation("birth must be in YYYY-MM-DD format", map[string]string{
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
