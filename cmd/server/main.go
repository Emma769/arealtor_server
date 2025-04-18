package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/emma769/a-realtor/internal/config"
	"github.com/emma769/a-realtor/internal/ctrl/landlord"
	"github.com/emma769/a-realtor/internal/ctrl/tenant"
	"github.com/emma769/a-realtor/internal/ctrl/user"
	"github.com/emma769/a-realtor/internal/middleware"
	"github.com/emma769/a-realtor/internal/repository/psql"
	"github.com/emma769/a-realtor/internal/token"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{}))

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	store, err := psql.New(ctx, cfg.PostgresUri, logger, &psql.RepositoryOptions{})
	if err != nil {
		return err
	}

	mgr := token.NewMgr(cfg, store)

	router := chi.NewRouter()

	router.Use(middleware.RecoverWithOptions(&middleware.RecoverOptions{
		Logger: logger,
	}))

	router.Use(middleware.LoggerWithOptions(&middleware.LoggerOptions{
		Logger: logger,
	}))

	router.Use(middleware.EnableCorsWithOptions(&middleware.CorsOptions{
		Origins:     []string{cfg.TrustedOrigin},
		Headers:     []string{"Content-Type", "Accept", "Authorization"},
		Methods:     []string{"GET", "POST", "OPTIONS", "PUT", "DELETE"},
		Credentials: true,
	}))

	router.Use(middleware.Authenticate(middleware.NewAuthService(mgr, store)))

	user := user.NewCtrl(store, cfg, mgr)
	router.Route("/api/auth", user.Routes)

	landlord := landlord.New(store, logger)
	router.Route("/api/landlords", landlord.Routes)

	tenant := tenant.New(store, logger)
	router.Route("/api/tenants", tenant.Routes)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		IdleTimeout:  cfg.IdleTimeout,
		WriteTimeout: cfg.WriteTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		Handler:      router,
	}

	errch := make(chan error)

	go func() {
		logger.LogAttrs(ctx, slog.LevelInfo, "server is starting", slog.Attr{
			Key:   "port",
			Value: slog.IntValue(cfg.Port),
		})

		if err := server.ListenAndServe(); err != nil {
			errch <- err
		}
	}()

	select {
	case err := <-errch:
		return err
	case <-ctx.Done():
		if err := store.Close(); err != nil {
			logger.LogAttrs(ctx, slog.LevelError, "could not close db", slog.Attr{
				Key:   "detail",
				Value: slog.StringValue(err.Error()),
			})
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		logger.LogAttrs(ctx, slog.LevelInfo, "server is shutting down")
		return server.Shutdown(ctx)
	}
}
