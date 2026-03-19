package httpapi

import (
	"log/slog"
	stdhttp "net/http"

	_ "scaffold-api/docs"
	"scaffold-api/internal/config"
	"scaffold-api/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/swaggo/swag"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

func NewHandler(cfg *config.Config, logger *slog.Logger, userService *service.UserService) stdhttp.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.CORSAllowOrigins,
		AllowedMethods:   []string{"GET", "HEAD", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		AllowCredentials: true,
		OptionsPassthrough: true,
	}))
	r.Use(PreflightNoContent)
	r.Use(RequestLogger(logger))

	registerDocsRoutes(r, logger)
	registerHealthRoutes(r, logger)
	registerUserRoutes(r, logger, NewUserHandler(userService))

	return r
}

func registerHealthRoutes(r chi.Router, logger *slog.Logger) {
	r.Method(stdhttp.MethodGet, "/healthz", wrap(logger, healthzHandler))
}

func registerDocsRoutes(r chi.Router, logger *slog.Logger) {
	r.Method(stdhttp.MethodGet, "/swagger/swagger.json", wrap(logger, func(w stdhttp.ResponseWriter, r *stdhttp.Request) error {
		doc, err := swag.ReadDoc()
		if err != nil {
			return err
		}
		writeSwaggerJSON(w, doc)
		return nil
	}))
	r.Handle("/swagger/*", httpSwagger.Handler(httpSwagger.URL("/swagger/swagger.json")))
}

func registerUserRoutes(r chi.Router, logger *slog.Logger, handler *UserHandler) {
	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/users", func(r chi.Router) {
			r.Method(stdhttp.MethodPost, "/", wrap(logger, handler.Create))
			r.Method(stdhttp.MethodGet, "/", wrap(logger, handler.List))
			r.Method(stdhttp.MethodGet, "/{id}", wrap(logger, handler.GetByID))
			r.Method(stdhttp.MethodPut, "/{id}", wrap(logger, handler.Update))
			r.Method(stdhttp.MethodPatch, "/{id}", wrap(logger, handler.Update))
			r.Method(stdhttp.MethodDelete, "/{id}", wrap(logger, handler.Delete))
		})
	})
}

// healthzHandler godoc
// @Summary Health check
// @Description Returns the service health status.
// @Tags system
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /healthz [get]
func healthzHandler(w stdhttp.ResponseWriter, _ *stdhttp.Request) error {
	writeJSON(w, stdhttp.StatusOK, HealthResponse{Status: "ok"})
	return nil
}
