package db

import (
	"context"
	"database/sql"
	"fmt"

	"scaffold-api/internal/db/query"
)

type Store struct {
	db      *sql.DB
	queries *query.Queries
}

func NewStore(db *sql.DB) *Store {
	return &Store{
		db:      db,
		queries: query.New(db),
	}
}

func (s *Store) CreateUser(ctx context.Context, arg query.CreateUserParams) (query.User, error) {
	return s.queries.CreateUser(ctx, arg)
}

func (s *Store) GetUserByID(ctx context.Context, id int64) (query.User, error) {
	return s.queries.GetUserByID(ctx, id)
}

func (s *Store) ListUsers(ctx context.Context, arg query.ListUsersParams) ([]query.User, error) {
	return s.queries.ListUsers(ctx, arg)
}

func (s *Store) CountUsers(ctx context.Context, arg query.CountUsersParams) (int64, error) {
	return s.queries.CountUsers(ctx, arg)
}

func (s *Store) UpdateUser(ctx context.Context, arg query.UpdateUserParams) (query.User, error) {
	return s.queries.UpdateUser(ctx, arg)
}

func (s *Store) DeleteUser(ctx context.Context, id int64) (int64, error) {
	return s.queries.DeleteUser(ctx, id)
}

func (s *Store) WithTx(ctx context.Context, fn func(*query.Queries) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	if err := fn(s.queries.WithTx(tx)); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
