package middleware

import (
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"stockpilot/pkg/gonerve/logging"
)

func RequestLogger(l logging.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			err := next(c)
			status := c.Response().Status
			latency := time.Since(start)
			req := c.Request()
			fields := []zap.Field{
				zap.Int("status", status),
				zap.String("method", req.Method),
				zap.String("path", req.URL.Path),
				zap.String("remote_addr", c.RealIP()),
				zap.Duration("latency", latency),
			}
			if err != nil {
				fields = append(fields, zap.Error(err))
				l.ErrorCtx(req.Context(), "http request", fields...)
				return err
			}
			l.InfoCtx(req.Context(), "http request", fields...)
			return nil
		}
	}
}
