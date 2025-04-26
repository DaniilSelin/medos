package models

import (
	"time"
)

type RefreshToken struct {
	ID              int   	  `json:"id"`
	UserID          string    `json:"user_id"`
	HashedToken     string    `json:"hashed_token"`
	AccessTokenJTI  string    `json:"access_token_jti"`
	ClientIP        string    `json:"client_ip"`
	ExpiresAt       time.Time `json:"expires_at"`
}