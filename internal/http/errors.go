package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"scaffold-api/internal/errs"
)

// appHandler 应用处理器类型.
// 返回错误以供中间件处理.
type appHandler func(http.ResponseWriter, *http.Request) error

// wrap 包装应用处理器为 HTTP 处理器.
// 捕获并处理错误.
func wrap(logger *slog.Logger, handler appHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := handler(w, r); err != nil {
			writeError(logger, w, r, err)
		}
	}
}

// writeError 写入错误响应.
// 根据错误类型确定 HTTP 状态码并记录日志.
func writeError(logger *slog.Logger, w http.ResponseWriter, r *http.Request, err error) {
	status := http.StatusInternalServerError
	response := ErrorEnvelope{
		Error: ErrorDetail{
			Code:    "internal_error",
			Message: "internal server error",
		},
	}

	var appErr *errs.AppError
	if errors.As(err, &appErr) {
		switch {
		case errors.Is(appErr, errs.ErrInvalidArgument):
			status = http.StatusBadRequest
		case errors.Is(appErr, errs.ErrConflict):
			status = http.StatusConflict
		case errors.Is(appErr, errs.ErrNotFound):
			status = http.StatusNotFound
		case errors.Is(appErr, errs.ErrInternal):
			status = http.StatusInternalServerError
		}

		response.Error = ErrorDetail{
			Code:    appErr.Code,
			Message: appErr.Message,
			Fields:  appErr.Fields,
		}
	} else {
		var httpErr interface{ StatusCode() int }
		if errors.As(err, &httpErr) {
			status = httpErr.StatusCode()
			response.Error = ErrorDetail{
				Code:    "http_error",
				Message: err.Error(),
			}
		}
	}

	level := slog.LevelWarn
	if status >= http.StatusInternalServerError {
		level = slog.LevelError
	}
	logger.LogAttrs(r.Context(), level, "request failed",
		slog.Any("error", err),
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
		slog.Int("status", status),
	)

	writeJSON(w, status, response)
}

// writeJSON 写入 JSON 响应.
func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

// writeSwaggerJSON 写入 Swagger JSON 文档.
func writeSwaggerJSON(w http.ResponseWriter, doc string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_, _ = fmt.Fprint(w, doc)
}
