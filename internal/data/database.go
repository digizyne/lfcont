package data

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"

	models "github.com/digizyne/lfcont/internal/data/models"
)

func InitializeDatabase() (*pgxpool.Pool, error) {
	log.Printf("Initializing database...")
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, "postgresql://postgres:postgres@postgres:5432/postgres")
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %v", err)
	}

	err = models.MigrateUserTable(pool)
	if err != nil {
		pool.Close() // Close on error
		return nil, fmt.Errorf("failed to migrate user table: %v", err)
	}

	return pool, nil
}
