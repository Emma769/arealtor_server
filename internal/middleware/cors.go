package middleware

import (
	"net/http"
	"strings"
)

type CorsOptions struct {
	Origins []string
	Headers []string
	Methods []string
}

func EnableCorsWithOptions(opts *CorsOptions) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Vary", "Origin")
			w.Header().Add("Vary", "Access-Control-Request-Method")

			origin := r.Header.Get("Origin")
			method := r.Header.Get("Access-Control-Request-Method")

			if origin != "" {
				for i := range opts.Origins {
					if opts.Origins[i] == origin {
						w.Header().Set("Access-Control-Allow-Origin", origin)

						if r.Method == "OPTIONS" && method != "" {
							methods := strings.Join(opts.Methods, ", ")
							w.Header().Set("Access-Control-Allow-Methods", methods)
							headers := strings.Join(opts.Headers, ", ")
							w.Header().Set("Access-Control-Allow-Headers", headers)
							w.WriteHeader(200)
							return
						}

						break
					}
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
