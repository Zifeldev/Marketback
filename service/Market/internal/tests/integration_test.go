//go:build integration
// +build integration

package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Zifeldev/marketback/service/Market/internal/controllers"
	"github.com/Zifeldev/marketback/service/Market/internal/models"
	"github.com/Zifeldev/marketback/service/Market/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// IntegrationTestSuite contains all integration tests
type IntegrationTestSuite struct {
	suite.Suite
	ctx        context.Context
	container  testcontainers.Container
	pool       *pgxpool.Pool
	router     *gin.Engine
	sellerCtrl *controllers.SellerController
	marketCtrl *controllers.MarketController
}

func TestIntegrationSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.ctx = context.Background()

	// Start PostgreSQL container
	req := testcontainers.ContainerRequest{
		Image:        "postgres:15-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpass",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections").
			WithOccurrence(2).
			WithStartupTimeout(60 * time.Second),
	}

	container, err := testcontainers.GenericContainer(s.ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	s.Require().NoError(err)
	s.container = container

	// Get host and port
	host, err := container.Host(s.ctx)
	s.Require().NoError(err)
	port, err := container.MappedPort(s.ctx, "5432")
	s.Require().NoError(err)

	// Connect to database
	connStr := fmt.Sprintf("postgres://testuser:testpass@%s:%s/testdb?sslmode=disable", host, port.Port())
	pool, err := pgxpool.New(s.ctx, connStr)
	s.Require().NoError(err)
	s.pool = pool

	// Run migrations
	s.runMigrations()

	// Setup repositories and controllers
	sellerRepo := repository.NewSellerRepository(pool)
	productRepo := repository.NewProductRepository(pool)
	cartRepo := repository.NewCartRepository(pool)
	categoryRepo := repository.NewCategoryRepository(pool, nil) // nil cache for tests
	orderRepo := repository.NewOrderRepository(pool)

	s.sellerCtrl = controllers.NewSellerController(sellerRepo, productRepo)
	s.marketCtrl = controllers.NewMarketController(productRepo, categoryRepo, cartRepo, orderRepo, nil)

	// Setup router
	gin.SetMode(gin.TestMode)
	s.router = gin.New()
	s.setupRoutes()
}

func (s *IntegrationTestSuite) TearDownSuite() {
	if s.pool != nil {
		s.pool.Close()
	}
	if s.container != nil {
		s.container.Terminate(s.ctx)
	}
}

func (s *IntegrationTestSuite) SetupTest() {
	// Clean tables before each test
	s.cleanTables()
}

func (s *IntegrationTestSuite) runMigrations() {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS sellers (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL UNIQUE,
			shop_name VARCHAR(255) NOT NULL,
			description TEXT,
			rating DECIMAL(3, 2) DEFAULT 0.00,
			is_active BOOLEAN DEFAULT false,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS categories (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS products (
			id SERIAL PRIMARY KEY,
			seller_id INTEGER NOT NULL REFERENCES sellers(id) ON DELETE CASCADE,
			category_id INTEGER REFERENCES categories(id),
			title VARCHAR(255) NOT NULL,
			description TEXT,
			price DECIMAL(10, 2) NOT NULL,
			sizes JSONB DEFAULT '[]'::jsonb,
			image_url VARCHAR(500),
			stock INTEGER DEFAULT 0,
			status VARCHAR(50) DEFAULT 'pending',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS carts (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS cart_items (
			id SERIAL PRIMARY KEY,
			cart_id INTEGER NOT NULL REFERENCES carts(id) ON DELETE CASCADE,
			product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
			quantity INTEGER NOT NULL DEFAULT 1,
			size VARCHAR(50),
			color VARCHAR(50),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(cart_id, product_id, size, color)
		)`,
		`CREATE TABLE IF NOT EXISTS orders (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL,
			total_amount DECIMAL(10, 2) NOT NULL,
			status VARCHAR(50) DEFAULT 'pending',
			payment_method VARCHAR(50),
			payment_status VARCHAR(50) DEFAULT 'pending',
			delivery_address TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS order_items (
			id SERIAL PRIMARY KEY,
			order_id INTEGER NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
			product_id INTEGER NOT NULL,
			quantity INTEGER NOT NULL,
			size VARCHAR(50),
			price DECIMAL(10, 2) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		// Insert test category
		`INSERT INTO categories (id, name, description) VALUES (1, 'Test Category', 'Test description') ON CONFLICT DO NOTHING`,
	}

	for _, migration := range migrations {
		_, err := s.pool.Exec(s.ctx, migration)
		s.Require().NoError(err)
	}
}

func (s *IntegrationTestSuite) cleanTables() {
	tables := []string{"order_items", "orders", "cart_items", "carts", "products", "sellers"}
	for _, table := range tables {
		_, _ = s.pool.Exec(s.ctx, fmt.Sprintf("TRUNCATE %s CASCADE", table))
	}
}

func (s *IntegrationTestSuite) setupRoutes() {
	api := s.router.Group("/api")

	// Seller routes
	seller := api.Group("/seller")
	seller.POST("/register", s.mockAuth(42), s.sellerCtrl.RegisterSeller)
	seller.GET("/profile", s.mockAuth(42), s.sellerCtrl.GetSellerProfile)
	seller.PUT("/profile", s.mockAuth(42), s.sellerCtrl.UpdateSellerProfile)
	seller.POST("/products", s.mockAuth(42), s.sellerCtrl.CreateProduct)
	seller.GET("/products", s.mockAuth(42), s.sellerCtrl.GetSellerProducts)
	seller.PUT("/products/:id", s.mockAuth(42), s.sellerCtrl.UpdateProduct)
	seller.DELETE("/products/:id", s.mockAuth(42), s.sellerCtrl.DeleteProduct)

	// Market routes
	api.GET("/products", s.marketCtrl.GetProducts)
	api.GET("/products/:id", s.marketCtrl.GetProduct)
	api.GET("/categories", s.marketCtrl.GetCategories)
	api.GET("/cart", s.mockAuth(42), s.marketCtrl.GetCart)
	api.POST("/cart", s.mockAuth(42), s.marketCtrl.AddToCart)
}

func (s *IntegrationTestSuite) mockAuth(userID int) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	}
}

// --- Integration Tests ---

func (s *IntegrationTestSuite) TestSellerRegistrationFlow() {
	// Register seller
	body := `{"shop_name":"Integration Test Shop","description":"Test Description"}`
	req := httptest.NewRequest("POST", "/api/seller/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Require().Equal(http.StatusCreated, w.Code)

	var seller models.Seller
	err := json.Unmarshal(w.Body.Bytes(), &seller)
	s.Require().NoError(err)
	s.Equal("Integration Test Shop", seller.ShopName)
	s.Equal(42, seller.UserID)
	s.False(seller.IsActive) // New sellers are inactive by default

	// Get seller profile
	req = httptest.NewRequest("GET", "/api/seller/profile", nil)
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Require().Equal(http.StatusOK, w.Code)
	err = json.Unmarshal(w.Body.Bytes(), &seller)
	s.Require().NoError(err)
	s.Equal("Integration Test Shop", seller.ShopName)
}

func (s *IntegrationTestSuite) TestProductCRUDFlow() {
	// First register seller
	body := `{"shop_name":"Product Test Shop","description":"Test"}`
	req := httptest.NewRequest("POST", "/api/seller/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	s.Require().Equal(http.StatusCreated, w.Code)

	// Create product
	productBody := `{"category_id":1,"title":"Test Product","description":"A test product","price":99.99,"stock":50}`
	req = httptest.NewRequest("POST", "/api/seller/products", strings.NewReader(productBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Require().Equal(http.StatusCreated, w.Code)

	var product models.Product
	err := json.Unmarshal(w.Body.Bytes(), &product)
	s.Require().NoError(err)
	s.Equal("Test Product", product.Title)
	s.Equal(99.99, product.Price)
	s.Equal(50, product.Stock)
	s.Equal("pending", product.Status)
	productID := product.ID

	// Get seller products
	req = httptest.NewRequest("GET", "/api/seller/products", nil)
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Require().Equal(http.StatusOK, w.Code)
	var products []models.Product
	err = json.Unmarshal(w.Body.Bytes(), &products)
	s.Require().NoError(err)
	s.Len(products, 1)
	s.Equal("Test Product", products[0].Title)

	// Update product
	updateBody := `{"title":"Updated Product","price":149.99}`
	req = httptest.NewRequest("PUT", fmt.Sprintf("/api/seller/products/%d", productID), strings.NewReader(updateBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Require().Equal(http.StatusOK, w.Code)
	err = json.Unmarshal(w.Body.Bytes(), &product)
	s.Require().NoError(err)
	s.Equal("Updated Product", product.Title)
	s.Equal(149.99, product.Price)

	// Delete product
	req = httptest.NewRequest("DELETE", fmt.Sprintf("/api/seller/products/%d", productID), nil)
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Require().Equal(http.StatusOK, w.Code)

	// Verify deletion - products list should be empty
	req = httptest.NewRequest("GET", "/api/seller/products", nil)
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Require().Equal(http.StatusOK, w.Code)
	err = json.Unmarshal(w.Body.Bytes(), &products)
	s.Require().NoError(err)
	s.Len(products, 0)
}

func (s *IntegrationTestSuite) TestAddToCartFlow() {
	// Setup: register seller and create product
	sellerBody := `{"shop_name":"Cart Test Shop","description":"Test"}`
	req := httptest.NewRequest("POST", "/api/seller/register", strings.NewReader(sellerBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	s.Require().Equal(http.StatusCreated, w.Code)

	productBody := `{"category_id":1,"title":"Cart Product","price":25.00,"stock":100}`
	req = httptest.NewRequest("POST", "/api/seller/products", strings.NewReader(productBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	s.Require().Equal(http.StatusCreated, w.Code)

	var product models.Product
	json.Unmarshal(w.Body.Bytes(), &product)
	productID := product.ID

	// Add to cart
	cartBody := fmt.Sprintf(`{"product_id":%d,"quantity":2,"size":"M"}`, productID)
	req = httptest.NewRequest("POST", "/api/cart", strings.NewReader(cartBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Require().Equal(http.StatusCreated, w.Code)

	// Get cart
	req = httptest.NewRequest("GET", "/api/cart", nil)
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Require().Equal(http.StatusOK, w.Code)
	var cartItems []models.CartItemWithDetails
	err := json.Unmarshal(w.Body.Bytes(), &cartItems)
	s.Require().NoError(err)
	s.Len(cartItems, 1)
	s.Equal(productID, cartItems[0].ProductID)
	s.Equal(2, cartItems[0].Quantity)
	s.Equal("M", cartItems[0].Size)
}

func (s *IntegrationTestSuite) TestGetProductsWithPagination() {
	// Setup: register seller
	sellerBody := `{"shop_name":"Pagination Test Shop","description":"Test"}`
	req := httptest.NewRequest("POST", "/api/seller/register", strings.NewReader(sellerBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	s.Require().Equal(http.StatusCreated, w.Code)

	// Create 5 products
	for i := 1; i <= 5; i++ {
		productBody := fmt.Sprintf(`{"category_id":1,"title":"Product %d","price":%d,"stock":10}`, i, i*10)
		req = httptest.NewRequest("POST", "/api/seller/products", strings.NewReader(productBody))
		req.Header.Set("Content-Type", "application/json")
		w = httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		s.Require().Equal(http.StatusCreated, w.Code)
	}

	// Get products with pagination
	req = httptest.NewRequest("GET", "/api/products?page=1&page_size=2", nil)
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Require().Equal(http.StatusOK, w.Code)

	var resp struct {
		Data       []models.ProductWithDetails `json:"data"`
		Pagination models.PaginationMeta       `json:"pagination"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	s.Require().NoError(err)
	s.Len(resp.Data, 2)
	s.Equal(int64(5), resp.Pagination.TotalItems)
	s.Equal(3, resp.Pagination.TotalPages)
}

func (s *IntegrationTestSuite) TestSellerCannotUpdateOthersProduct() {
	// This test verifies that a seller cannot update products belonging to another seller

	// Create first seller
	sellerBody := `{"shop_name":"First Shop","description":"Test"}`
	req := httptest.NewRequest("POST", "/api/seller/register", strings.NewReader(sellerBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	s.Require().Equal(http.StatusCreated, w.Code)

	// Create product
	productBody := `{"category_id":1,"title":"First Product","price":50,"stock":10}`
	req = httptest.NewRequest("POST", "/api/seller/products", strings.NewReader(productBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	s.Require().Equal(http.StatusCreated, w.Code)

	var product models.Product
	json.Unmarshal(w.Body.Bytes(), &product)
	productID := product.ID

	// Now try to update with a different user (simulating second seller)
	// We need to create a new route with different user_id
	router2 := gin.New()
	api2 := router2.Group("/api/seller")
	api2.PUT("/products/:id", func(c *gin.Context) {
		c.Set("user_id", 999) // Different user
		c.Next()
	}, s.sellerCtrl.UpdateProduct)

	updateBody := `{"title":"Hacked Product"}`
	req = httptest.NewRequest("PUT", fmt.Sprintf("/api/seller/products/%d", productID), strings.NewReader(updateBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router2.ServeHTTP(w, req)

	// Should return 403 because user 999 doesn't have a seller profile
	s.Equal(http.StatusForbidden, w.Code)
}
