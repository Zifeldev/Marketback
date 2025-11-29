package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Zifeldev/marketback/service/Market/internal/logger"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID int    `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func JWTAuth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims := &Claims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			logger.GetLogger().WithField("err", err).Warn("invalid or expired token")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			c.Abort()
			return
		}

		if claims.UserID != 0 {
			c.Set("user_id", claims.UserID)
			c.Set("role", claims.Role)
			c.Next()
			return
		}

		if mc, ok := token.Claims.(jwt.MapClaims); ok {
			if v, exists := mc["user_id"]; exists {
				uid, convErr := toInt(v)
				if convErr != nil {
					logger.GetLogger().WithField("err", convErr).Warn("invalid user_id in token")
					c.Error(fmt.Errorf("invalid user_id in token: %w", convErr))
					c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token user_id"})
					c.Abort()
					return
				}
				c.Set("user_id", uid)
			}
			if rv, ok := mc["role"]; ok {
				c.Set("role", fmt.Sprintf("%v", rv))
			}
		}

		c.Next()
	}
}

func JWTAuthOptional(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Next()
			return
		}

		tokenString := parts[1]
		claims := &Claims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtSecret), nil
		})

		if err == nil && token.Valid {
			if claims.UserID != 0 {
				c.Set("user_id", claims.UserID)
				c.Set("role", claims.Role)
			} else if mc, ok := token.Claims.(jwt.MapClaims); ok {
				if v, exists := mc["user_id"]; exists {
					if uid, convErr := toInt(v); convErr == nil {
						c.Set("user_id", uid)
					} else {
						logger.GetLogger().WithField("err", convErr).Warn("invalid user_id in optional token")
					}
				}
				if rv, ok := mc["role"]; ok {
					c.Set("role", fmt.Sprintf("%v", rv))
				}
			}
		}

		c.Next()
	}
}

func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "role not found in token"})
			c.Abort()
			return
		}

		userRole := fmt.Sprintf("%v", role)
		for _, r := range roles {
			if userRole == r {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		c.Abort()
	}
}

func toInt(v interface{}) (int, error) {
	switch t := v.(type) {
	case float64:
		return int(t), nil
	case float32:
		return int(t), nil
	case int:
		return t, nil
	case int64:
		return int(t), nil
	case uint64:
		return int(t), nil
	case string:
		i, err := strconv.Atoi(t)
		if err != nil {
			return 0, fmt.Errorf("cannot parse string to int: %w", err)
		}
		return i, nil
	default:
		return 0, fmt.Errorf("unsupported user_id type: %T", v)
	}
}
