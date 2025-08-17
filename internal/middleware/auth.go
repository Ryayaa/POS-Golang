package middleware

import (
	"POS-Golang/internal/utils"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.ErrorResponse(c, "Authorization header is required", nil)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader { // If no "Bearer " prefix
			utils.ErrorResponse(c, "Invalid authorization format", nil)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			utils.ErrorResponse(c, "Invalid or expired token", err)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// Ambil claims dan set ke context
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			if userID, ok := claims["user_id"]; ok {
				c.Set("user_id", userID)
			}
			if role, ok := claims["role"]; ok {
				c.Set("role", role)
			}
		}

		c.Next()
	}
}

func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")
		if !exists || userRole != "admin" {
			utils.ErrorResponse(c, "Admin access required", nil)
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		c.Next()
	}
}
