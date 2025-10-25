package app

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

// AppRouter holds the application dependencies
type AppRouter struct {
	Pool *pgxpool.Pool
}

// NewAppRouter creates a new AppRouter instance
func NewAppRouter(pool *pgxpool.Pool) *AppRouter {
	return &AppRouter{
		Pool: pool,
	}
}
