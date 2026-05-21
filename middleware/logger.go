package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/labstack/echo/v5"
)

func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, "loggerKey", logger)
}
func FromContext(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value("loggerKey").(*slog.Logger); ok {
		return logger
	}
	return slog.Default()
}

func RequestLogger(logger *slog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			start := time.Now()

			err := next(c)

			status := http.StatusOK
			if resp, unwrapErr := echo.UnwrapResponse(c.Response()); unwrapErr == nil {
				status = resp.Status
			}
			if err != nil {
				if he, ok := err.(*echo.HTTPError); ok {
					status = he.Code
				} else {
					status = 500
				}
			}

			logger.Info("request",
				"method", c.Request().Method,
				"path", c.Request().URL.Path,
				"status", status,
				"latency", time.Since(start).String(),
				"ip", c.RealIP(),
			)

			return err
		}
	}
}
