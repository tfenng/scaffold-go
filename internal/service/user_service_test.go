package service

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"scaffold-api/internal/db/query"
	"scaffold-api/internal/errs"

	"github.com/stretchr/testify/require"
	turso "turso.tech/database/tursogo"
)

type stubUserStore struct {
	createFn func(context.Context, query.CreateUserParams) (query.User, error)
	getFn    func(context.Context, int64) (query.User, error)
	listFn   func(context.Context, query.ListUsersParams) ([]query.User, error)
	countFn  func(context.Context, query.CountUsersParams) (int64, error)
	updateFn func(context.Context, query.UpdateUserParams) (query.User, error)
	deleteFn func(context.Context, int64) (int64, error)
}

func (s stubUserStore) CreateUser(ctx context.Context, arg query.CreateUserParams) (query.User, error) {
	return s.createFn(ctx, arg)
}

func (s stubUserStore) GetUserByID(ctx context.Context, id int64) (query.User, error) {
	return s.getFn(ctx, id)
}

func (s stubUserStore) ListUsers(ctx context.Context, arg query.ListUsersParams) ([]query.User, error) {
	return s.listFn(ctx, arg)
}

func (s stubUserStore) CountUsers(ctx context.Context, arg query.CountUsersParams) (int64, error) {
	return s.countFn(ctx, arg)
}

func (s stubUserStore) UpdateUser(ctx context.Context, arg query.UpdateUserParams) (query.User, error) {
	return s.updateFn(ctx, arg)
}

func (s stubUserStore) DeleteUser(ctx context.Context, id int64) (int64, error) {
	return s.deleteFn(ctx, id)
}

func TestCreateNormalizesOptionalFields(t *testing.T) {
	t.Parallel()

	var captured query.CreateUserParams
	svc := NewUserService(stubUserStore{
		createFn: func(_ context.Context, arg query.CreateUserParams) (query.User, error) {
			captured = arg
			return query.User{}, nil
		},
		getFn: func(context.Context, int64) (query.User, error) {
			return query.User{}, nil
		},
		listFn: func(context.Context, query.ListUsersParams) ([]query.User, error) {
			return nil, nil
		},
		countFn: func(context.Context, query.CountUsersParams) (int64, error) {
			return 0, nil
		},
		updateFn: func(context.Context, query.UpdateUserParams) (query.User, error) {
			return query.User{}, nil
		},
		deleteFn: func(context.Context, int64) (int64, error) {
			return 1, nil
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
	require.Equal(t, "user-001", captured.Uid)
	require.Equal(t, "Alice", captured.Name)
	require.True(t, captured.Email != nil)
	require.Equal(t, "alice@example.com", *captured.Email)
	require.Equal(t, "ali", captured.UsedName)
	require.Equal(t, "ACME", captured.Company)
	require.NotNil(t, captured.Birth)
	require.Equal(t, "2024-10-01", *captured.Birth)
}

func TestCreateRejectsInvalidBirth(t *testing.T) {
	t.Parallel()

	svc := NewUserService(stubUserStore{
		createFn: func(_ context.Context, arg query.CreateUserParams) (query.User, error) {
			return query.User{}, nil
		},
		getFn: func(context.Context, int64) (query.User, error) {
			return query.User{}, nil
		},
		listFn: func(context.Context, query.ListUsersParams) ([]query.User, error) {
			return nil, nil
		},
		countFn: func(context.Context, query.CountUsersParams) (int64, error) {
			return 0, nil
		},
		updateFn: func(context.Context, query.UpdateUserParams) (query.User, error) {
			return query.User{}, nil
		},
		deleteFn: func(context.Context, int64) (int64, error) {
			return 1, nil
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

	var captured query.ListUsersParams
	svc := NewUserService(stubUserStore{
		createFn: func(_ context.Context, arg query.CreateUserParams) (query.User, error) {
			return query.User{}, nil
		},
		getFn: func(context.Context, int64) (query.User, error) {
			return query.User{}, nil
		},
		listFn: func(_ context.Context, arg query.ListUsersParams) ([]query.User, error) {
			captured = arg
			return []query.User{{
				ID:        1,
				Name:      "Alice",
				CreatedAt: "2026-03-29T12:00:00Z",
			}}, nil
		},
		countFn: func(context.Context, query.CountUsersParams) (int64, error) {
			return 1, nil
		},
		updateFn: func(context.Context, query.UpdateUserParams) (query.User, error) {
			return query.User{}, nil
		},
		deleteFn: func(context.Context, int64) (int64, error) {
			return 1, nil
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
	require.Equal(t, int64(maxPageSize), captured.Limit)
	require.Equal(t, int64(0), captured.Offset)
	require.NotNil(t, captured.NameLike)
	require.Equal(t, "Alice", *captured.NameLike)
}

func TestTranslateStoreErrorMapsNoRowsToNotFound(t *testing.T) {
	t.Parallel()

	err := translateStoreError(sql.ErrNoRows)

	require.Error(t, err)
	require.ErrorIs(t, err, errs.ErrNotFound)
}

func TestTranslateStoreErrorMapsUniqueViolationToConflict(t *testing.T) {
	t.Parallel()

	err := translateStoreError(turso.ErrTursoConstraint)

	require.Error(t, err)
	require.ErrorIs(t, err, errs.ErrConflict)
}

func TestTranslateStoreErrorWrapsUnknownAsInternal(t *testing.T) {
	t.Parallel()

	root := errors.New("boom")
	err := translateStoreError(root)

	require.Error(t, err)
	require.ErrorIs(t, err, errs.ErrInternal)
}

func TestPatchPreservesOmittedFieldsAndClearsNullableOnNull(t *testing.T) {
	t.Parallel()

	email := "alice@example.com"
	birth := "2024-10-01"

	var captured query.UpdateUserParams
	svc := NewUserService(stubUserStore{
		createFn: func(_ context.Context, arg query.CreateUserParams) (query.User, error) {
			return query.User{}, nil
		},
		getFn: func(context.Context, int64) (query.User, error) {
			return query.User{
				ID:       7,
				Name:     "Alice",
				Email:    &email,
				UsedName: "Ali",
				Company:  "ACME",
				Birth:    &birth,
			}, nil
		},
		listFn: func(context.Context, query.ListUsersParams) ([]query.User, error) {
			return nil, nil
		},
		countFn: func(context.Context, query.CountUsersParams) (int64, error) {
			return 0, nil
		},
		updateFn: func(_ context.Context, arg query.UpdateUserParams) (query.User, error) {
			captured = arg
			return query.User{ID: arg.ID, Name: arg.Name, UsedName: arg.UsedName, Company: arg.Company}, nil
		},
		deleteFn: func(context.Context, int64) (int64, error) {
			return 1, nil
		},
	})

	company := "  Example Co  "
	result, err := svc.Patch(context.Background(), 7, PatchUserInput{
		Company: OptionalString{Set: true, Value: &company},
		Email:   OptionalString{Set: true, Value: nil},
		Birth:   OptionalString{Set: true, Value: nil},
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, int64(7), captured.ID)
	require.Equal(t, "Alice", captured.Name)
	require.Equal(t, "Ali", captured.UsedName)
	require.Equal(t, "Example Co", captured.Company)
	require.Nil(t, captured.Email)
	require.Nil(t, captured.Birth)
}

func TestPatchRejectsBlankNameWhenProvided(t *testing.T) {
	t.Parallel()

	name := "   "
	svc := NewUserService(stubUserStore{
		createFn: func(_ context.Context, arg query.CreateUserParams) (query.User, error) {
			return query.User{}, nil
		},
		getFn: func(context.Context, int64) (query.User, error) {
			return query.User{Name: "Alice"}, nil
		},
		listFn: func(context.Context, query.ListUsersParams) ([]query.User, error) {
			return nil, nil
		},
		countFn: func(context.Context, query.CountUsersParams) (int64, error) {
			return 0, nil
		},
		updateFn: func(context.Context, query.UpdateUserParams) (query.User, error) {
			return query.User{}, nil
		},
		deleteFn: func(context.Context, int64) (int64, error) {
			return 1, nil
		},
	})

	_, err := svc.Patch(context.Background(), 1, PatchUserInput{
		Name: OptionalString{Set: true, Value: &name},
	})

	require.Error(t, err)
	require.ErrorIs(t, err, errs.ErrInvalidArgument)
}

func TestPatchReturnsCurrentUserWhenPayloadHasNoChanges(t *testing.T) {
	t.Parallel()

	current := query.User{ID: 9, Name: "Alice"}
	updateCalled := false
	svc := NewUserService(stubUserStore{
		createFn: func(_ context.Context, arg query.CreateUserParams) (query.User, error) {
			return query.User{}, nil
		},
		getFn: func(context.Context, int64) (query.User, error) {
			return current, nil
		},
		listFn: func(context.Context, query.ListUsersParams) ([]query.User, error) {
			return nil, nil
		},
		countFn: func(context.Context, query.CountUsersParams) (int64, error) {
			return 0, nil
		},
		updateFn: func(context.Context, query.UpdateUserParams) (query.User, error) {
			updateCalled = true
			return query.User{}, nil
		},
		deleteFn: func(context.Context, int64) (int64, error) {
			return 1, nil
		},
	})

	result, err := svc.Patch(context.Background(), 9, PatchUserInput{})

	require.NoError(t, err)
	require.Equal(t, current, *result)
	require.False(t, updateCalled)
}
