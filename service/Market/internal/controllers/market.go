package controllers

import (
	"net/http"
	"strconv"

	"github.com/Zifeldev/marketback/service/Market/internal/metrics"
	"github.com/Zifeldev/marketback/service/Market/internal/models"
	"github.com/Zifeldev/marketback/service/Market/internal/repository"
	"github.com/Zifeldev/marketback/service/Market/internal/service"
	"github.com/gin-gonic/gin"
)

type MarketController struct {
	productRepo   repository.ProductRepo
	categoryRepo  repository.CategoryRepo
	cartRepo      repository.CartRepo
	orderRepo     repository.OrderRepo
	marketService *service.MarketService
}

func NewMarketController(
	productRepo repository.ProductRepo,
	categoryRepo repository.CategoryRepo,
	cartRepo repository.CartRepo,
	orderRepo repository.OrderRepo,
	marketService *service.MarketService,
) *MarketController {
	return &MarketController{
		productRepo:   productRepo,
		categoryRepo:  categoryRepo,
		cartRepo:      cartRepo,
		orderRepo:     orderRepo,
		marketService: marketService,
	}
}

func (mc *MarketController) GetProducts(c *gin.Context) {
	var categoryID, sellerID *int
	status := c.Query("status")

	if catIDStr := c.Query("category_id"); catIDStr != "" {
		if catID, err := strconv.Atoi(catIDStr); err == nil {
			categoryID = &catID
		}
	}

	if sellIDStr := c.Query("seller_id"); sellIDStr != "" {
		if sellID, err := strconv.Atoi(sellIDStr); err == nil {
			sellerID = &sellID
		}
	}

	var pagination models.PaginationParams
	if err := c.ShouldBindQuery(&pagination); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid pagination parameters"})
		return
	}

	products, totalItems, err := mc.productRepo.GetAll(c.Request.Context(), categoryID, sellerID, status, &pagination)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := models.PaginatedResponse{
		Data:       products,
		Pagination: models.NewPaginationMeta(pagination.Page, pagination.GetLimit(), totalItems),
	}

	c.JSON(http.StatusOK, response)
}

func (mc *MarketController) GetProduct(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product ID"})
		return
	}

	product, err := mc.productRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}

	metrics.ProductsViewedTotal.Inc()

	c.JSON(http.StatusOK, product)
}

func (mc *MarketController) GetCategories(c *gin.Context) {
	categories, err := mc.categoryRepo.GetAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, categories)
}

func (mc *MarketController) GetCategory(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category ID"})
		return
	}

	category, err := mc.categoryRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "category not found"})
		return
	}

	c.JSON(http.StatusOK, category)
}

func (mc *MarketController) GetCart(c *gin.Context) {
	userID, _ := c.Get("user_id")

	cartItems, err := mc.cartRepo.GetUserCart(c.Request.Context(), userID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, cartItems)
}

func (mc *MarketController) AddToCart(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req models.AddToCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	item, err := mc.cartRepo.AddItem(c.Request.Context(), userID.(int), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	metrics.CartItemsAddedTotal.Inc()

	c.JSON(http.StatusCreated, item)
}

func (mc *MarketController) UpdateCartItem(c *gin.Context) {
	userID, _ := c.Get("user_id")
	itemID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid item ID"})
		return
	}

	var req models.UpdateCartItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	item, err := mc.cartRepo.UpdateItem(c.Request.Context(), itemID, userID.(int), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, item)
}

func (mc *MarketController) DeleteCartItem(c *gin.Context) {
	userID, _ := c.Get("user_id")
	itemID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid item ID"})
		return
	}

	if err := mc.cartRepo.DeleteItem(c.Request.Context(), itemID, userID.(int)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "item removed from cart"})
}

func (mc *MarketController) CreateOrder(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req models.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	order, err := mc.marketService.CreateOrder(c.Request.Context(), userID.(int), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	metrics.OrdersCreatedTotal.Inc()

	c.JSON(http.StatusCreated, order)
}

func (mc *MarketController) GetUserOrders(c *gin.Context) {
	userID, _ := c.Get("user_id")

	orders, err := mc.orderRepo.GetUserOrders(c.Request.Context(), userID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, orders)
}

func (mc *MarketController) GetOrder(c *gin.Context) {
	orderID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order ID"})
		return
	}

	order, err := mc.orderRepo.GetByID(c.Request.Context(), orderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}

	c.JSON(http.StatusOK, order)
}
