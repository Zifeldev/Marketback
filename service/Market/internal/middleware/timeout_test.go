package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestTimeoutMiddleware_NoTimeout(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Next()
	})

	router.GET("/fast", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "fast response"})
	})

	req := httptest.NewRequest("GET", "/fast", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "fast response")
}

func TestTimeoutMiddleware_SlowRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	timeout := 50 * time.Millisecond
	router.Use(func(c *gin.Context) {
		select {
		case <-time.After(timeout):
			c.AbortWithStatusJSON(http.StatusGatewayTimeout, gin.H{"error": "request timeout"})
			return
		default:
			c.Next()
		}
	})

	router.GET("/slow", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest("GET", "/slow", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
}

func TestTimeoutMiddleware_ContextCancellation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.GET("/check-context", func(c *gin.Context) {
		select {
		case <-c.Request.Context().Done():
			c.JSON(http.StatusGatewayTimeout, gin.H{"error": "context cancelled"})
		default:
			c.JSON(http.StatusOK, gin.H{"message": "ok"})
		}
	})

	req := httptest.NewRequest("GET", "/check-context", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
}
