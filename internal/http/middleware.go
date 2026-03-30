package httpapi

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

// statusRecorder 状态码记录器.
// 用于在响应写入时记录 HTTP 状态码.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

// WriteHeader 写入响应头并记录状态码.
func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

// RequestLogger 请求日志中间件.
// 记录每个 HTTP 请求的方法、路径、状态码和延迟.
func RequestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			start := time.Now()
			next.ServeHTTP(rec, r)

			logger.Info("http request",
				slog.String("request_id", middleware.GetReqID(r.Context())),
				slog.String("method", r.Method),
				slog.String("uri", r.RequestURI),
				slog.Int("status", rec.status),
				slog.Duration("latency", time.Since(start)),
			)
		})
	}
}

// PreflightNoContent 处理预检请求.
// 对于 OPTIONS 预检请求返回 204 No Content.
func PreflightNoContent(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
