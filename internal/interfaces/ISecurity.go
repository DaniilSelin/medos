package interfaces

import (
	"AuthService/internal/models"
)
	

type ISecurity interface {
	GenerateToken(userID , ip string) (token string, refresh string, err error)
	ValidateToken(access string) (*models.Claims, error)
	RefreshToken(oldToken string) (string, string, error)
	HashRefreshToken(token string) (string, error) 
	ValidateRefreshToken(storedToken *models.RefreshToken, inputRefresh string, accessJTI string) error
	GenerateRefreshToken() string
}