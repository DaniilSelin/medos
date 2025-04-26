package security

import (
	"AuthService/internal/errdefs"
	"AuthService/internal/models"
	"AuthService/config"

	"crypto/rand"
	"encoding/base64"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"github.com/google/uuid"
)

// Использованные ошибки -
// ErrInvalidToken
// ErrExpiredToken
// UnExp

type Security struct {
	cfg *config.Config
}

func NewSecurity(cfg *config.Config) *Security {
	return &Security{
		cfg: cfg,
	}
}

func (s *Security) GenerateToken(userID string, ip string) (string, string, error) {
	expirationTime := time.Now().Add(time.Duration(s.cfg.Jwt.Expiration) * time.Minute)
	
	claims := &models.Claims{
		UserID: userID,
		Jti:uuid.New().String(),
		IP: ip,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	tStr, err := token.SignedString([]byte(s.cfg.Jwt.SecretKey))
	return tStr, claims.Jti, err
}

func (s *Security) ValidateToken(tokenString string) (*models.Claims, error) {
	claims := &models.Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if token.Method.Alg() != jwt.SigningMethodHS512.Alg() {
			return nil, errdefs.ErrInvalidToken
		}
		return []byte(s.cfg.Jwt.SecretKey), nil
	})
	
	if err != nil {
		if errdefs.Is(err, jwt.ErrTokenExpired) {
			return nil, errdefs.ErrExpiredToken
		}
		return nil, errdefs.ErrInvalidToken
	}
	
	if !token.Valid {
		return nil, errdefs.ErrInvalidToken
	}
	
	return claims, nil
}

func (s *Security) GenerateRefreshToken() string {
    b := make([]byte, 32)
    rand.Read(b)
    return base64.RawURLEncoding.EncodeToString(b)
}

// я им не пользуюсь, поэтму его я не тестировал
func (s *Security) RefreshToken(oldToken string) (string, string, error) {
    claims, err := s.ValidateToken(oldToken)
    if err != nil && !errdefs.Is(err, errdefs.ErrExpiredToken) {
        return "", "", err
    }
    
    newAccess, newJTI, err := s.GenerateToken(claims.UserID, claims.IP)
    if err != nil {
        return "", "", err
    }
    
    return newAccess, newJTI, nil
}

func (s *Security) HashRefreshToken(token string) (string, error) {
    hash, err := bcrypt.GenerateFromPassword([]byte(token), bcrypt.DefaultCost)
    return string(hash), err
}

func (s *Security) ValidateRefreshToken(storedToken *models.RefreshToken, inputRefresh string, accessJTI string) error {
    if storedToken.AccessTokenJTI != accessJTI {
        return errdefs.ErrInvalidToken
    }
    
    return bcrypt.CompareHashAndPassword(
        []byte(storedToken.HashedToken), 
        []byte(inputRefresh),
    )
}