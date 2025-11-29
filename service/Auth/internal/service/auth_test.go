package service

import (
	"context"
	"testing"
	"time"

	"github.com/Zifeldev/marketback/service/Auth/internal/config"
	"github.com/Zifeldev/marketback/service/Auth/internal/models"
)

type fakeUserRepo struct{ user *models.User }

func (f *fakeUserRepo) Create(ctx context.Context, email, passwordHash string) (*models.User, error) {
	return f.user, nil
}
func (f *fakeUserRepo) CreateWithRole(ctx context.Context, email, passwordHash, role string) (*models.User, error) {
	f.user = &models.User{ID: 1, Email: email, PasswordHash: passwordHash, Role: role}
	return f.user, nil
}
func (f *fakeUserRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	return f.user, nil
}
func (f *fakeUserRepo) GetByID(ctx context.Context, id int64) (*models.User, error) {
	return f.user, nil
}
func (f *fakeUserRepo) UpdateRole(ctx context.Context, id int64, role string) (*models.User, error) {
	f.user.Role = role
	return f.user, nil
}
func (f *fakeUserRepo) Delete(ctx context.Context, id int64) error { return nil }
func (f *fakeUserRepo) List(ctx context.Context, limit, offset int) ([]*models.User, error) {
	return []*models.User{f.user}, nil
}

type fakeTokenRepo struct{}

func (f *fakeTokenRepo) CreateRefreshToken(ctx context.Context, userID int64, token string, expiresAt time.Time) (*models.RefreshToken, error) {
	return &models.RefreshToken{ID: 1, UserID: userID, Token: token, ExpiresAt: expiresAt, CreatedAt: time.Now()}, nil
}
func (f *fakeTokenRepo) GetRefreshToken(ctx context.Context, token string) (*models.RefreshToken, error) {
	return &models.RefreshToken{ID: 1, UserID: 1, Token: token, ExpiresAt: time.Now().Add(time.Hour)}, nil
}
func (f *fakeTokenRepo) RevokeRefreshToken(ctx context.Context, token string) error  { return nil }
func (f *fakeTokenRepo) RevokeAllUserTokens(ctx context.Context, userID int64) error { return nil }
func (f *fakeTokenRepo) CleanupExpiredTokens(ctx context.Context) error              { return nil }

func TestGenerateAndValidateAccessToken_WithSellerRole(t *testing.T) {
	cfg := &config.JWTConfig{
		AccessSecret:      "test-access-secret-32-bytes-minimum-test",
		RefreshSecret:     "test-refresh-secret-32-bytes-minimum-test",
		AccessExpiration:  time.Minute,
		RefreshExpiration: time.Hour,
		Issuer:            "auth-test",
		FirstAdminEmail:   "",
	}

	uRepo := &fakeUserRepo{}
	tRepo := &fakeTokenRepo{}
	svc := NewAuthService(cfg, uRepo, tRepo).(*authService)

	// Register with seller role
	pair, err := svc.Register(context.Background(), "seller@example.com", "password123", models.RoleSeller)
	if err != nil {
		t.Fatalf("register error: %v", err)
	}
	if pair.AccessToken == "" || pair.RefreshToken == "" {
		t.Fatalf("expected tokens to be generated")
	}

	claims, err := svc.ValidateAccessToken(pair.AccessToken)
	if err != nil {
		t.Fatalf("validate access token error: %v", err)
	}
	if claims.Role != models.RoleSeller {
		t.Fatalf("expected role seller, got %s", claims.Role)
	}
	if claims.Email != "seller@example.com" {
		t.Fatalf("expected email, got %s", claims.Email)
	}
}
