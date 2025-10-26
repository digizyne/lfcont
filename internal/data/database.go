package data

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/digizyne/lfcont/internal/data/models"
)

func InitializeDatabase() (*pgxpool.Pool, error) {
	log.Printf("Initializing database...")
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, "postgresql://postgres:postgres@postgres:5432/postgres")
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %v", err)
	}

	migrations := []struct {
		name string
		fn   func(*pgxpool.Pool) error
	}{
		{"users", models.MigrateUserTable},
		{"container_images", models.MigrateContainerImageTable},
		{"deployments", models.MigrateDeploymentTable},
	}

	for _, migration := range migrations {
		log.Printf("Running migration: %s", migration.name)
		if err := migration.fn(pool); err != nil {
			pool.Close()
			return nil, fmt.Errorf("failed to migrate %s table: %v", migration.name, err)
		}
	}

	log.Printf("Database migrations completed successfully")
	return pool, nil
}
