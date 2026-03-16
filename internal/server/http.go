package server

import (
	"context"
	"errors"
	"net/http"

	_ "scaffold-api/docs"
	"scaffold-api/internal/config"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
	echoSwagger "github.com/swaggo/echo-swagger"
	"github.com/swaggo/swag"
	"go.uber.org/fx"
)

func NewEcho(logger zerolog.Logger) *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Validator = NewCustomValidator()
	e.HTTPErrorHandler = NewHTTPErrorHandler(logger)

	e.Use(middleware.RequestID())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:    true,
		LogURI:       true,
		LogMethod:    true,
		LogLatency:   true,
		LogRequestID: true,
		HandleError:  true,
		LogValuesFunc: func(c echo.Context, values middleware.RequestLoggerValues) error {
			event := logger.Info()
			if values.Error != nil {
				event = logger.Error().Err(values.Error)
			}

			event.
				Str("request_id", values.RequestID).
				Str("method", values.Method).
				Str("uri", values.URI).
				Int("status", values.Status).
				Dur("latency", values.Latency).
				Msg("http request")
			return nil
		},
	}))

	return e
}

func RegisterHTTPServer(lc fx.Lifecycle, e *echo.Echo, cfg *config.Config, logger zerolog.Logger) {
	srv := &http.Server{
		Addr:         cfg.Address(),
		Handler:      e,
		ReadTimeout:  cfg.ReadTimeout(),
		WriteTimeout: cfg.WriteTimeout(),
	}

	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			go func() {
				if err := e.StartServer(srv); err != nil && !errors.Is(err, http.ErrServerClosed) {
					logger.Error().Err(err).Msg("http server crashed")
				}
			}()

			logger.Info().
				Str("addr", cfg.Address()).
				Str("env", cfg.Environment).
				Msg("http server started")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			shutdownCtx, cancel := context.WithTimeout(ctx, cfg.ShutdownTimeout())
			defer cancel()

			logger.Info().Msg("shutting down http server")
			return e.Shutdown(shutdownCtx)
		},
	})
}

func RegisterHealthRoutes(e *echo.Echo) {
	e.GET("/healthz", healthzHandler)
}

func RegisterDocsRoutes(e *echo.Echo) {
	e.GET("/swagger/swagger.json", func(c echo.Context) error {
		doc, err := swag.ReadDoc()
		if err != nil {
			return err
		}
		return c.Blob(http.StatusOK, echo.MIMEApplicationJSONCharsetUTF8, []byte(doc))
	})
	e.GET("/swagger/*", echoSwagger.EchoWrapHandler(echoSwagger.URL("/swagger/swagger.json")))
}

// healthzHandler godoc
// @Summary Health check
// @Description Returns the service health status.
// @Tags system
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /healthz [get]
func healthzHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, HealthResponse{Status: "ok"})
}
