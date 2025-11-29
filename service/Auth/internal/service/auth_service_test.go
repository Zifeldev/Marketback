package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Zifeldev/marketback/service/Auth/internal/config"
	"github.com/Zifeldev/marketback/service/Auth/internal/models"
	"github.com/Zifeldev/marketback/service/Auth/internal/repository"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

// --- Mocks ---
type mockUserRepo struct {
	createWithRoleFn func(ctx context.Context, email, passHash, role string) (*models.User, error)
	getByEmailFn     func(ctx context.Context, email string) (*models.User, error)
	getByIDFn        func(ctx context.Context, id int64) (*models.User, error)
}

func (m *mockUserRepo) Create(ctx context.Context, email, passHash string) (*models.User, error) {
	return m.createWithRoleFn(ctx, email, passHash, models.RoleUser)
}
func (m *mockUserRepo) CreateWithRole(ctx context.Context, email, passHash, role string) (*models.User, error) {
	return m.createWithRoleFn(ctx, email, passHash, role)
}
func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	return m.getByEmailFn(ctx, email)
}
func (m *mockUserRepo) GetByID(ctx context.Context, id int64) (*models.User, error) {
	return m.getByIDFn(ctx, id)
}
func (m *mockUserRepo) UpdateRole(ctx context.Context, id int64, role string) (*models.User, error) {
	return nil, errors.New("not implemented")
}
func (m *mockUserRepo) Delete(ctx context.Context, id int64) error {
	return errors.New("not implemented")
}
func (m *mockUserRepo) List(ctx context.Context, limit, offset int) ([]*models.User, error) {
	return nil, errors.New("not implemented")
}

type mockTokenRepo struct {
	createFn       func(ctx context.Context, userID int64, token string, expiresAt time.Time) (*models.RefreshToken, error)
	getFn          func(ctx context.Context, token string) (*models.RefreshToken, error)
	revokeFn       func(ctx context.Context, token string) error
	revokeAllFn    func(ctx context.Context, userID int64) error
	cleanupExpired func(ctx context.Context) error
}

func (m *mockTokenRepo) CreateRefreshToken(ctx context.Context, userID int64, token string, expiresAt time.Time) (*models.RefreshToken, error) {
	return m.createFn(ctx, userID, token, expiresAt)
}
func (m *mockTokenRepo) GetRefreshToken(ctx context.Context, token string) (*models.RefreshToken, error) {
	return m.getFn(ctx, token)
}
func (m *mockTokenRepo) RevokeRefreshToken(ctx context.Context, token string) error {
	return m.revokeFn(ctx, token)
}
func (m *mockTokenRepo) RevokeAllUserTokens(ctx context.Context, userID int64) error {
	return m.revokeAllFn(ctx, userID)
}
func (m *mockTokenRepo) CleanupExpiredTokens(ctx context.Context) error { return m.cleanupExpired(ctx) }

// --- Helpers ---
func testConfig() *config.JWTConfig {
	return &config.JWTConfig{
		AccessSecret:      "access-secret",
		RefreshSecret:     "refresh-secret",
		AccessExpiration:  15 * time.Minute,
		RefreshExpiration: 7 * 24 * time.Hour,
		Issuer:            "test-issuer",
		FirstAdminEmail:   "admin@example.com",
	}
}

// --- Tests ---
func TestAuthService_Register_DefaultRole(t *testing.T) {
	cfg := testConfig()
	var capturedRole string
	uRepo := &mockUserRepo{createWithRoleFn: func(ctx context.Context, email, passHash, role string) (*models.User, error) {
		capturedRole = role
		return &models.User{ID: 10, Email: email, Role: role, PasswordHash: passHash, CreatedAt: time.Now()}, nil
	}}
	tRepo := &mockTokenRepo{createFn: func(ctx context.Context, userID int64, token string, expiresAt time.Time) (*models.RefreshToken, error) {
		return &models.RefreshToken{ID: 1, UserID: userID, Token: token, ExpiresAt: expiresAt}, nil
	}, getFn: func(ctx context.Context, token string) (*models.RefreshToken, error) {
		return nil, repository.ErrTokenNotFound
	}, revokeFn: func(ctx context.Context, token string) error { return nil }, revokeAllFn: func(ctx context.Context, userID int64) error { return nil }, cleanupExpired: func(ctx context.Context) error { return nil }}

	svc := NewAuthService(cfg, uRepo, tRepo)
	tp, err := svc.Register(context.Background(), "user@example.com", "pass123", "")
	require.NoError(t, err)
	require.NotNil(t, tp)
	require.Equal(t, models.RoleUser, capturedRole)
}

func TestAuthService_Register_ExplicitRole(t *testing.T) {
	cfg := testConfig()
	uRepo := &mockUserRepo{createWithRoleFn: func(ctx context.Context, email, passHash, role string) (*models.User, error) {
		return &models.User{ID: 11, Email: email, Role: role, PasswordHash: passHash}, nil
	}}
	tRepo := &mockTokenRepo{createFn: func(ctx context.Context, userID int64, token string, expiresAt time.Time) (*models.RefreshToken, error) {
		return &models.RefreshToken{ID: 2, UserID: userID, Token: token, ExpiresAt: expiresAt}, nil
	}, getFn: func(ctx context.Context, token string) (*models.RefreshToken, error) {
		return nil, repository.ErrTokenNotFound
	}, revokeFn: func(ctx context.Context, token string) error { return nil }, revokeAllFn: func(ctx context.Context, userID int64) error { return nil }, cleanupExpired: func(ctx context.Context) error { return nil }}
	svc := NewAuthService(cfg, uRepo, tRepo)
	tp, err := svc.Register(context.Background(), "seller@example.com", "pass123", models.RoleSeller)
	require.NoError(t, err)
	// We don't decode JWT here; just ensure token pair produced and role captured by mock user
	require.NotNil(t, tp)
}

func TestAuthService_Register_TokenContainsSellerRole(t *testing.T) {
	cfg := testConfig()
	uRepo := &mockUserRepo{createWithRoleFn: func(ctx context.Context, email, passHash, role string) (*models.User, error) {
		return &models.User{ID: 100, Email: email, Role: role, PasswordHash: passHash}, nil
	}}
	tRepo := &mockTokenRepo{createFn: func(ctx context.Context, userID int64, token string, expiresAt time.Time) (*models.RefreshToken, error) {
		return &models.RefreshToken{ID: 200, UserID: userID, Token: token, ExpiresAt: expiresAt}, nil
	}, getFn: func(ctx context.Context, token string) (*models.RefreshToken, error) {
		return nil, repository.ErrTokenNotFound
	}, revokeFn: func(ctx context.Context, token string) error { return nil }, revokeAllFn: func(ctx context.Context, userID int64) error { return nil }, cleanupExpired: func(ctx context.Context) error { return nil }}

	svc := NewAuthService(cfg, uRepo, tRepo)
	tp, err := svc.Register(context.Background(), "seller.jwt@example.com", "pass12345", models.RoleSeller)
	require.NoError(t, err)
	require.NotNil(t, tp)
	require.NotEmpty(t, tp.AccessToken)

	claims, err := svc.ValidateAccessToken(tp.AccessToken)
	require.NoError(t, err)
	require.Equal(t, models.RoleSeller, claims.Role)
	require.Equal(t, int64(100), claims.UserID)
}

func TestAuthService_Register_UserExists(t *testing.T) {
	cfg := testConfig()
	uRepo := &mockUserRepo{createWithRoleFn: func(ctx context.Context, email, passHash, role string) (*models.User, error) {
		return nil, repository.ErrUserExists
	}}
	tRepo := &mockTokenRepo{createFn: func(ctx context.Context, userID int64, token string, expiresAt time.Time) (*models.RefreshToken, error) {
		return nil, errors.New("should not be called")
	}, getFn: func(ctx context.Context, token string) (*models.RefreshToken, error) {
		return nil, repository.ErrTokenNotFound
	}, revokeFn: func(ctx context.Context, token string) error { return nil }, revokeAllFn: func(ctx context.Context, userID int64) error { return nil }, cleanupExpired: func(ctx context.Context) error { return nil }}
	svc := NewAuthService(cfg, uRepo, tRepo)
	tp, err := svc.Register(context.Background(), "exists@example.com", "pass123", "")
	require.Error(t, err)
	require.Nil(t, tp)
	require.ErrorIs(t, err, repository.ErrUserExists)
}

func TestAuthService_Login_Success(t *testing.T) {
	cfg := testConfig()
	hashed, _ := bcryptGenerate("pass123")
	uRepo := &mockUserRepo{getByEmailFn: func(ctx context.Context, email string) (*models.User, error) {
		return &models.User{ID: 22, Email: email, PasswordHash: hashed, Role: models.RoleUser}, nil
	}, createWithRoleFn: func(ctx context.Context, email, passHash, role string) (*models.User, error) {
		return nil, errors.New("unused")
	}, getByIDFn: func(ctx context.Context, id int64) (*models.User, error) { return nil, errors.New("unused") }}
	tRepo := &mockTokenRepo{createFn: func(ctx context.Context, userID int64, token string, expiresAt time.Time) (*models.RefreshToken, error) {
		return &models.RefreshToken{ID: 3, UserID: userID, Token: token, ExpiresAt: expiresAt}, nil
	}, getFn: func(ctx context.Context, token string) (*models.RefreshToken, error) {
		return nil, repository.ErrTokenNotFound
	}, revokeFn: func(ctx context.Context, token string) error { return nil }, revokeAllFn: func(ctx context.Context, userID int64) error { return nil }, cleanupExpired: func(ctx context.Context) error { return nil }}
	svc := NewAuthService(cfg, uRepo, tRepo)
	tp, err := svc.Login(context.Background(), "user@example.com", "pass123")
	require.NoError(t, err)
	require.NotEmpty(t, tp.AccessToken)
	require.NotEmpty(t, tp.RefreshToken)
}

func TestAuthService_Login_InvalidCredentials(t *testing.T) {
	cfg := testConfig()
	hashed, _ := bcryptGenerate("otherpass")
	uRepo := &mockUserRepo{getByEmailFn: func(ctx context.Context, email string) (*models.User, error) {
		return &models.User{ID: 23, Email: email, PasswordHash: hashed, Role: models.RoleUser}, nil
	}, createWithRoleFn: func(ctx context.Context, email, passHash, role string) (*models.User, error) {
		return nil, errors.New("unused")
	}, getByIDFn: func(ctx context.Context, id int64) (*models.User, error) { return nil, errors.New("unused") }}
	tRepo := &mockTokenRepo{createFn: func(ctx context.Context, userID int64, token string, expiresAt time.Time) (*models.RefreshToken, error) {
		return &models.RefreshToken{}, nil
	}, getFn: func(ctx context.Context, token string) (*models.RefreshToken, error) {
		return nil, repository.ErrTokenNotFound
	}, revokeFn: func(ctx context.Context, token string) error { return nil }, revokeAllFn: func(ctx context.Context, userID int64) error { return nil }, cleanupExpired: func(ctx context.Context) error { return nil }}
	svc := NewAuthService(cfg, uRepo, tRepo)
	tp, err := svc.Login(context.Background(), "user@example.com", "wrongpass")
	require.Error(t, err)
	require.Nil(t, tp)
	require.ErrorIs(t, err, ErrInvalidCredentials)
}

func TestAuthService_RefreshTokens_Success(t *testing.T) {
	cfg := testConfig()
	user := &models.User{ID: 33, Email: "u@example.com", Role: models.RoleUser, PasswordHash: "hash"}
	uRepo := &mockUserRepo{getByIDFn: func(ctx context.Context, id int64) (*models.User, error) { return user, nil }, getByEmailFn: func(ctx context.Context, email string) (*models.User, error) { return user, nil }, createWithRoleFn: func(ctx context.Context, email, passHash, role string) (*models.User, error) { return user, nil }}
	oldRT := &models.RefreshToken{ID: 5, UserID: user.ID, Token: "oldtoken", ExpiresAt: time.Now().Add(time.Hour)}
	tRepo := &mockTokenRepo{
		getFn:    func(ctx context.Context, token string) (*models.RefreshToken, error) { return oldRT, nil },
		revokeFn: func(ctx context.Context, token string) error { oldRT.Revoked = true; return nil },
		createFn: func(ctx context.Context, userID int64, token string, expiresAt time.Time) (*models.RefreshToken, error) {
			return &models.RefreshToken{ID: 6, UserID: userID, Token: token, ExpiresAt: expiresAt}, nil
		},
		revokeAllFn:    func(ctx context.Context, userID int64) error { return nil },
		cleanupExpired: func(ctx context.Context) error { return nil },
	}
	svc := NewAuthService(cfg, uRepo, tRepo)
	tp, err := svc.RefreshTokens(context.Background(), "oldtoken")
	require.NoError(t, err)
	require.NotNil(t, tp)
	require.True(t, oldRT.Revoked)
}

func TestAuthService_RefreshTokens_Invalid(t *testing.T) {
	cfg := testConfig()
	uRepo := &mockUserRepo{getByIDFn: func(ctx context.Context, id int64) (*models.User, error) { return nil, errors.New("unused") }, getByEmailFn: func(ctx context.Context, email string) (*models.User, error) { return nil, errors.New("unused") }, createWithRoleFn: func(ctx context.Context, email, passHash, role string) (*models.User, error) {
		return nil, errors.New("unused")
	}}
	tRepo := &mockTokenRepo{getFn: func(ctx context.Context, token string) (*models.RefreshToken, error) {
		return nil, repository.ErrTokenNotFound
	}, createFn: func(ctx context.Context, userID int64, token string, expiresAt time.Time) (*models.RefreshToken, error) {
		return nil, errors.New("unused")
	}, revokeFn: func(ctx context.Context, token string) error { return nil }, revokeAllFn: func(ctx context.Context, userID int64) error { return nil }, cleanupExpired: func(ctx context.Context) error { return nil }}
	svc := NewAuthService(cfg, uRepo, tRepo)
	tp, err := svc.RefreshTokens(context.Background(), "badtoken")
	require.Error(t, err)
	require.Nil(t, tp)
}

func TestAuthService_RevokeToken(t *testing.T) {
	cfg := testConfig()
	uRepo := &mockUserRepo{createWithRoleFn: func(ctx context.Context, email, passHash, role string) (*models.User, error) {
		return &models.User{ID: 44, Email: email, Role: role}, nil
	}, getByEmailFn: func(ctx context.Context, email string) (*models.User, error) {
		return &models.User{ID: 44, Email: email, Role: models.RoleUser}, nil
	}, getByIDFn: func(ctx context.Context, id int64) (*models.User, error) {
		return &models.User{ID: id, Email: "x@x", Role: models.RoleUser}, nil
	}}
	revoked := false
	tRepo := &mockTokenRepo{revokeFn: func(ctx context.Context, token string) error { revoked = true; return nil }, getFn: func(ctx context.Context, token string) (*models.RefreshToken, error) {
		return nil, repository.ErrTokenNotFound
	}, createFn: func(ctx context.Context, userID int64, token string, expiresAt time.Time) (*models.RefreshToken, error) {
		return &models.RefreshToken{}, nil
	}, revokeAllFn: func(ctx context.Context, userID int64) error { return nil }, cleanupExpired: func(ctx context.Context) error { return nil }}
	svc := NewAuthService(cfg, uRepo, tRepo)
	err := svc.RevokeToken(context.Background(), "tkn")
	require.NoError(t, err)
	require.True(t, revoked)
}

// bcryptGenerate small helper without exposing bcrypt directly in test names
func bcryptGenerate(p string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(p), bcrypt.DefaultCost)
	return string(hash), err
}
