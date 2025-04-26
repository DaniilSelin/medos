package service

import (
	"AuthService/internal/models"
	"AuthService/internal/errdefs"
	"AuthService/config"
	"AuthService/internal/interfaces"

	"context"
	"strings"
	"time"
	"fmt"

	"github.com/google/uuid"
)

// Использованные ошибки -
// ErrDB - ошибка на уровне БД
// ErrByScript - Ошибка на уровне библиотеки bcrypt
// ErrNotFound - запись не найдена
// ErrGetHashPswd - ошибка ъеширования пароля
// ErrGenerateToken - ошибка генерации токена
// ErrInvalidToken - токен не раскодировался
// ErrExpiredToken - токен просрочен
// ErrInvalidCredentials - неправильные данные
// ErrEmailAlreadyExists - email уже существует

type AuthService struct {
	repo     interfaces.IRepository
	security interfaces.ISecurity
	cfg *config.Config
}

func NewAuthService(repo interfaces.IRepository, security interfaces.ISecurity, cfg *config.Config) *AuthService {
	return &AuthService{
		repo:     repo,
		security: security,
		cfg: cfg,
	}
}

func (s *AuthService) Register(ctx context.Context, user *models.User, ip string) (string, string, string, error) {
	if !strings.Contains(user.Email, "@") {
		return "", "", "", errdefs.Wrapf(errdefs.ErrInvalidCredentials, "invalid email format")
	}

	userID := uuid.New().String()

	user.ID = userID

	err := s.repo.RegisterUser(ctx, user)
	if err != nil {
		return "", "", "", err
	}

	accessToken, jti, err := s.security.GenerateToken(userID, ip)
	if err != nil {
		return "", "", "", errdefs.Wrapf(errdefs.ErrGenerateToken, "error generating token: %w", err)
	}

	refreshToken := s.security.GenerateRefreshToken()
	hashedRefresh, err := s.security.HashRefreshToken(refreshToken)
	if err != nil {
		return "", "", "", err
	}

	err = s.repo.SaveRefreshToken(ctx, &models.RefreshToken{
		UserID:          userID,
		HashedToken:     hashedRefresh,
		AccessTokenJTI:  jti,
		ClientIP:        ip,
		ExpiresAt:       time.Now().Add(
        	time.Duration(s.cfg.Jwt.RefreshExp) * time.Minute),
	})

	if err != nil {
		return "", "", "", errdefs.Wrapf(errdefs.ErrDB, "failed to save refresh token: %v", err)
	}

	return userID, accessToken, refreshToken, nil
}

func (s *AuthService) Login(ctx context.Context, userID, ip string) (string, string, error) {
	_, err := s.repo.GetUser(ctx, userID)
	if err != nil {
		return "", "", err
	}

	access, jti, err := s.security.GenerateToken(userID, ip)
	if err != nil {
		return "", "", errdefs.ErrGenerateToken
	}

	refresh := s.security.GenerateRefreshToken()
	hashedRefresh, err := s.security.HashRefreshToken(refresh)

	err = s.repo.SaveRefreshToken(ctx, &models.RefreshToken{
		UserID:          userID,
		HashedToken:     hashedRefresh,
		AccessTokenJTI:  jti,
		ClientIP:        ip,
		ExpiresAt:       time.Now().Add(
        	time.Duration(s.cfg.Jwt.RefreshExp) * time.Minute),
	})

	if err != nil {
		return "", "", errdefs.Wrapf(errdefs.ErrDB, "failed to save refresh token: %v", err)
	}

	return access, refresh, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, oldAccess, refresh, ip string) (string, string, error) {
	claims, err := s.security.ValidateToken(oldAccess)
	if err != nil && !errdefs.Is(err, errdefs.ErrExpiredToken) {
		return "", "", errdefs.ErrInvalidToken
	}

	storedToken, err := s.repo.GetRefreshToken(ctx, claims.UserID, claims.Jti)
	if err != nil {
		return "", "", errdefs.ErrInvalidToken
	}

	if err := s.security.ValidateRefreshToken(storedToken, refresh, claims.Jti); err != nil {
        return "", "", errdefs.ErrInvalidToken
    }

	if claims.IP != ip {
		// затычка
		s.sendIPChangeWarning(claims.UserID, ip)
	}

	newAccess, jti, err := s.security.GenerateToken(claims.UserID, ip)
	if err != nil {
		return "", "", errdefs.ErrGenerateToken
	}

	newRefresh := s.security.GenerateRefreshToken()
	newHashed, err := s.security.HashRefreshToken(newRefresh)
	if err != nil {
		return "", "", errdefs.ErrByScript
	}

	if err := s.repo.RevokeRefreshToken(ctx, storedToken.ID); err != nil {
		return "", "", err
	}

	err = s.repo.SaveRefreshToken(ctx, &models.RefreshToken{
		UserID:          claims.UserID,
		HashedToken:     newHashed,
		AccessTokenJTI:  jti,
		ClientIP:        ip,
		ExpiresAt:       time.Now().Add(
        	time.Duration(s.cfg.Jwt.RefreshExp) * time.Minute),
	})

	if err != nil {
		return "", "", err
	}

	return newAccess, newRefresh, nil
}

// Моковая реализация отправки email
func (s *AuthService) sendIPChangeWarning(userID, newIP string) {
	user, err := s.repo.GetUser(context.Background(), userID)
	if err != nil {
		return
	}

	fmt.Printf("[MOCK] Security warning email sent to %s. New login from IP: %s\n", user.Email, newIP)
}