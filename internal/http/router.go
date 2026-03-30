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
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"github.com/swaggo/swag"
)

// NewHandler 创建 HTTP 路由处理器.
// 配置中间件、路由和文档路由.
func NewHandler(cfg *config.Config, logger *slog.Logger, userService *service.UserService) stdhttp.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:     cfg.CORSAllowOrigins,
		AllowedMethods:     []string{"GET", "HEAD", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		AllowCredentials:   true,
		OptionsPassthrough: true,
	}))
	r.Use(PreflightNoContent)
	r.Use(RequestLogger(logger))

	registerDocsRoutes(r, logger)
	registerHealthRoutes(r, logger)
	registerUserRoutes(r, logger, NewUserHandler(userService))

	return r
}

// registerHealthRoutes 注册健康检查路由.
func registerHealthRoutes(r chi.Router, logger *slog.Logger) {
	r.Method(stdhttp.MethodGet, "/healthz", wrap(logger, healthzHandler))
}

// registerDocsRoutes 注册 API 文档路由.
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

// registerUserRoutes 注册用户 API 路由.
// 配置 /api/v1/users 下的 CRUD 路由.
func registerUserRoutes(r chi.Router, logger *slog.Logger, handler *UserHandler) {
	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/users", func(r chi.Router) {
			r.Method(stdhttp.MethodPost, "/", wrap(logger, handler.Create))
			r.Method(stdhttp.MethodGet, "/", wrap(logger, handler.List))
			r.Method(stdhttp.MethodGet, "/{id}", wrap(logger, handler.GetByID))
			r.Method(stdhttp.MethodPut, "/{id}", wrap(logger, handler.Update))
			r.Method(stdhttp.MethodPatch, "/{id}", wrap(logger, handler.Patch))
			r.Method(stdhttp.MethodDelete, "/{id}", wrap(logger, handler.Delete))
		})
	})
}

// healthzHandler 健康检查处理器.
// 返回服务健康状态.
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
