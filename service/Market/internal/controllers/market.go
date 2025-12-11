package controllers

import (
	"net/http"
	"strconv"

	"github.com/Zifeldev/marketback/service/Market/internal/apperrors"
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

// GetProducts godoc
// @Summary Get all products
// @Description Get paginated list of products with optional filters
// @Tags products
// @Accept json
// @Produce json
// @Param category_id query int false "Filter by category ID"
// @Param seller_id query int false "Filter by seller ID"
// @Param status query string false "Filter by status"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} models.PaginatedResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/products [get]
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
		respondError(c, apperrors.BadRequest("invalid pagination parameters"))
		return
	}

	products, totalItems, err := mc.productRepo.GetAll(c.Request.Context(), categoryID, sellerID, status, &pagination)
	if handleError(c, err, apperrors.Internal("failed to get products")) {
		return
	}

	response := models.PaginatedResponse{
		Data:       products,
		Pagination: models.NewPaginationMeta(pagination.Page, pagination.GetLimit(), totalItems),
	}

	c.JSON(http.StatusOK, response)
}

// GetProduct godoc
// @Summary Get product by ID
// @Description Get detailed product information
// @Tags products
// @Accept json
// @Produce json
// @Param id path int true "Product ID"
// @Success 200 {object} models.ProductWithDetails
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/products/{id} [get]
func (mc *MarketController) GetProduct(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondError(c, apperrors.InvalidID("product"))
		return
	}

	product, err := mc.productRepo.GetByID(c.Request.Context(), id)
	if handleError(c, err, apperrors.ProductNotFound(id)) {
		return
	}

	metrics.ProductsViewedTotal.Inc()

	c.JSON(http.StatusOK, product)
}

// GetCategories godoc
// @Summary Get all categories
// @Description Get list of all product categories
// @Tags categories
// @Accept json
// @Produce json
// @Success 200 {array} models.Category
// @Failure 500 {object} map[string]string
// @Router /api/categories [get]
func (mc *MarketController) GetCategories(c *gin.Context) {
	categories, err := mc.categoryRepo.GetAll(c.Request.Context())
	if handleError(c, err, apperrors.Internal("failed to get categories")) {
		return
	}

	c.JSON(http.StatusOK, categories)
}

// GetCategory godoc
// @Summary Get category by ID
// @Description Get category details by ID
// @Tags categories
// @Accept json
// @Produce json
// @Param id path int true "Category ID"
// @Success 200 {object} models.Category
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/categories/{id} [get]
func (mc *MarketController) GetCategory(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondError(c, apperrors.InvalidID("category"))
		return
	}

	category, err := mc.categoryRepo.GetByID(c.Request.Context(), id)
	if handleError(c, err, apperrors.CategoryNotFound(id)) {
		return
	}

	c.JSON(http.StatusOK, category)
}

// GetCart godoc
// @Summary Get user cart
// @Description Get current user's cart items
// @Tags cart
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.CartItemWithDetails
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/cart [get]
func (mc *MarketController) GetCart(c *gin.Context) {
	userID, _ := c.Get("user_id")

	cartItems, err := mc.cartRepo.GetUserCart(c.Request.Context(), userID.(int))
	if handleError(c, err, apperrors.Internal("failed to get cart")) {
		return
	}

	c.JSON(http.StatusOK, cartItems)
}

// AddToCart godoc
// @Summary Add item to cart
// @Description Add a product to user's cart
// @Tags cart
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.AddToCartRequest true "Cart item data"
// @Success 201 {object} models.CartItem
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/cart/items [post]
func (mc *MarketController) AddToCart(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req models.AddToCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, apperrors.BadRequest(err.Error()))
		return
	}

	item, err := mc.cartRepo.AddItem(c.Request.Context(), userID.(int), &req)
	if handleError(c, err, apperrors.Internal("failed to add item to cart")) {
		return
	}

	metrics.CartItemsAddedTotal.Inc()

	c.JSON(http.StatusCreated, item)
}

// UpdateCartItem godoc
// @Summary Update cart item
// @Description Update quantity of a cart item
// @Tags cart
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Cart item ID"
// @Param request body models.UpdateCartItemRequest true "Update data"
// @Success 200 {object} models.CartItem
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/cart/items/{id} [put]
func (mc *MarketController) UpdateCartItem(c *gin.Context) {
	userID, _ := c.Get("user_id")
	itemID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondError(c, apperrors.InvalidID("cart item"))
		return
	}

	var req models.UpdateCartItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, apperrors.BadRequest(err.Error()))
		return
	}

	item, err := mc.cartRepo.UpdateItem(c.Request.Context(), itemID, userID.(int), &req)
	if handleError(c, err, apperrors.Internal("failed to update cart item")) {
		return
	}

	c.JSON(http.StatusOK, item)
}

// DeleteCartItem godoc
// @Summary Remove item from cart
// @Description Delete a cart item
// @Tags cart
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Cart item ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/cart/items/{id} [delete]
func (mc *MarketController) DeleteCartItem(c *gin.Context) {
	userID, _ := c.Get("user_id")
	itemID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondError(c, apperrors.InvalidID("cart item"))
		return
	}

	if err := mc.cartRepo.DeleteItem(c.Request.Context(), itemID, userID.(int)); err != nil {
		handleError(c, err, apperrors.Internal("failed to delete cart item"))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "item removed from cart"})
}

// CreateOrder godoc
// @Summary Create order
// @Description Create a new order from cart items
// @Tags orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.CreateOrderRequest true "Order data"
// @Success 201 {object} models.OrderWithItems
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/user/orders [post]
func (mc *MarketController) CreateOrder(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req models.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, apperrors.BadRequest(err.Error()))
		return
	}

	order, err := mc.marketService.CreateOrder(c.Request.Context(), userID.(int), &req)
	if handleError(c, err, apperrors.Internal("failed to create order")) {
		return
	}

	metrics.OrdersCreatedTotal.Inc()

	c.JSON(http.StatusCreated, order)
}

// GetUserOrders godoc
// @Summary Get user orders
// @Description Get all orders for current user with pagination
// @Tags orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} models.PaginatedResponse
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/user/orders [get]
func (mc *MarketController) GetUserOrders(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var pagination models.PaginationParams
	if err := c.ShouldBindQuery(&pagination); err != nil {
		pagination = models.PaginationParams{Page: 1, PageSize: models.DefaultPageSize}
	}
	if pagination.Page < 1 {
		pagination.Page = 1
	}

	orders, totalItems, err := mc.orderRepo.GetUserOrders(c.Request.Context(), userID.(int), &pagination)
	if handleError(c, err, apperrors.Internal("failed to get orders")) {
		return
	}

	response := models.PaginatedResponse{
		Data:       orders,
		Pagination: models.NewPaginationMeta(pagination.Page, pagination.GetLimit(), totalItems),
	}

	c.JSON(http.StatusOK, response)
}

// GetOrder godoc
// @Summary Get order by ID
// @Description Get detailed order information
// @Tags orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Order ID"
// @Success 200 {object} models.OrderWithItems
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/user/orders/{id} [get]
func (mc *MarketController) GetOrder(c *gin.Context) {
	orderID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondError(c, apperrors.InvalidID("order"))
		return
	}

	order, err := mc.orderRepo.GetByID(c.Request.Context(), orderID)
	if handleError(c, err, apperrors.OrderNotFound(orderID)) {
		return
	}

	c.JSON(http.StatusOK, order)
}
