package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/jackc/pgx/stdlib"
)

type Config struct {
	username, password, host, port, dbName, sslmode string
}

func NewDBConnect(ctx context.Context, cfg Config) *sql.DB {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.username, cfg.password, cfg.host, cfg.port, cfg.dbName, cfg.sslmode)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("failed to load driver: %v", err)
	}

	err = db.PingContext(ctx)
	if err != nil {
		log.Fatalf("failed to connect to db: %w", err)
	}
	return db
}
