package api

import (
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthRequestBody struct {
	Username string `json:"username" binding:"required,min=8,max=32"`
	Password string `json:"password" binding:"required,min=16"`
}

type UserClaims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func (app *App) register(c *gin.Context) {
	var req AuthRequestBody
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"error":   "invalid request payload",
			"message": err.Error(),
		})
		return
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(500, gin.H{
			"error":   "failed to hash password",
			"message": err.Error(),
		})
		return
	}
	ctx := c.Request.Context()
	_, err = app.Pool.Exec(ctx, "INSERT INTO users (username, password_hash) VALUES ($1, $2)", req.Username, string(hashedPassword))
	if err != nil {
		c.JSON(500, gin.H{
			"error":   "failed to create user",
			"message": err.Error(),
		})
		return
	}
	c.JSON(201, gin.H{
		"message": "user registered successfully",
	})
}

func (app *App) login(c *gin.Context) {
	var req AuthRequestBody
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"error":   "invalid request payload",
			"message": err.Error(),
		})
		return
	}
	ctx := c.Request.Context()
	var storedHashedPassword string
	err := app.Pool.QueryRow(ctx, "SELECT password_hash FROM users WHERE username = $1", req.Username).Scan(&storedHashedPassword)
	if err != nil {
		c.JSON(401, gin.H{
			"error":   "unauthorized",
			"message": "invalid username or password",
		})
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(storedHashedPassword), []byte(req.Password)); err != nil {
		c.JSON(401, gin.H{
			"error":   "unauthorized",
			"message": "invalid username or password",
		})
		return
	}

	var jwtSecret = os.Getenv("JWT_SECRET")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, UserClaims{
		Username: req.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
	})
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		c.JSON(500, gin.H{
			"error":   "failed to generate token",
			"message": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"token": tokenString,
	})
}
