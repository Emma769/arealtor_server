package middleware

import (
	"net/http"

	handlerlib "github.com/emma769/a-realtor/internal/lib/handler"
)

func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := handlerlib.GetCtxUser(r)

		if user.IsAnonymous() {
			handlerlib.WriteJson(w, 401, map[string]string{
				"error": "unauthorized",
			})

			return
		}

		next.ServeHTTP(w, r)
	})
}
