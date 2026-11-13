package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadAppliesDefaults(t *testing.T) {
	t.Setenv("APP_ENV", "")
	t.Setenv("HTTP_HOST", "")
	t.Setenv("HTTP_PORT", "")
	t.Setenv("DB_PATH", "")
	t.Setenv("LOG_LEVEL", "")
	t.Setenv("CORS_ALLOW_ORIGINS", "")

	cfg, err := Load()

	require.NoError(t, err)
	require.Equal(t, defaultAppEnv, cfg.AppEnv)
	require.Equal(t, defaultHTTPHost, cfg.HTTPHost)
	require.Equal(t, defaultHTTPPort, cfg.HTTPPort)
	require.Equal(t, defaultDBPath, cfg.DBPath)
	require.Equal(t, defaultLogLevel, cfg.LogLevel)
	require.Equal(t, []string{"http://localhost:3000", "http://127.0.0.1:3000"}, cfg.CORSAllowOrigins)
}

func TestLoadRejectsBlankDBPath(t *testing.T) {
	t.Setenv("DB_PATH", "   ")

	_, err := Load()

	// A blank value falls back to the default path, so Load succeeds.
	require.NoError(t, err)
}

func TestLoadRejectsInvalidHTTPPort(t *testing.T) {
	t.Setenv("HTTP_PORT", "abc")

	_, err := Load()

	require.Error(t, err)
	require.Contains(t, err.Error(), "HTTP_PORT must be an integer")
}

func TestCSVFromEnvSkipsEmptyValues(t *testing.T) {
	key := "TEST_CORS_ALLOW_ORIGINS"
	require.NoError(t, os.Setenv(key, " http://a.com, ,http://b.com "))
	defer os.Unsetenv(key)

	values := csvFromEnv(key, nil)

	require.Equal(t, []string{"http://a.com", "http://b.com"}, values)
}
