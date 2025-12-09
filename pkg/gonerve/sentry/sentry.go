package sentry

import (
	"context"
	"net/http"

	"github.com/getsentry/sentry-go"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"stockpilot/pkg/gonerve/logging"
)

type Config struct {
	DSN           string  `mapstructure:"dsn" json:"dsn" yaml:"dsn"`
	ErrSampleRate float64 `mapstructure:"err_sample_rate" json:"err_sample_rate" yaml:"err_sample_rate"`
}

func ErrEchoMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		err := next(ctx)
		if err != nil {
			sentry.CaptureException(err)
		}
		return err
	}
}

func PanicEchoMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		defer recoverPanicTo500(ctx.Request().Context(), ctx.Response())
		return next(ctx)
	}
}

func recoverPanicTo500(ctx context.Context, r *echo.Response) {
	if rec := recover(); rec != nil {
		logging.Error(ctx, "panic caught", zap.Any("error", rec))
		r.WriteHeader(http.StatusInternalServerError)
	}
}
