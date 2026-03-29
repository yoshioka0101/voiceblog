package entity

import "time"

type ExternalToken struct {
	ID             int64
	UserID         int64
	Provider       string
	EncryptedToken []byte
	Nonce          []byte
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
