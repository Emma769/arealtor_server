package entity

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	SessionID int64     `json:"sessionID"`
	Hash      []byte    `json:"hash"`
	ValidTill time.Time `json:"validTill"`
	UserID    uuid.UUID `json:"userID"`
	CreatedAt time.Time `json:"createdAt"`
}
