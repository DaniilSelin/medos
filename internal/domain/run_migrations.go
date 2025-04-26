package domain

import (
	"AuthService/config"

	"context"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v5/pgxpool"
)

func RunMigrations(ctx context.Context, cfg *config.Config, conn *pgxpool.Pool) error {
	// Лучше вынести в конфигурацию, либо же просто запускать все *.sql
	// Но такой подход мне показался на короткой дистанции более удобным
	files := []string{
		"internal/domain/migrations/create_user.sql",
		"internal/domain/migrations/refreshToken.sql",
	}
	for _, file := range files {
		sqlContent, err := ioutil.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read SQL file %s: %w", file, err)
		}

		sqlQuery := fmt.Sprintf(string(sqlContent), cfg.DB.Schema, cfg.DB.Schema)

		_, err = conn.Exec(ctx, sqlQuery)
		if err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", file, err)
		}

		log.Printf("Successfully executed migration: %s", file)
	}

	return nil
}

func CreateDataBase(cfg *config.Config) error {
    sysDsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=postgres sslmode=%s",
        cfg.DB.Host, cfg.DB.Port, cfg.DB.User, cfg.DB.Password, cfg.DB.Sslmode)

    conn, err := pgx.Connect(context.Background(), sysDsn)
    if err != nil {
        return fmt.Errorf("failed to connect to system database: %w", err)
    }
    defer conn.Close(context.Background())

    var exists bool
    err = conn.QueryRow(context.Background(), 
        "SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = $1)", 
        cfg.DB.Dbname).Scan(&exists)
    if err != nil {
        return fmt.Errorf("failed to check database existence: %w", err)
    }

    if !exists {
        _, err = conn.Exec(context.Background(), fmt.Sprintf(`
            CREATE DATABASE %s 
            WITH ENCODING = 'UTF8'
            LC_COLLATE = 'en_US.UTF-8'
            LC_CTYPE = 'en_US.UTF-8'
            TEMPLATE = template0;`, 
            pgx.Identifier{cfg.DB.Dbname}.Sanitize()))
        
        if err != nil {
            return fmt.Errorf("failed to create database: %w", err)
        }
        log.Printf("Database %s created successfully", cfg.DB.Dbname)
    }

    return nil
}