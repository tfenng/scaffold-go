package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"scaffold-api/internal/config"
	"scaffold-api/internal/db/query"
	"scaffold-api/internal/service"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

type stubUserStore struct {
	getFn    func(context.Context, int64) (query.User, error)
	updateFn func(context.Context, query.UpdateUserParams) (query.User, error)
}

func (stubUserStore) CreateUser(_ context.Context, arg query.CreateUserParams) (query.User, error) {
	return query.User{}, nil
}

func (s stubUserStore) GetUserByID(ctx context.Context, id int64) (query.User, error) {
	if s.getFn != nil {
		return s.getFn(ctx, id)
	}
	return query.User{}, nil
}

func (stubUserStore) ListUsers(context.Context, query.ListUsersParams) ([]query.User, error) {
	return nil, nil
}

func (stubUserStore) CountUsers(context.Context, query.CountUsersParams) (int64, error) {
	return 0, nil
}

func (s stubUserStore) UpdateUser(ctx context.Context, arg query.UpdateUserParams) (query.User, error) {
	if s.updateFn != nil {
		return s.updateFn(ctx, arg)
	}
	return query.User{}, nil
}

func (stubUserStore) DeleteUser(context.Context, int64) (int64, error) {
	return 1, nil
}

func newTestHandler(store service.UserStore) http.Handler {
	cfg := &config.Config{
		ServiceName:      "scaffold-api",
		AppEnv:           "test",
		HTTPHost:         "127.0.0.1",
		HTTPPort:         8080,
		DBDSN:            "postgres://test:test@localhost:5432/test?sslmode=disable",
		LogLevel:         "info",
		CORSAllowOrigins: []string{"http://localhost:3000", "http://127.0.0.1:3000"},
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	userService := service.NewUserService(store)
	return NewHandler(cfg, logger, userService)
}

func TestRegisterDocsRoutesServesSwaggerJSON(t *testing.T) {
	t.Parallel()

	handler := newTestHandler(stubUserStore{})
	req := httptest.NewRequest(http.MethodGet, "/swagger/swagger.json", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var doc struct {
		Paths       map[string]any `json:"paths"`
		Definitions map[string]any `json:"definitions"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &doc))
	require.Contains(t, doc.Paths, "/api/v1/users")
	require.Contains(t, doc.Paths, "/api/v1/users/{id}")
	require.Contains(t, doc.Paths, "/healthz")
	require.Contains(t, doc.Definitions, "httpapi.UserDetailEnvelope")
	require.Contains(t, doc.Definitions, "httpapi.ErrorEnvelope")
}

func TestRegisterDocsRoutesServesSwaggerUI(t *testing.T) {
	t.Parallel()

	handler := newTestHandler(stubUserStore{})
	req := httptest.NewRequest(http.MethodGet, "/swagger/index.html", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), "Swagger UI")
}

func TestHandlerHandlesCORSPreflightForLocalFrontend(t *testing.T) {
	t.Parallel()

	handler := newTestHandler(stubUserStore{})
	req := httptest.NewRequest(http.MethodOptions, "/healthz", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", http.MethodGet)
	req.Header.Set("Access-Control-Request-Headers", "Content-Type, Authorization")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNoContent, rec.Code)
	require.Equal(t, "http://localhost:3000", rec.Header().Get("Access-Control-Allow-Origin"))
	require.Contains(t, rec.Header().Get("Access-Control-Allow-Methods"), http.MethodGet)
	require.Contains(t, rec.Header().Get("Access-Control-Allow-Headers"), "Authorization")
	require.Equal(t, "true", rec.Header().Get("Access-Control-Allow-Credentials"))
}

func TestPatchUserPartiallyUpdatesAndClearsNullableFields(t *testing.T) {
	t.Parallel()

	currentEmail := "alice@example.com"
	currentBirth := time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)
	now := time.Date(2026, 3, 29, 12, 0, 0, 0, time.UTC)

	var captured query.UpdateUserParams
	handler := newTestHandler(stubUserStore{
		getFn: func(context.Context, int64) (query.User, error) {
			return query.User{
				ID:       1,
				Uid:      "user-001",
				Email:    &currentEmail,
				Name:     "Alice",
				UsedName: "Ali",
				Company:  "ACME",
				Birth:    &currentBirth,
				CreatedAt: pgtype.Timestamptz{
					Time:  now,
					Valid: true,
				},
				UpdatedAt: pgtype.Timestamptz{
					Time:  now,
					Valid: true,
				},
			}, nil
		},
		updateFn: func(_ context.Context, arg query.UpdateUserParams) (query.User, error) {
			captured = arg
			return query.User{
				ID:       1,
				Uid:      "user-001",
				Email:    nil,
				Name:     arg.Name,
				UsedName: arg.UsedName,
				Company:  arg.Company,
				Birth:    &currentBirth,
				CreatedAt: pgtype.Timestamptz{
					Time:  now,
					Valid: true,
				},
				UpdatedAt: pgtype.Timestamptz{
					Time:  now,
					Valid: true,
				},
			}, nil
		},
	})

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/users/1", bytes.NewBufferString(`{"company":"Example Co","email":null}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "Alice", captured.Name)
	require.Equal(t, "Ali", captured.UsedName)
	require.Equal(t, "Example Co", captured.Company)
	require.False(t, captured.Email.Valid)
	require.True(t, captured.Birth.Valid)

	var body UserDetailEnvelope
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, int64(1), body.Data.ID)
	require.Equal(t, "Example Co", body.Data.Company)
	require.Nil(t, body.Data.Email)
	require.Equal(t, "Alice", body.Data.Name)
}

func TestPatchUserReturnsValidationErrorForBlankName(t *testing.T) {
	t.Parallel()

	handler := newTestHandler(stubUserStore{
		getFn: func(context.Context, int64) (query.User, error) {
			return query.User{
				ID:   1,
				Uid:  "user-001",
				Name: "Alice",
			}, nil
		},
	})

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/users/1", bytes.NewBufferString(`{"name":"   "}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)

	var body ErrorEnvelope
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, "validation_error", body.Error.Code)
	require.Equal(t, "name is required", body.Error.Message)
	require.Equal(t, "name is required", body.Error.Fields["name"])
}
