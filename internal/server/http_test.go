package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

func TestRegisterDocsRoutesServesSwaggerJSON(t *testing.T) {
	t.Parallel()

	e := echo.New()
	RegisterDocsRoutes(e)

	req := httptest.NewRequest(http.MethodGet, "/swagger/swagger.json", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var doc struct {
		Paths map[string]any `json:"paths"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &doc))
	require.Contains(t, doc.Paths, "/api/v1/users")
	require.Contains(t, doc.Paths, "/api/v1/users/{id}")
	require.Contains(t, doc.Paths, "/healthz")
}

func TestRegisterDocsRoutesServesSwaggerUI(t *testing.T) {
	t.Parallel()

	e := echo.New()
	RegisterDocsRoutes(e)

	req := httptest.NewRequest(http.MethodGet, "/swagger/index.html", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), "Swagger UI")
}
