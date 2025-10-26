package models

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Deployment struct {
	Name           string `json:"name"`
	Url            string `json:"url"`
	Tier           string `json:"tier"`
	ContainerImage string `json:"container_image"`
	Username       string `json:"username"`
}

func MigrateDeploymentTable(pool *pgxpool.Pool) error {
	ctx := context.Background()
	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS deployments (
			name TEXT PRIMARY KEY,
			url TEXT NOT NULL,
			tier TEXT NOT NULL,
			container_image TEXT NOT NULL REFERENCES container_images(fqin),
			username TEXT NOT NULL REFERENCES users(username)
		);
	`)
	return err
}
