package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"scaffold-api/internal/errs"
	"scaffold-api/internal/model"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

type UserListFilter struct {
	Email    *string
	NameLike *string
	Limit    int
	Offset   int
}

type UserUpdatePatch struct {
	Name     string
	Email    *string
	UsedName string
	Company  string
	Birth    *time.Time
}

type UserRepository interface {
	Create(ctx context.Context, user *model.User) (*model.User, error)
	GetByID(ctx context.Context, id uint64) (*model.User, error)
	List(ctx context.Context, filter UserListFilter) ([]model.User, int64, error)
	Update(ctx context.Context, id uint64, patch UserUpdatePatch) (*model.User, error)
	Delete(ctx context.Context, id uint64) error
}

type GormUserRepository struct {
	db *gorm.DB
}

func NewGormUserRepository(db *gorm.DB) UserRepository {
	return &GormUserRepository{db: db}
}

func (r *GormUserRepository) Create(ctx context.Context, user *model.User) (*model.User, error) {
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		return nil, translateError(err)
	}
	return user, nil
}

func (r *GormUserRepository) GetByID(ctx context.Context, id uint64) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).First(&user, id).Error; err != nil {
		return nil, translateError(err)
	}
	return &user, nil
}

func (r *GormUserRepository) List(ctx context.Context, filter UserListFilter) ([]model.User, int64, error) {
	var total int64
	query := applyUserFilters(r.db.WithContext(ctx).Model(&model.User{}), filter)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, translateError(err)
	}

	var users []model.User
	if err := query.
		Order("created_at DESC, id DESC").
		Limit(filter.Limit).
		Offset(filter.Offset).
		Find(&users).Error; err != nil {
		return nil, 0, translateError(err)
	}

	return users, total, nil
}

func (r *GormUserRepository) Update(ctx context.Context, id uint64, patch UserUpdatePatch) (*model.User, error) {
	updates := map[string]any{
		"name":       patch.Name,
		"email":      patch.Email,
		"used_name":  patch.UsedName,
		"company":    patch.Company,
		"birth":      patch.Birth,
		"updated_at": time.Now(),
	}

	result := r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return nil, translateError(result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, errs.NewNotFound("user not found")
	}

	return r.GetByID(ctx, id)
}

func (r *GormUserRepository) Delete(ctx context.Context, id uint64) error {
	result := r.db.WithContext(ctx).Delete(&model.User{}, id)
	if result.Error != nil {
		return translateError(result.Error)
	}
	if result.RowsAffected == 0 {
		return errs.NewNotFound("user not found")
	}
	return nil
}

func applyUserFilters(db *gorm.DB, filter UserListFilter) *gorm.DB {
	if filter.Email != nil {
		db = db.Where("email = ?", strings.TrimSpace(*filter.Email))
	}
	if filter.NameLike != nil {
		db = db.Where("name ILIKE ?", "%"+strings.TrimSpace(*filter.NameLike)+"%")
	}
	return db
}

func translateError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errs.NewNotFound("user not found")
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == "23505" {
			return errs.NewConflict("user uid or email already exists")
		}
	}

	return fmt.Errorf("repository failure: %w", err)
}
