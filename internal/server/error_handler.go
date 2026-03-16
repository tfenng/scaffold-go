package server

import (
	"errors"
	"fmt"
	"net/http"

	"scaffold-api/internal/errs"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

func NewHTTPErrorHandler(logger zerolog.Logger) echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		if c.Response().Committed {
			return
		}

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
			case errors.Is(appErr, errs.ErrValidation):
				status = http.StatusBadRequest
			case errors.Is(appErr, errs.ErrConflict):
				status = http.StatusConflict
			case errors.Is(appErr, errs.ErrNotFound):
				status = http.StatusNotFound
			}

			response.Error = ErrorDetail{
				Code:    appErr.Code,
				Message: appErr.Message,
				Fields:  appErr.Fields,
			}
		} else {
			var httpErr *echo.HTTPError
			if errors.As(err, &httpErr) {
				status = httpErr.Code
				response.Error = ErrorDetail{
					Code:    "http_error",
					Message: fmt.Sprint(httpErr.Message),
				}
			}
		}

		event := logger.Warn()
		if status >= http.StatusInternalServerError {
			event = logger.Error()
		}
		event.Err(err).
			Str("method", c.Request().Method).
			Str("path", c.Request().URL.Path).
			Int("status", status).
			Msg("request failed")

		_ = c.JSON(status, response)
	}
}
