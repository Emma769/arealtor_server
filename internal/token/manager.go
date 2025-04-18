package token

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/emma769/a-realtor/internal/config"
	"github.com/emma769/a-realtor/internal/entity"
)

var signingMethod = jwt.SigningMethodHS256

type storer interface {
	CreateSession(context.Context, *entity.Session) error
}

type Manager struct {
	store  storer
	config *config.Config
}

func NewMgr(cfg *config.Config, store storer) *Manager {
	return &Manager{store, cfg}
}

type RefreshToken struct {
	Token     string
	ValidTill time.Time
}

type TokenPair struct {
	Access  *jwt.Token
	Refresh *RefreshToken
}

func (mgr *Manager) GetTokenPair(ctx context.Context, id uuid.UUID) (*TokenPair, error) {
	access, err := mgr.GetAccessToken(id)
	if err != nil {
		return nil, err
	}

	refresh, err := mgr.getRefreshToken(ctx, id)
	if err != nil {
		return nil, err
	}

	return &TokenPair{access, refresh}, nil
}

type Payload struct {
	UserID uuid.UUID
	jwt.RegisteredClaims
}

func newPayload(id uuid.UUID, exp time.Duration) *Payload {
	now := time.Now()

	return &Payload{
		UserID: id,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(exp)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}
}

func (mgr *Manager) GetAccessToken(id uuid.UUID) (t *jwt.Token, err error) {
	t = jwt.NewWithClaims(signingMethod, newPayload(id, mgr.config.JwtAccessExpire))
	t.Raw, err = t.SignedString([]byte(mgr.config.JwtAccessSecret))
	return
}

func (mgr *Manager) getRefreshToken(ctx context.Context, id uuid.UUID) (*RefreshToken, error) {
	b := make([]byte, 32)

	if _, err := rand.Read(b); err != nil {
		return nil, err
	}

	raw := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b)

	validTill := time.Now().Add(mgr.config.SessionExpire)
	h := sha256.Sum256([]byte(raw))

	if err := mgr.store.CreateSession(ctx, &entity.Session{
		Hash:      h[:],
		ValidTill: validTill,
		UserID:    id,
	}); err != nil {
		return nil, err
	}

	return &RefreshToken{
		Token:     raw,
		ValidTill: validTill,
	}, nil
}

func (mgr *Manager) DecodeAccessToken(raw string) (uuid.UUID, error) {
	t, err := jwt.ParseWithClaims(raw, &Payload{}, func(t *jwt.Token) (interface{}, error) {
		if signingMethod != t.Method {
			return nil, fmt.Errorf("invalid signing method")
		}
		return []byte(mgr.config.JwtAccessSecret), nil
	})
	if err != nil {
		return uuid.UUID{}, err
	}

	payload, ok := t.Claims.(*Payload)
	if !ok || !t.Valid {
		return uuid.UUID{}, fmt.Errorf("invalid token")
	}

	return payload.UserID, nil
}
