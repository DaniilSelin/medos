package repository

import (
	"AuthService/config"
	"AuthService/internal/models"
	"AuthService/internal/errdefs"

	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Используемые ошибки -
// ErrConflict - нарушена уникальность поля
// ErrInvalidInput - не прошли проверки целостности данных
// ErrNotFound - не найдена запись
// ErrDB - ошибка

type Repository struct {
	db  *pgxpool.Pool
	cfg *config.Config
}

func NewRepository(db *pgxpool.Pool, cfg *config.Config) *Repository {
	return &Repository{
		db:  db,
		cfg: cfg,
	}
}

func (r *Repository) RegisterUser(ctx context.Context, user *models.User) error {
    query := fmt.Sprintf(`
        INSERT INTO %s.users (id, email)
        VALUES ($1, $2)
    `, r.cfg.DB.Schema)

    _, err := r.db.Exec(ctx, query, user.ID, user.Email)
    if err != nil {
        var pgErr *pgconn.PgError
        if errors.As(err, &pgErr) {
            switch pgErr.Code {
            case "23505":
                return errdefs.Wrapf(errdefs.ErrConflict, "email '%s' already exists", user.Email)
            case "23514":
                return errdefs.Wrapf(errdefs.ErrInvalidInput, "validation failed: %s", pgErr.Message)
            }
        }
        return errdefs.Wrapf(errdefs.ErrDB, "failed to register user: %v", err)
    }
    return nil
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	query := fmt.Sprintf(`
		SELECT id, email 
		FROM %s.users 
		WHERE email = $1
	`, r.cfg.DB.Schema)

	var user models.User
	err := r.db.QueryRow(ctx, query, email).Scan(&user.ID, &user.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errdefs.ErrNotFound
		}
		return nil, errdefs.Wrapf(errdefs.ErrDB, "failed to get user: %v", err)
	}
	return &user, nil
}

func (r *Repository) GetUser(ctx context.Context, userID string) (*models.User, error) {
	query := fmt.Sprintf(`
		SELECT id, email 
		FROM %s.users 
		WHERE id = $1
	`, r.cfg.DB.Schema)

	var user models.User
	err := r.db.QueryRow(ctx, query, userID).Scan(&user.ID, &user.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errdefs.ErrNotFound
		}
		return nil, errdefs.Wrapf(errdefs.ErrDB, "failed to get user: %v", err)
	}
	return &user, nil
}

func (r *Repository) SaveRefreshToken(ctx context.Context, token *models.RefreshToken) error {
	query := fmt.Sprintf(`
		INSERT INTO %s.refresh_tokens 
		(user_id, hashed_refresh_token, access_token_jti, client_ip, expires_at)
		VALUES ($1, $2, $3, $4, $5)
	`, r.cfg.DB.Schema)

	_, err := r.db.Exec(
		ctx,
		query,
		token.UserID,
		token.HashedToken,
		token.AccessTokenJTI,
		token.ClientIP,
		token.ExpiresAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return errdefs.Wrapf(errdefs.ErrConflict, "jti '%s' already exists", token.AccessTokenJTI)
		}
		return errdefs.Wrapf(errdefs.ErrDB, "failed to save refresh token: %v", err)
	}
	return nil
}

func (r *Repository) GetRefreshToken(ctx context.Context, userID, jti string) (*models.RefreshToken, error) {
	query := fmt.Sprintf(`
		SELECT id, user_id, hashed_refresh_token, access_token_jti, client_ip, expires_at
		FROM %s.refresh_tokens
		WHERE user_id = $1 AND access_token_jti = $2
	`, r.cfg.DB.Schema)

	var token models.RefreshToken
	err := r.db.QueryRow(ctx, query, userID, jti).Scan(
		&token.ID,
		&token.UserID,
		&token.HashedToken,
		&token.AccessTokenJTI,
		&token.ClientIP,
		&token.ExpiresAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errdefs.ErrNotFound
		}
		return nil, errdefs.Wrapf(errdefs.ErrDB, "failed to get refresh token: %v", err)
	}
	return &token, nil
}

func (r *Repository) RevokeRefreshToken(ctx context.Context, tokenID int) error {
	query := fmt.Sprintf(`
		DELETE FROM %s.refresh_tokens 
		WHERE id = $1
	`, r.cfg.DB.Schema)

	_, err := r.db.Exec(ctx, query, tokenID)
	if err != nil {
		return errdefs.Wrapf(errdefs.ErrDB, "failed to revoke refresh token: %v", err)
	}
	return nil
}