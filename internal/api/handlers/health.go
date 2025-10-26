package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/digizyne/lfcont/tools"
)

func CheckHealth(app *tools.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		var version string
		if err := app.Pool.QueryRow(ctx, "SELECT version()").Scan(&version); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":  "failed to query postgres version",
				"detail": err.Error(),
			})
			return
		}
		fmt.Printf("Postgres version: %s", version)
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
		})
	}
}
