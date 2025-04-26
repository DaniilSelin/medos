package interfaces

import (
	"AuthService/internal/models"

	"context"
)

type IAuthService interface {
	Login(ctx context.Context, userID string, ip string) (access, refresh string, err error)
	RefreshToken(ctx context.Context, oldAccess, refreshToken, ip string) (newAccess, newRefresh string, err error)
	Register(ctx context.Context, user *models.User, ip string) (string, string, string, error)
}
