package db

import (
	"context"
	"testing"

	"scaffold-api/internal/config"
	"scaffold-api/internal/db/query"

	"github.com/stretchr/testify/require"
	turso "turso.tech/database/tursogo"
)

func newTestDB(t *testing.T) *Store {
	t.Helper()

	ctx := context.Background()
	sqlDB, err := Open(ctx, &config.Config{DBPath: ":memory:"})
	require.NoError(t, err)
	t.Cleanup(func() { sqlDB.Close() })

	require.NoError(t, MigrateUp(ctx, sqlDB))
	// Migrations must be idempotent.
	require.NoError(t, MigrateUp(ctx, sqlDB))

	return NewStore(sqlDB)
}

func TestStoreCRUDRoundTrip(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := newTestDB(t)

	created, err := store.CreateUser(ctx, query.CreateUserParams{
		Uid:      "user-001",
		Name:     "Alice",
		UsedName: "Ali",
		Company:  "ACME",
	})
	require.NoError(t, err)
	require.Equal(t, int64(1), created.ID)
	require.NotEmpty(t, created.CreatedAt)
	require.NotEmpty(t, created.UpdatedAt)

	got, err := store.GetUserByID(ctx, created.ID)
	require.NoError(t, err)
	require.Equal(t, "Alice", got.Name)

	email := "alice@example.com"
	birth := "1990-01-01"
	updated, err := store.UpdateUser(ctx, query.UpdateUserParams{
		ID:       created.ID,
		Name:     "Alice Chen",
		Email:    &email,
		UsedName: "Ali",
		Company:  "ACME",
		Birth:    &birth,
	})
	require.NoError(t, err)
	require.Equal(t, "Alice Chen", updated.Name)
	require.NotNil(t, updated.Email)
	require.Equal(t, email, *updated.Email)
	require.NotNil(t, updated.Birth)
	require.Equal(t, birth, *updated.Birth)

	count, err := store.CountUsers(ctx, query.CountUsersParams{})
	require.NoError(t, err)
	require.Equal(t, int64(1), count)

	list, err := store.ListUsers(ctx, query.ListUsersParams{Offset: 0, Limit: 20})
	require.NoError(t, err)
	require.Len(t, list, 1)
	require.Equal(t, "Alice Chen", list[0].Name)

	filtered, err := store.ListUsers(ctx, query.ListUsersParams{Email: &email, Offset: 0, Limit: 20})
	require.NoError(t, err)
	require.Len(t, filtered, 1)

	rows, err := store.DeleteUser(ctx, created.ID)
	require.NoError(t, err)
	require.Equal(t, int64(1), rows)
}

func TestStoreDuplicateUIDMapsToConstraint(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := newTestDB(t)

	_, err := store.CreateUser(ctx, query.CreateUserParams{Uid: "user-001", Name: "Alice"})
	require.NoError(t, err)

	_, err = store.CreateUser(ctx, query.CreateUserParams{Uid: "user-001", Name: "Bob"})
	require.Error(t, err)
	require.ErrorIs(t, err, turso.ErrTursoConstraint)
}

func TestMigrateDownRevertsLatest(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	sqlDB, err := Open(ctx, &config.Config{DBPath: ":memory:"})
	require.NoError(t, err)
	t.Cleanup(func() { sqlDB.Close() })

	require.NoError(t, MigrateUp(ctx, sqlDB))
	require.NoError(t, MigrateDown(ctx, sqlDB))

	// The users table should be gone, so querying it fails.
	_, err = query.New(sqlDB).CountUsers(ctx, query.CountUsersParams{})
	require.Error(t, err)
}
