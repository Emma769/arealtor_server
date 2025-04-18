package psql

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/google/uuid"

	"github.com/emma769/a-realtor/internal/entity"
	"github.com/emma769/a-realtor/internal/repository"
)

type UserParam interface {
	Name() string
	Email() string
	Password() []byte
}

func (q *queries) CreateUser(ctx context.Context, param UserParam) (*entity.User, error) {
	const query = `
  INSERT INTO users (name, email, password) VALUES ($1, $2, $3)
  RETURNING user_id, name, email, password, created_at;
  `
	row := q.db.QueryRowContext(ctx, query, param.Name(), param.Email(), param.Password())

	var user entity.User

	err := ScanUser(row, &user)

	if err != nil && strings.Contains(err.Error(), "duplicate") {
		return nil, repository.ErrDuplicateKey
	}

	return &user, err
}

func (q *queries) FindUserByEmail(ctx context.Context, email string) (*entity.User, error) {
	const query = `
  SELECT user_id, name, email, password, created_at FROM users WHERE email = $1;
  `
	row := q.db.QueryRowContext(ctx, query, email)

	var user entity.User

	err := ScanUser(row, &user)

	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, repository.ErrNotFound
	}

	return &user, err
}

func (q *queries) FindUserByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	const query = `
  SELECT user_id, name, email, password, created_at FROM users WHERE user_id = $1;
  `
	row := q.db.QueryRowContext(ctx, query, id)

	var user entity.User

	err := ScanUser(row, &user)

	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, repository.ErrNotFound
	}

	return &user, err
}

func ScanUser(row scanner, user *entity.User) error {
	return row.Scan(
		&user.UserID,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
	)
}
