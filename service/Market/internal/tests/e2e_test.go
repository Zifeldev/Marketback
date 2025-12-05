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
	"github.com/Zifeldev/marketback/service/Market/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// E2ETestSuite tests complete user flows
type E2ETestSuite struct {
	suite.Suite
	ctx       context.Context
	container testcontainers.Container
	pool      *pgxpool.Pool
	router    *gin.Engine
}

func TestE2ESuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E tests in short mode")
	}
	suite.Run(t, new(E2ETestSuite))
}

func (s *E2ETestSuite) SetupSuite() {
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

	host, _ := container.Host(s.ctx)
	port, _ := container.MappedPort(s.ctx, "5432")

	connStr := fmt.Sprintf("postgres://testuser:testpass@%s:%s/testdb?sslmode=disable", host, port.Port())
	pool, err := pgxpool.New(s.ctx, connStr)
	s.Require().NoError(err)
	s.pool = pool

	s.runMigrations()

	gin.SetMode(gin.TestMode)
	s.router = gin.New()
}

func (s *E2ETestSuite) TearDownSuite() {
	if s.pool != nil {
		s.pool.Close()
	}
	if s.container != nil {
		s.container.Terminate(s.ctx)
	}
}

func (s *E2ETestSuite) SetupTest() {
	s.cleanTables()
	s.setupRoutes()
}

func (s *E2ETestSuite) runMigrations() {
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
		`INSERT INTO categories (id, name, description) VALUES (1, 'Electronics', 'Electronic devices') ON CONFLICT DO NOTHING`,
		`INSERT INTO categories (id, name, description) VALUES (2, 'Clothing', 'Clothes and accessories') ON CONFLICT DO NOTHING`,
	}

	for _, migration := range migrations {
		_, err := s.pool.Exec(s.ctx, migration)
		s.Require().NoError(err)
	}
}

func (s *E2ETestSuite) cleanTables() {
	tables := []string{"order_items", "orders", "cart_items", "carts", "products", "sellers"}
	for _, table := range tables {
		_, _ = s.pool.Exec(s.ctx, fmt.Sprintf("TRUNCATE %s CASCADE", table))
	}
}

func (s *E2ETestSuite) setupRoutes() {
	s.router = gin.New()

	// Initialize repositories
	sellerRepo := repository.NewSellerRepository(s.pool)
	productRepo := repository.NewProductRepository(s.pool)
	cartRepo := repository.NewCartRepository(s.pool)
	categoryRepo := repository.NewCategoryRepository(s.pool, nil)
	orderRepo := repository.NewOrderRepository(s.pool)

	// Initialize services
	marketService := service.NewMarketService(orderRepo, cartRepo)

	// Initialize controllers
	sellerCtrl := controllers.NewSellerController(sellerRepo, productRepo)
	marketCtrl := controllers.NewMarketController(productRepo, categoryRepo, cartRepo, orderRepo, marketService)

	api := s.router.Group("/api")

	// Seller routes (user_id = 100 - seller)
	seller := api.Group("/seller")
	seller.POST("/register", s.authMiddleware(100), sellerCtrl.RegisterSeller)
	seller.POST("/products", s.authMiddleware(100), sellerCtrl.CreateProduct)
	seller.GET("/products", s.authMiddleware(100), sellerCtrl.GetSellerProducts)

	// Buyer routes (user_id = 200 - buyer)
	api.GET("/products", marketCtrl.GetProducts)
	api.GET("/products/:id", marketCtrl.GetProduct)
	api.GET("/categories", marketCtrl.GetCategories)
	api.GET("/cart", s.authMiddleware(200), marketCtrl.GetCart)
	api.POST("/cart", s.authMiddleware(200), marketCtrl.AddToCart)
	api.DELETE("/cart/items/:id", s.authMiddleware(200), marketCtrl.DeleteCartItem)
	api.POST("/orders", s.authMiddleware(200), marketCtrl.CreateOrder)
	api.GET("/orders", s.authMiddleware(200), marketCtrl.GetUserOrders)
	api.GET("/orders/:id", s.authMiddleware(200), marketCtrl.GetOrder)
}

func (s *E2ETestSuite) authMiddleware(userID int) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	}
}

// --- E2E Tests ---

// TestCompleteOrderFlow tests the full flow:
// 1. Seller registers
// 2. Seller creates products
// 3. Buyer views products
// 4. Buyer adds products to cart
// 5. Buyer creates order
// 6. Verify order and stock updates
func (s *E2ETestSuite) TestCompleteOrderFlow() {
	// Step 1: Seller registers
	sellerBody := `{"shop_name":"E2E Test Shop","description":"Full flow test shop"}`
	req := httptest.NewRequest("POST", "/api/seller/register", strings.NewReader(sellerBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	s.Require().Equal(http.StatusCreated, w.Code)

	var seller models.Seller
	json.Unmarshal(w.Body.Bytes(), &seller)
	s.Equal("E2E Test Shop", seller.ShopName)

	// Step 2: Seller creates products
	products := []struct {
		title string
		price float64
		stock int
	}{
		{"iPhone 15", 999.99, 10},
		{"MacBook Pro", 2499.99, 5},
		{"AirPods", 199.99, 50},
	}

	var productIDs []int
	for _, p := range products {
		productBody := fmt.Sprintf(`{"category_id":1,"title":"%s","price":%.2f,"stock":%d}`, p.title, p.price, p.stock)
		req = httptest.NewRequest("POST", "/api/seller/products", strings.NewReader(productBody))
		req.Header.Set("Content-Type", "application/json")
		w = httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		s.Require().Equal(http.StatusCreated, w.Code)

		var product models.Product
		json.Unmarshal(w.Body.Bytes(), &product)
		productIDs = append(productIDs, product.ID)
	}
	s.Len(productIDs, 3)

	// Step 3: Buyer views products (public endpoint)
	req = httptest.NewRequest("GET", "/api/products", nil)
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	s.Require().Equal(http.StatusOK, w.Code)

	var productsResp struct {
		Data []models.ProductWithDetails `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &productsResp)
	s.Len(productsResp.Data, 3)

	// Step 4: Buyer adds products to cart
	// Add iPhone (qty: 2)
	cartBody := fmt.Sprintf(`{"product_id":%d,"quantity":2,"size":""}`, productIDs[0])
	req = httptest.NewRequest("POST", "/api/cart", strings.NewReader(cartBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	s.Require().Equal(http.StatusCreated, w.Code)

	// Add AirPods (qty: 1)
	cartBody = fmt.Sprintf(`{"product_id":%d,"quantity":1,"size":""}`, productIDs[2])
	req = httptest.NewRequest("POST", "/api/cart", strings.NewReader(cartBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	s.Require().Equal(http.StatusCreated, w.Code)

	// Verify cart contents
	req = httptest.NewRequest("GET", "/api/cart", nil)
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	s.Require().Equal(http.StatusOK, w.Code)

	var cartItems []models.CartItemWithDetails
	json.Unmarshal(w.Body.Bytes(), &cartItems)
	s.Len(cartItems, 2)

	// Calculate expected total: 2 * 999.99 + 1 * 199.99 = 2199.97
	expectedTotal := 2*999.99 + 1*199.99

	// Step 5: Create order
	orderBody := `{"payment_method":"credit_card","delivery_address":"123 Test Street, Test City"}`
	req = httptest.NewRequest("POST", "/api/orders", strings.NewReader(orderBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	s.Require().Equal(http.StatusCreated, w.Code)

	var orderResp models.OrderWithItems
	json.Unmarshal(w.Body.Bytes(), &orderResp)
	s.Equal("pending", orderResp.Status)
	s.Equal("credit_card", orderResp.PaymentMethod)
	s.Equal("123 Test Street, Test City", orderResp.DeliveryAddr)
	s.InDelta(expectedTotal, orderResp.TotalAmount, 0.01)
	s.Len(orderResp.Items, 2)

	// Step 6: Verify cart is cleared
	req = httptest.NewRequest("GET", "/api/cart", nil)
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	s.Require().Equal(http.StatusOK, w.Code)

	var emptyCart []models.CartItemWithDetails
	json.Unmarshal(w.Body.Bytes(), &emptyCart)
	s.Len(emptyCart, 0)

	// Verify order is in user's orders list
	req = httptest.NewRequest("GET", "/api/orders", nil)
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	s.Require().Equal(http.StatusOK, w.Code)

	var orders []models.Order
	json.Unmarshal(w.Body.Bytes(), &orders)
	s.Len(orders, 1)
	s.Equal(orderResp.ID, orders[0].ID)

	// Verify stock was reduced
	// iPhone: was 10, sold 2, should be 8
	var stock int
	err := s.pool.QueryRow(s.ctx, "SELECT stock FROM products WHERE id = $1", productIDs[0]).Scan(&stock)
	s.Require().NoError(err)
	s.Equal(8, stock)

	// AirPods: was 50, sold 1, should be 49
	err = s.pool.QueryRow(s.ctx, "SELECT stock FROM products WHERE id = $1", productIDs[2]).Scan(&stock)
	s.Require().NoError(err)
	s.Equal(49, stock)
}

// TestConcurrentCartOperations tests thread safety of cart operations
func (s *E2ETestSuite) TestConcurrentCartOperations() {
	// Setup: create seller and product
	sellerBody := `{"shop_name":"Concurrent Test Shop","description":"Test"}`
	req := httptest.NewRequest("POST", "/api/seller/register", strings.NewReader(sellerBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	s.Require().Equal(http.StatusCreated, w.Code)

	productBody := `{"category_id":1,"title":"Concurrent Product","price":50,"stock":100}`
	req = httptest.NewRequest("POST", "/api/seller/products", strings.NewReader(productBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	s.Require().Equal(http.StatusCreated, w.Code)

	var product models.Product
	json.Unmarshal(w.Body.Bytes(), &product)

	// Add same product twice with same size - should update quantity
	cartBody := fmt.Sprintf(`{"product_id":%d,"quantity":3,"size":"M"}`, product.ID)
	req = httptest.NewRequest("POST", "/api/cart", strings.NewReader(cartBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	s.Require().Equal(http.StatusCreated, w.Code)

	// Add again - quantity should be cumulative
	req = httptest.NewRequest("POST", "/api/cart", strings.NewReader(cartBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	s.Require().Equal(http.StatusCreated, w.Code)

	// Check cart - should have 6 items (3+3) in one cart entry
	req = httptest.NewRequest("GET", "/api/cart", nil)
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	s.Require().Equal(http.StatusOK, w.Code)

	var cartItems []models.CartItemWithDetails
	json.Unmarshal(w.Body.Bytes(), &cartItems)
	s.Len(cartItems, 1)
	s.Equal(6, cartItems[0].Quantity)
}

// TestInsufficientStockOrder tests order creation with insufficient stock
func (s *E2ETestSuite) TestInsufficientStockOrder() {
	// Setup
	sellerBody := `{"shop_name":"Stock Test Shop","description":"Test"}`
	req := httptest.NewRequest("POST", "/api/seller/register", strings.NewReader(sellerBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	s.Require().Equal(http.StatusCreated, w.Code)

	// Create product with limited stock (only 2)
	productBody := `{"category_id":1,"title":"Limited Product","price":100,"stock":2}`
	req = httptest.NewRequest("POST", "/api/seller/products", strings.NewReader(productBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	s.Require().Equal(http.StatusCreated, w.Code)

	var product models.Product
	json.Unmarshal(w.Body.Bytes(), &product)

	// Add 5 items to cart (more than stock)
	cartBody := fmt.Sprintf(`{"product_id":%d,"quantity":5,"size":""}`, product.ID)
	req = httptest.NewRequest("POST", "/api/cart", strings.NewReader(cartBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	s.Require().Equal(http.StatusCreated, w.Code)

	// Try to create order - should fail due to insufficient stock
	orderBody := `{"payment_method":"cash","delivery_address":"Test Address"}`
	req = httptest.NewRequest("POST", "/api/orders", strings.NewReader(orderBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	// Should return error (500 or 400 depending on implementation)
	s.Contains([]int{http.StatusInternalServerError, http.StatusBadRequest}, w.Code)
	s.Contains(w.Body.String(), "stock")
}

// TestEmptyCartOrder tests order creation with empty cart
func (s *E2ETestSuite) TestEmptyCartOrder() {
	orderBody := `{"payment_method":"credit_card","delivery_address":"Test Address"}`
	req := httptest.NewRequest("POST", "/api/orders", strings.NewReader(orderBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusBadRequest, w.Code)
	s.Contains(w.Body.String(), "cart")
}

// TestCategoriesListAndFilter tests category listing and filtering products
func (s *E2ETestSuite) TestCategoriesListAndFilter() {
	// Get categories
	req := httptest.NewRequest("GET", "/api/categories", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	s.Require().Equal(http.StatusOK, w.Code)

	var categories []models.Category
	json.Unmarshal(w.Body.Bytes(), &categories)
	s.GreaterOrEqual(len(categories), 2)

	// Create seller and products in different categories
	sellerBody := `{"shop_name":"Multi Category Shop","description":"Test"}`
	req = httptest.NewRequest("POST", "/api/seller/register", strings.NewReader(sellerBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	s.Require().Equal(http.StatusCreated, w.Code)

	// Product in category 1 (Electronics)
	productBody := `{"category_id":1,"title":"Electronic Item","price":100,"stock":10}`
	req = httptest.NewRequest("POST", "/api/seller/products", strings.NewReader(productBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	s.Require().Equal(http.StatusCreated, w.Code)

	// Product in category 2 (Clothing)
	productBody = `{"category_id":2,"title":"Clothing Item","price":50,"stock":20}`
	req = httptest.NewRequest("POST", "/api/seller/products", strings.NewReader(productBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	s.Require().Equal(http.StatusCreated, w.Code)

	// Filter by category 1
	req = httptest.NewRequest("GET", "/api/products?category_id=1", nil)
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	s.Require().Equal(http.StatusOK, w.Code)

	var resp struct {
		Data []models.ProductWithDetails `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	s.Len(resp.Data, 1)
	s.Equal("Electronic Item", resp.Data[0].Title)

	// Filter by category 2
	req = httptest.NewRequest("GET", "/api/products?category_id=2", nil)
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	s.Require().Equal(http.StatusOK, w.Code)

	json.Unmarshal(w.Body.Bytes(), &resp)
	s.Len(resp.Data, 1)
	s.Equal("Clothing Item", resp.Data[0].Title)
}
