package middleware

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/regestrac/master-management-api/internal/db"
	"github.com/regestrac/master-management-api/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func RequireAuth(c *gin.Context) {
	var tokenString string

	// 1. First try to read from Authorization header
	authHeader := c.GetHeader("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		tokenString = strings.TrimPrefix(authHeader, "Bearer ")
	} else {
		// 2. Fallback: check cookie (if you sometimes set tokens as cookies)
		cookie, err := c.Cookie("Authorization")
		if err == nil {
			tokenString = cookie
		}
	}

	if tokenString == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: token missing"})
		return
	}

	// Decode / validate it
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil || !token.Valid {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: invalid token"})
		return
	}

	// Claims must be map claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: invalid claims"})
		return
	}

	// Check expiration
	if exp, ok := claims["exp"].(float64); ok {
		if float64(time.Now().Unix()) > exp {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: token expired"})
			return
		}
	}

	// Find the user with token sub
	var user models.User
	if sub, ok := claims["sub"].(float64); ok {
		db.DB.First(&user, uint(sub))
	}

	if user.ID == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: user not found"})
		return
	}

	// Attach user to request
	c.Set("user", user)

	// Continue
	c.Next()
}
