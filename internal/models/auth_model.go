package models

import (
	"github.com/golang-jwt/jwt/v5"
)

type LoginRequest struct {
	Email    string `json:"email"`
}

type LoginResponse struct {
	Access string `json:"access"`
	Refresh string `json:"refresh"`
}

type RegistrationRequest struct {
	Email    string `json:"email"`
}

type RegistrationResponse struct {
    UserID       string `json:"user_id"`
    Access        string `json:"access"`
    Refresh string `json:"refresh"`
}

type Claims struct {
    IP     string `json:"ip"`
    UserID string `json:"userID"`
    Jti    string `json:"jti"`
    jwt.RegisteredClaims
}