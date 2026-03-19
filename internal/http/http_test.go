package httpapi

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"scaffold-api/internal/config"
	"scaffold-api/internal/db/query"
	"scaffold-api/internal/service"

	"github.com/stretchr/testify/require"
)

type stubUserStore struct{}

func (stubUserStore) CreateUser(_ context.Context, arg query.CreateUserParams) (query.User, error) {
	return query.User{}, nil
}

func (stubUserStore) GetUserByID(context.Context, int64) (query.User, error) {
	return query.User{}, nil
}

func (stubUserStore) ListUsers(context.Context, query.ListUsersParams) ([]query.User, error) {
	return nil, nil
}

func (stubUserStore) CountUsers(context.Context, query.CountUsersParams) (int64, error) {
	return 0, nil
}

func (stubUserStore) UpdateUser(context.Context, query.UpdateUserParams) (query.User, error) {
	return query.User{}, nil
}

func (stubUserStore) DeleteUser(context.Context, int64) (int64, error) {
	return 1, nil
}

func newTestHandler() http.Handler {
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
	userService := service.NewUserService(stubUserStore{})
	return NewHandler(cfg, logger, userService)
}

func TestRegisterDocsRoutesServesSwaggerJSON(t *testing.T) {
	t.Parallel()

	handler := newTestHandler()
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

	handler := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/swagger/index.html", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), "Swagger UI")
}

func TestHandlerHandlesCORSPreflightForLocalFrontend(t *testing.T) {
	t.Parallel()

	handler := newTestHandler()
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
