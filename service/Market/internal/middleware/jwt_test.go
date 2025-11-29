package middleware

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const testSecret = "test-secret-123"

// Test toInt with different types
func TestToIntVariants(t *testing.T) {
	cases := []struct {
		name    string
		in      interface{}
		want    int
		wantErr bool
	}{
		{"float64", float64(3.0), 3, false},
		{"int", int(5), 5, false},
		{"int64", int64(7), 7, false},
		{"string", "9", 9, false},
		{"bad string", "x", 0, true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := toInt(c.in)
			if c.wantErr {
				if err == nil {
					t.Fatalf("expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != c.want {
				t.Fatalf("got %d want %d", got, c.want)
			}
		})
	}
}

// Test middleware sets user_id and role when struct claims provided
func TestJWTAuth_SetsContextFromStructClaims(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)

	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{UserID: 42, Role: "user", RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour))}})
	signed, err := tok.SignedString([]byte(testSecret))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	// create request with header
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+signed)
	c.Request = req

	// run middleware
	h := JWTAuth(testSecret)
	h(c)

	// check context
	uid, exists := c.Get("user_id")
	if !exists {
		t.Fatalf("user_id not set in context")
	}
	if uid.(int) != 42 {
		t.Fatalf("unexpected user_id %v", uid)
	}

	role, exists := c.Get("role")
	if !exists {
		t.Fatalf("role not set in context")
	}
	if role.(string) != "user" {
		t.Fatalf("unexpected role %v", role)
	}
}

// Test middleware supports MapClaims numeric types (float64)
func TestJWTAuth_SetsContextFromMapClaimsFloat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)

	mc := jwt.MapClaims{}
	mc["user_id"] = float64(13)
	mc["role"] = "user"
	mc["exp"] = time.Now().Add(time.Hour).Unix()
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, mc)
	signed, err := tok.SignedString([]byte(testSecret))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+signed)
	c.Request = req

	h := JWTAuth(testSecret)
	h(c)

	uid, exists := c.Get("user_id")
	if !exists {
		t.Fatalf("user_id not set in context for mapclaims")
	}
	if uid.(int) != 13 {
		t.Fatalf("unexpected user_id %v", uid)
	}
}
