package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"

	handlerlib "github.com/emma769/a-realtor/internal/lib/handler"
)

type RecoverOptions struct {
	*slog.Logger
}

func RecoverWithOptions(opts *RecoverOptions) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					attrs := []slog.Attr{
						{Key: "detail", Value: slog.StringValue(fmt.Sprintf("%v", err))},
						{Key: "trace", Value: slog.StringValue(string(debug.Stack()))},
					}
					opts.LogAttrs(r.Context(), slog.LevelError, "server error", attrs...)

					if err := handlerlib.WriteJson(w, 500, handlerlib.ErrResp{
						Detail: "internal server error",
					}); err != nil {
						panic(err)
					}
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
