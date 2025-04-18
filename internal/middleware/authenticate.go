package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/emma769/a-realtor/internal/entity"
	handlerlib "github.com/emma769/a-realtor/internal/lib/handler"
	"github.com/emma769/a-realtor/internal/repository"
)

type manager interface {
	DecodeAccessToken(string) (uuid.UUID, error)
}

type storer interface {
	FindUserByID(context.Context, uuid.UUID) (*entity.User, error)
}

type AuthService struct {
	mgr   manager
	store storer
}

func NewAuthService(mgr manager, store storer) *AuthService {
	return &AuthService{mgr, store}
}

func Authenticate(svc *AuthService) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Vary", "Authorization")

			auth := r.Header.Get("Authorization")

			if strings.TrimSpace(auth) == "" {
				next.ServeHTTP(w, handlerlib.SetCtxUser(r, entity.AnonymousUser))
				return
			}

			parts := strings.Fields(auth)

			if len(parts) != 2 {
				handlerlib.WriteJson(w, 401, handlerlib.ErrResp{
					Detail: "unauthorized",
				})

				return
			}

			if parts[0] != "Bearer" {
				w.Header().Set("WWW-Authenticate", "Bearer")
				handlerlib.WriteJson(w, 401, handlerlib.ErrResp{
					Detail: "unauthorized",
				})

				return
			}

			id, err := svc.mgr.DecodeAccessToken(parts[1])
			if err != nil {
				w.Header().Set("WWW-Authenticate", "Bearer")
				handlerlib.WriteJson(w, 401, handlerlib.ErrResp{
					Detail: "unauthorized",
				})

				return
			}

			user, err := svc.store.FindUserByID(r.Context(), id)

			if err != nil && errors.Is(err, repository.ErrNotFound) {
				handlerlib.WriteJson(w, 401, handlerlib.ErrResp{
					Detail: "unauthorized",
				})

				return
			}

			if err != nil {
				panic(err)
			}

			next.ServeHTTP(w, handlerlib.SetCtxUser(r, user))
		})
	}
}
