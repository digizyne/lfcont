package models

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ContainerImage struct {
	Fqin     string `json:"fqin"`
	Username string `json:"username"`
}

func MigrateContainerImageTable(pool *pgxpool.Pool) error {
	ctx := context.Background()
	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS container_images (
			fqin TEXT PRIMARY KEY,
			username TEXT NOT NULL REFERENCES users(username)
		);
	`)
	return err
}
