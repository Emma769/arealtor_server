package user

import (
	"context"
	"crypto/sha256"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/emma769/a-realtor/internal/entity"
	"github.com/emma769/a-realtor/internal/repository"
	"github.com/emma769/a-realtor/internal/repository/psql"
)

var (
	ErrNotFound       = errors.New("user not found")
	ErrDuplicateEmail = errors.New("duplicate email")
)

type storer interface {
	CreateUser(context.Context, psql.UserParam) (*entity.User, error)
	FindUserByEmail(context.Context, string) (*entity.User, error)
	FindUserBySession(context.Context, []byte) (*entity.User, error)
}

type Service struct {
	store   storer
	timeout time.Duration
}

type UserParam struct {
	name,
	email string
	password []byte
}

func (param UserParam) Name() string {
	return param.name
}

func (param UserParam) Email() string {
	return param.email
}

func (param UserParam) Password() []byte {
	return param.password
}

func (s *Service) create(ctx context.Context, in entity.UserIn) (*entity.User, error) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	password, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	param := UserParam{
		name:     in.Name,
		email:    in.Email,
		password: password,
	}

	user, err := s.store.CreateUser(ctx, param)

	if err != nil && errors.Is(err, repository.ErrDuplicateKey) {
		return nil, ErrDuplicateEmail
	}

	return user, err
}

func (s *Service) findByEmail(ctx context.Context, email string) (*entity.User, error) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	user, err := s.store.FindUserByEmail(ctx, email)

	if err != nil && errors.Is(err, repository.ErrNotFound) {
		return nil, ErrNotFound
	}

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *Service) findBySession(ctx context.Context, plain string) (*entity.User, error) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	h := sha256.Sum256([]byte(plain))

	user, err := s.store.FindUserBySession(ctx, h[:])

	if err != nil && errors.Is(err, repository.ErrNotFound) {
		return nil, ErrNotFound
	}

	if err != nil {
		return nil, err
	}

	return user, nil
}
