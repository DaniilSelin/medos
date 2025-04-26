package domain

import (
	"AuthService/config"
	
	"context"
	"time"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, error) {
	if err := CreateDataBase(cfg); err != nil {
        return nil, fmt.Errorf("database creation failed: %w", err)
    }

	poolConfig, err := pgxpool.ParseConfig(cfg.DB.ConnString())
	if err != nil {
		return nil, err
	}

	poolConfig.MaxConns = int32(cfg.DB.Pool.MaxConns)
	poolConfig.MinConns = int32(cfg.DB.Pool.MinConns)
	poolConfig.MaxConnLifetime = time.Duration(cfg.DB.Pool.MaxConnLifetime) * time.Second
	poolConfig.MaxConnIdleTime = time.Duration(cfg.DB.Pool.MaxConnIdleTime) * time.Second
	poolConfig.HealthCheckPeriod = time.Duration(cfg.DB.Pool.HealthCheckPeriod) * time.Second

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		pool, err = Retry(poolConfig,
			cfg.DB.ConnectRetries,
			time.Duration(cfg.DB.ConnectRetryDelay)*time.Second,
		)
		if err != nil {
			return nil, err
		}
	}
	return pool, nil
}

func Retry(poolConfig *pgxpool.Config, retries int, delay time.Duration) (*pgxpool.Pool, error) {
	var err error
	for i := 0; i < retries; i++ {
		pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
		if err == nil {
			return pool, nil
		}
		time.Sleep(delay)
	}
	return nil, err
}
