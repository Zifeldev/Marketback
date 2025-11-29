package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/Zifeldev/marketback/service/Auth/internal/service"
	"github.com/gin-gonic/gin"
)

const (
	HeaderAuthorization = "Authorization"
	HeaderUserID        = "X-User-ID"
	HeaderUserEmail     = "X-User-Email"
	HeaderUserRole      = "X-User-Role"
	ContextUserID       = "user_id"
	ContextUserEmail    = "user_email"
	ContextUserRole     = "user_role"
)

func JWTAuth(authService service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader(HeaderAuthorization)
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
			return
		}


		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			return
		}

		token := parts[1]
		claims, err := authService.ValidateAccessToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}


		c.Set(ContextUserID, claims.UserID)
		c.Set(ContextUserEmail, claims.Email)
		c.Set(ContextUserRole, claims.Role)


		c.Request.Header.Set(HeaderUserID, strconv.FormatInt(claims.UserID, 10))
		c.Request.Header.Set(HeaderUserEmail, claims.Email)
		c.Request.Header.Set(HeaderUserRole, claims.Role)

		c.Next()
	}
}


func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := GetUserRole(c)
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "role not found in context"})
			return
		}

		if userRole != role {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			return
		}

		c.Next()
	}
}


func GetUserID(c *gin.Context) (int64, bool) {
	userID, exists := c.Get(ContextUserID)
	if !exists {
		return 0, false
	}
	id, ok := userID.(int64)
	return id, ok
}


func GetUserEmail(c *gin.Context) (string, bool) {
	email, exists := c.Get(ContextUserEmail)
	if !exists {
		return "", false
	}
	e, ok := email.(string)
	return e, ok
}


func GetUserRole(c *gin.Context) (string, bool) {
	role, exists := c.Get(ContextUserRole)
	if !exists {
		return "", false
	}
	r, ok := role.(string)
	return r, ok
}
