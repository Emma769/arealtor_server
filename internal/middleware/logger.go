package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

type LoggerOptions struct {
	*slog.Logger
}

func LoggerWithOptions(opts *LoggerOptions) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			now := time.Now()

			defer func() {
				attrs := []slog.Attr{
					{Key: "method", Value: slog.StringValue(r.Method)},
					{Key: "path", Value: slog.StringValue(r.RequestURI)},
					{Key: "latency", Value: slog.StringValue(fmt.Sprintf("%v", time.Since(now)))},
				}
				opts.LogAttrs(r.Context(), slog.LevelInfo, "incoming request", attrs...)
			}()

			next.ServeHTTP(w, r)
		})
	}
}
