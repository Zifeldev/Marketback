package middleware

import (
	"net/http/httptest"
	"testing"

	"context"

	"github.com/Zifeldev/marketback/service/Auth/internal/models"
	"github.com/gin-gonic/gin"
)

type stubAuth struct{ claims *models.AccessTokenClaims }

// Implement service.AuthService methods (minimal stubs)
func (s *stubAuth) Register(ctx context.Context, email, password, role string) (*models.TokenPair, error) {
	return nil, nil
}
func (s *stubAuth) Login(ctx context.Context, email, password string) (*models.TokenPair, error) {
	return nil, nil
}
func (s *stubAuth) RefreshTokens(ctx context.Context, refreshToken string) (*models.TokenPair, error) {
	return nil, nil
}
func (s *stubAuth) RevokeToken(ctx context.Context, refreshToken string) error { return nil }
func (s *stubAuth) ValidateAccessToken(tokenString string) (*models.AccessTokenClaims, error) {
	return s.claims, nil
}

func TestJWTAuthMiddleware_SetsContextAndHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	claims := &models.AccessTokenClaims{UserID: 42, Email: "seller@example.com", Role: models.RoleSeller}
	stub := &stubAuth{claims: claims}

	r.GET("/protected", JWTAuth(stub), RequireRole(models.RoleSeller), func(c *gin.Context) {
		id, _ := GetUserID(c)
		email, _ := GetUserEmail(c)
		role, _ := GetUserRole(c)
		if id != 42 || email != "seller@example.com" || role != models.RoleSeller {
			c.AbortWithStatus(500)
			return
		}
		c.Status(200)
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer token")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestRequireRole_Forbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	claims := &models.AccessTokenClaims{UserID: 1, Email: "user@example.com", Role: models.RoleUser}
	stub := &stubAuth{claims: claims}

	r.GET("/seller-only", JWTAuth(stub), RequireRole(models.RoleSeller), func(c *gin.Context) { c.Status(200) })

	req := httptest.NewRequest("GET", "/seller-only", nil)
	req.Header.Set("Authorization", "Bearer token")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 403 {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}
