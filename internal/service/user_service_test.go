package service

import (
	"context"
	"testing"
	"time"

	"scaffold-api/internal/model"
	"scaffold-api/internal/repository"

	"github.com/stretchr/testify/require"
)

type stubUserRepository struct {
	createFn func(context.Context, *model.User) (*model.User, error)
	listFn   func(context.Context, repository.UserListFilter) ([]model.User, int64, error)
	updateFn func(context.Context, uint64, repository.UserUpdatePatch) (*model.User, error)
}

func (s stubUserRepository) Create(ctx context.Context, user *model.User) (*model.User, error) {
	return s.createFn(ctx, user)
}

func (s stubUserRepository) GetByID(context.Context, uint64) (*model.User, error) {
	return nil, nil
}

func (s stubUserRepository) List(ctx context.Context, filter repository.UserListFilter) ([]model.User, int64, error) {
	return s.listFn(ctx, filter)
}

func (s stubUserRepository) Update(ctx context.Context, id uint64, patch repository.UserUpdatePatch) (*model.User, error) {
	return s.updateFn(ctx, id, patch)
}

func (s stubUserRepository) Delete(context.Context, uint64) error {
	return nil
}

func TestCreateNormalizesOptionalFields(t *testing.T) {
	t.Parallel()

	var captured *model.User
	svc := NewUserService(stubUserRepository{
		createFn: func(_ context.Context, user *model.User) (*model.User, error) {
			captured = user
			return user, nil
		},
		listFn: func(context.Context, repository.UserListFilter) ([]model.User, int64, error) {
			return nil, 0, nil
		},
		updateFn: func(context.Context, uint64, repository.UserUpdatePatch) (*model.User, error) {
			return nil, nil
		},
	})

	email := "  alice@example.com "
	birth := "2024-10-01"
	_, err := svc.Create(context.Background(), CreateUserInput{
		UID:      " user-001 ",
		Name:     " Alice ",
		Email:    &email,
		UsedName: " ali ",
		Company:  " ACME ",
		Birth:    &birth,
	})

	require.NoError(t, err)
	require.NotNil(t, captured)
	require.Equal(t, "user-001", captured.UID)
	require.Equal(t, "Alice", captured.Name)
	require.NotNil(t, captured.Email)
	require.Equal(t, "alice@example.com", *captured.Email)
	require.Equal(t, "ali", captured.UsedName)
	require.Equal(t, "ACME", captured.Company)
	require.NotNil(t, captured.Birth)
	require.Equal(t, "2024-10-01", captured.Birth.Format("2006-01-02"))
}

func TestCreateRejectsInvalidBirth(t *testing.T) {
	t.Parallel()

	svc := NewUserService(stubUserRepository{
		createFn: func(_ context.Context, user *model.User) (*model.User, error) {
			return user, nil
		},
		listFn: func(context.Context, repository.UserListFilter) ([]model.User, int64, error) {
			return nil, 0, nil
		},
		updateFn: func(context.Context, uint64, repository.UserUpdatePatch) (*model.User, error) {
			return nil, nil
		},
	})

	birth := "2024/10/01"
	_, err := svc.Create(context.Background(), CreateUserInput{
		UID:   "user-001",
		Name:  "Alice",
		Birth: &birth,
	})

	require.Error(t, err)
}

func TestListAppliesPageDefaultsAndCapsSize(t *testing.T) {
	t.Parallel()

	var captured repository.UserListFilter
	svc := NewUserService(stubUserRepository{
		createFn: func(_ context.Context, user *model.User) (*model.User, error) {
			return user, nil
		},
		listFn: func(_ context.Context, filter repository.UserListFilter) ([]model.User, int64, error) {
			captured = filter
			return []model.User{{ID: 1, Name: "Alice", CreatedAt: time.Now()}}, 1, nil
		},
		updateFn: func(context.Context, uint64, repository.UserUpdatePatch) (*model.User, error) {
			return nil, nil
		},
	})

	nameLike := " Alice "
	page, err := svc.List(context.Background(), ListUsersInput{
		NameLike: &nameLike,
		Page:     0,
		PageSize: 999,
	})

	require.NoError(t, err)
	require.Equal(t, 1, page.Page)
	require.Equal(t, maxPageSize, page.PageSize)
	require.Equal(t, maxPageSize, captured.Limit)
	require.Equal(t, 0, captured.Offset)
	require.NotNil(t, captured.NameLike)
	require.Equal(t, "Alice", *captured.NameLike)
}
