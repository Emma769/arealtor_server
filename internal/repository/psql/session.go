package psql

import (
	"context"
	"database/sql"
	"errors"

	"github.com/emma769/a-realtor/internal/entity"
	"github.com/emma769/a-realtor/internal/repository"
)

func (q *queries) CreateSession(ctx context.Context, session *entity.Session) error {
	stmt := `INSERT INTO sessions (hash, user_id, valid_till) VALUES ($1, $2, $3);`
	_, err := q.db.ExecContext(ctx, stmt, session.Hash, session.UserID, session.ValidTill)
	return err
}

func (q *queries) FindUserBySession(ctx context.Context, hash []byte) (*entity.User, error) {
	stmt := `
  SELECT user_id, name, email, password, created_at FROM users WHERE user_id in (
    SELECT user_id FROM sessions WHERE hash = $1 AND valid_till > current_timestamp
  );
  `
	row := q.db.QueryRowContext(ctx, stmt, hash)
	var user entity.User

	err := ScanUser(row, &user)

	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, repository.ErrNotFound
	}

	return &user, err
}
