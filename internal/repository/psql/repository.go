package psql

import (
	"cmp"
	"context"
	"database/sql"
	"log/slog"
	"time"

	_ "github.com/lib/pq"
)

const driver = "postgres"

type Repository struct {
	*queries
	db     *sql.DB
	logger *slog.Logger
}

type RepositoryOptions struct {
	MaxIdleConns,
	MaxOpenConns,
	ConnMaxIdleTime int
}

func New(
	ctx context.Context,
	uri string,
	logger *slog.Logger,
	options *RepositoryOptions,
) (*Repository, error) {
	db, err := sql.Open(driver, uri)
	if err != nil {
		return nil, err
	}

	db.SetMaxIdleConns(cmp.Or(options.MaxIdleConns, 25))
	db.SetMaxOpenConns(cmp.Or(options.MaxOpenConns, 25))
	db.SetConnMaxIdleTime(time.Duration(cmp.Or(options.ConnMaxIdleTime, 15)) * time.Second)

	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	return &Repository{
		db:      db,
		queries: newQueries(db),
		logger:  logger,
	}, nil
}

func (r *Repository) Close() error {
	return r.db.Close()
}
