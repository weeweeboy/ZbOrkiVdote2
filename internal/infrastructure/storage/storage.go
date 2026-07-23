package storage

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	db     *pgxpool.Pool
	logger *slog.Logger
}

func configurationPool(config *pgxpool.Config) {
	config.MaxConns = int32(20)
	config.MinConns = int32(5)
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = 30 * time.Minute
	config.HealthCheckPeriod = time.Minute
	config.ConnConfig.ConnectTimeout = 5 * time.Second
}

func NewStorage(connString string, logger *slog.Logger) (*Storage, error) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	configurationPool(config)

	dbPool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}

	logger.Info("connected to database", "path to db", connString)
	return &Storage{db: dbPool}, nil
}

func (s *Storage) Close() error {
	s.db.Close()
	return nil
}

func (s *Storage) GetH(ctx context.Context, name string) (int32, error) {

	var heroId int32

	err := s.db.QueryRow(ctx, `SELECT id FROM heroes WHERE name = $1`, name).Scan(&heroId)
	if err != nil {
		return 0, fmt.Errorf("failed to execute query: %w", err)
	}

	return heroId, nil

}

func (s *Storage) GetI(ctx context.Context, id int) (string, error) {

	var itemName string

	err := s.db.QueryRow(ctx, `SELECT name FROM items WHERE id = $1`, id).Scan(&itemName)
	if err != nil {
		return "", fmt.Errorf("failed to execute query: %w", err)
	}

	return itemName, nil

}
