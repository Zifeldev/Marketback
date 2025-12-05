package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Zifeldev/marketback/service/Market/internal/cache"
	"github.com/Zifeldev/marketback/service/Market/internal/config"
	"github.com/Zifeldev/marketback/service/Market/internal/controllers"
	"github.com/Zifeldev/marketback/service/Market/internal/db"
	"github.com/Zifeldev/marketback/service/Market/internal/logger"
	"github.com/Zifeldev/marketback/service/Market/internal/middleware"
	"github.com/Zifeldev/marketback/service/Market/internal/repository"
	"github.com/Zifeldev/marketback/service/Market/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	ginprometheus "github.com/zsais/go-gin-prometheus"

	_ "github.com/Zifeldev/marketback/service/Market/docs"
)

// Version is set at build time
var Version = "1.0.0"

// @title Market Service API
// @version 1.0
// @description Marketplace API with products, categories, cart, orders, seller and admin management
// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

func main() {
	startTime := time.Now()

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	log := logger.InitLogger(cfg.Logger.Level)
	log.Info("Starting Market Service...")

	// Initialize database
	pool, err := db.InitDB(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()
	log.Info("Database connection established")

	// Initialize Redis cache
	var redisCache *cache.RedisCache
	var redisClient *redis.Client
	if cfg.Redis.Enabled {
		redisCache, err = cache.NewRedisCache(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB)
		if err != nil {
			log.Warnf("Redis connection failed: %v", err)
			log.Warn("Service will continue without Redis features:")
			log.Warn("  - Rate limiting: DISABLED")
			log.Warn("  - Category caching: DISABLED")
			redisCache = nil
		} else {
			defer redisCache.Close()
			redisClient = redisCache.GetClient()
			log.Info("Redis connection established")
			if cfg.RateLimit.Enabled {
				log.Infof("  - Rate limiting: ENABLED (%d req/%s)", cfg.RateLimit.Max, cfg.RateLimit.Interval)
			}
			log.Info("  - Category caching: ENABLED")
		}
	} else {
		log.Info("Redis is disabled by configuration (REDIS_ENABLED=false)")
		log.Info("  - Rate limiting: DISABLED")
		log.Info("  - Category caching: DISABLED")
	}

	// Initialize repositories
	sellerRepo := repository.NewSellerRepository(pool)
	categoryRepo := repository.NewCategoryRepository(pool, redisCache)
	productRepo := repository.NewProductRepository(pool)
	cartRepo := repository.NewCartRepository(pool)
	orderRepo := repository.NewOrderRepository(pool)

	// Initialize services
	marketService := service.NewMarketService(
		orderRepo,
		cartRepo,
	)

	// Upload directory setup
	uploadDir := cfg.UploadDir
	if uploadDir == "" {
		uploadDir = "./uploads"
	}
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	// Initialize controllers
	marketController := controllers.NewMarketController(
		productRepo,
		categoryRepo,
		cartRepo,
		orderRepo,
		marketService,
	)
	sellerController := controllers.NewSellerController(
		sellerRepo,
		productRepo,
	)
	adminController := controllers.NewAdminController(
		categoryRepo,
		productRepo,
		sellerRepo,
		orderRepo,
	)
	healthController := controllers.NewHealthController(pool, redisClient, startTime, Version)
	uploadController := controllers.NewUploadController(uploadDir, baseURL)

	// Setup Gin router
	if cfg.Strict {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.Default()

	// Prometheus metrics middleware
	p := ginprometheus.NewPrometheus("market")
	p.Use(router)

	// Middleware
	router.Use(middleware.CORS())

	// Rate limiting
	if redisCache != nil && cfg.RateLimit.Enabled {
		router.Use(middleware.RateLimiter(redisCache, cfg.RateLimit.Max, cfg.RateLimit.Interval))
	}

	// Health check
	router.GET("/health", healthController.Health)

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Static files for uploaded images
	router.Static("/uploads", uploadDir)

	// API routes
	api := router.Group("/api/")
	{
		// Public routes - no authentication required
		public := api.Group("")
		{
			// Products
			public.GET("/products", marketController.GetProducts)
			public.GET("/products/:id", marketController.GetProduct)

			// Categories
			public.GET("/categories", marketController.GetCategories)
			public.GET("/categories/:id", marketController.GetCategory)
		}

		// Upload routes - authentication required
		upload := api.Group("/upload")
		upload.Use(middleware.JWTAuth(cfg.JWT.AccessSecret))
		{
			upload.POST("/image", uploadController.UploadImage)
			upload.DELETE("/image/:filename", uploadController.DeleteImage)
		}

		// Cart routes - authentication required
		cart := api.Group("/cart")
		cart.Use(middleware.JWTAuth(cfg.JWT.AccessSecret))
		{
			cart.GET("", marketController.GetCart)
			cart.POST("/items", marketController.AddToCart)
			cart.PUT("/items/:id", marketController.UpdateCartItem)
			cart.DELETE("/items/:id", marketController.DeleteCartItem)
		}

		// User routes - authentication required
		user := api.Group("/user")
		user.Use(middleware.JWTAuth(cfg.JWT.AccessSecret))
		{
			user.POST("/orders", marketController.CreateOrder)
			user.GET("/orders", marketController.GetUserOrders)
			user.GET("/orders/:id", marketController.GetOrder)
		}

		// Seller routes - seller role required
		seller := api.Group("/seller")
		seller.Use(middleware.JWTAuth(cfg.JWT.AccessSecret))
		seller.Use(middleware.RequireRole("seller", "admin"))
		{
			seller.POST("/register", sellerController.RegisterSeller)
			seller.GET("/profile", sellerController.GetSellerProfile)
			seller.PUT("/profile", sellerController.UpdateSellerProfile)
			seller.POST("/products", sellerController.CreateProduct)
			seller.GET("/products", sellerController.GetSellerProducts)
			seller.PUT("/products/:id", sellerController.UpdateProduct)
			seller.DELETE("/products/:id", sellerController.DeleteProduct)
		}

		// Admin routes - admin role required
		admin := api.Group("/admin")
		admin.Use(middleware.JWTAuth(cfg.JWT.AccessSecret))
		admin.Use(middleware.RequireRole("admin"))
		{
			admin.POST("/categories", adminController.CreateCategory)
			admin.PUT("/categories/:id", adminController.UpdateCategory)
			admin.DELETE("/categories/:id", adminController.DeleteCategory)
			admin.GET("/sellers", adminController.GetAllSellers)
			admin.PUT("/sellers/:id/status", adminController.UpdateSellerStatus)
			admin.PUT("/products/:id/status", adminController.UpdateProductStatus)
			admin.GET("/orders", adminController.GetAllOrders)
			admin.PUT("/orders/:id/status", adminController.UpdateOrderStatus)
		}
	}

	srv := &http.Server{
		Addr:    cfg.HTTP.Host,
		Handler: router,
	}

	go func() {
		log.Infof("Server starting on %s", cfg.HTTP.Host)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Info("Server exited")
}
