package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (app *App) CheckHealth(c *gin.Context) {
	ctx := c.Request.Context()
	if _, err := app.Pool.Exec(ctx, "SELECT version()"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  "failed to query postgres version",
			"detail": err,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"http server": "healthy",
		"database":    "healthy",
	})
}
