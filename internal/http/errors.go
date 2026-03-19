package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"scaffold-api/internal/errs"
)

type appHandler func(http.ResponseWriter, *http.Request) error

func wrap(logger *slog.Logger, handler appHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := handler(w, r); err != nil {
			writeError(logger, w, r, err)
		}
	}
}

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

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeSwaggerJSON(w http.ResponseWriter, doc string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_, _ = fmt.Fprint(w, doc)
}
