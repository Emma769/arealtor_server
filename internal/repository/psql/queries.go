package psql

import (
	"context"
	"database/sql"
)

type scanner interface {
	Scan(...any) error
}

type PaginationParam interface {
	Offset() int
	Limit() int
	SetTotal(int)
}

type dbtx interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}

type queries struct {
	db dbtx
}

func newQueries(db dbtx) *queries {
	return &queries{db}
}

func (q *queries) WithTX(db dbtx) *queries {
	return newQueries(db)
}
