package user

import (
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/emma769/a-realtor/internal/config"
	"github.com/emma769/a-realtor/internal/entity"
	handlerlib "github.com/emma769/a-realtor/internal/lib/handler"
	"github.com/emma769/a-realtor/internal/middleware"
	"github.com/emma769/a-realtor/internal/token"
	"github.com/emma769/a-realtor/internal/validator"
)

const timeout = 5 * time.Second

type Ctrl struct {
	*Service
	cfg *config.Config
	mgr *token.Manager
}

func NewCtrl(store storer, cfg *config.Config, mgr *token.Manager) *Ctrl {
	return &Ctrl{
		mgr: mgr,
		cfg: cfg,
		Service: &Service{
			store,
			timeout,
		},
	}
}

func (ctrl Ctrl) Routes(r chi.Router) {
	r.Post("/register", ctrl.register())
	r.Post("/login", ctrl.login())
	r.Delete("/logout", ctrl.logout())
	r.Get("/refresh", ctrl.refresh())
	r.With(middleware.RequireAuth).Get("/me", ctrl.getMe())
}

func (ctrl *Ctrl) register() http.HandlerFunc {
	return handlerlib.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		in, err := handlerlib.Bind[entity.UserIn](w, r)
		if err != nil {
			return handlerlib.NewError(422, err.Error())
		}

		v := validator.New()

		if entity.ValidateUserIn(v, in); !v.Valid() {
			return handlerlib.WriteJson(w, 422, v.Err())
		}

		user, err := ctrl.create(r.Context(), in)

		if err != nil && errors.Is(err, ErrDuplicateEmail) {
			return handlerlib.NewError(409, "email already in use")
		}

		if err != nil {
			return err
		}

		return handlerlib.WriteJson(w, 201, user)
	})
}

func (ctrl *Ctrl) login() http.HandlerFunc {
	return handlerlib.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		in, err := handlerlib.Bind[entity.LoginIn](w, r)
		if err != nil {
			return handlerlib.NewError(422, err.Error())
		}

		v := validator.New()

		if entity.ValidateLoginIn(v, in); !v.Valid() {
			return handlerlib.WriteJson(w, 422, v.Err())
		}

		user, err := ctrl.findByEmail(r.Context(), in.Email)

		if err != nil && errors.Is(err, ErrNotFound) {
			return handlerlib.NewError(401, "unauthorized")
		}

		if err != nil {
			return err
		}

		err = bcrypt.CompareHashAndPassword(user.Password, []byte(in.Password))

		if err != nil && errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return handlerlib.NewError(401, "unauthorized")
		}

		if err != nil {
			return err
		}

		pair, err := ctrl.mgr.GetTokenPair(r.Context(), user.UserID)
		if err != nil {
			return err
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "refresh_token",
			Value:    pair.Refresh.Token,
			SameSite: http.SameSiteLaxMode,
			Path:     "/",
			HttpOnly: true,
			Secure:   ctrl.cfg.GoEnv == "production",
			MaxAge:   int(ctrl.cfg.SessionExpire.Seconds()),
		})

		return handlerlib.WriteJson(w, 201, map[string]any{
			"token": pair.Access.Raw,
			"type":  "Bearer",
		})
	})
}

func (ctrl *Ctrl) refresh() http.HandlerFunc {
	return handlerlib.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		c, err := r.Cookie("refresh_token")
		if err != nil {
			return handlerlib.NewError(403, "no cookie")
		}

		user, err := ctrl.findBySession(r.Context(), c.Value)

		if err != nil && errors.Is(err, ErrNotFound) {
			return handlerlib.NewError(403, "not logged in, login for access")
		}

		if err != nil {
			return err
		}

		t, err := ctrl.mgr.GetAccessToken(user.UserID)
		if err != nil {
			return err
		}

		return handlerlib.WriteJson(w, 200, map[string]any{
			"token": t.Raw,
			"type":  "Bearer",
		})
	})
}

func (ctrl *Ctrl) logout() http.HandlerFunc {
	return handlerlib.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		http.SetCookie(w, &http.Cookie{
			Name:     "refresh_token",
			Value:    "",
			SameSite: http.SameSiteLaxMode,
			HttpOnly: true,
			MaxAge:   -1,
			Secure:   ctrl.cfg.GoEnv == "production",
			Path:     "/",
		})

		w.WriteHeader(204)
		return nil
	})
}

func (ctrl *Ctrl) getMe() http.HandlerFunc {
	return handlerlib.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		return handlerlib.WriteJson(w, 200, handlerlib.GetCtxUser(r))
	})
}
