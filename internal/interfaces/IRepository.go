package interfaces

import (
    "AuthService/internal/models"

    "context"
)

type IRepository interface {
    RegisterUser(ctx context.Context, user *models.User) error
    GetUserByEmail(ctx context.Context, email string) (*models.User, error)
    SaveRefreshToken(ctx context.Context, token *models.RefreshToken) error
    GetRefreshToken(ctx context.Context, userID, jti string) (*models.RefreshToken, error)
    RevokeRefreshToken(ctx context.Context, tokenID int) error
    GetUser(ctx context.Context, userID string) (*models.User, error)
}