package controllers

import (
	"net/http"
	"strconv"

	"github.com/Zifeldev/marketback/service/Market/internal/apperrors"
	"github.com/Zifeldev/marketback/service/Market/internal/models"
	"github.com/Zifeldev/marketback/service/Market/internal/repository"
	"github.com/gin-gonic/gin"
)

type AdminController struct {
	categoryRepo *repository.CategoryRepository
	productRepo  *repository.ProductRepository
	sellerRepo   *repository.SellerRepository
	orderRepo    *repository.OrderRepository
}

func NewAdminController(
	categoryRepo *repository.CategoryRepository,
	productRepo *repository.ProductRepository,
	sellerRepo *repository.SellerRepository,
	orderRepo *repository.OrderRepository,
) *AdminController {
	return &AdminController{
		categoryRepo: categoryRepo,
		productRepo:  productRepo,
		sellerRepo:   sellerRepo,
		orderRepo:    orderRepo,
	}
}

// CreateCategory godoc
// @Summary Create category
// @Description Create a new product category (admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.CreateCategoryRequest true "Category data"
// @Success 201 {object} models.Category
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/admin/categories [post]
func (ac *AdminController) CreateCategory(c *gin.Context) {
	var req models.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, apperrors.BadRequest(err.Error()))
		return
	}

	category, err := ac.categoryRepo.Create(c.Request.Context(), &req)
	if handleError(c, err, apperrors.Internal("failed to create category")) {
		return
	}

	c.JSON(http.StatusCreated, category)
}

// UpdateCategory godoc
// @Summary Update category
// @Description Update an existing category (admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Category ID"
// @Param request body models.UpdateCategoryRequest true "Update data"
// @Success 200 {object} models.Category
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/admin/categories/{id} [put]
func (ac *AdminController) UpdateCategory(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondError(c, apperrors.InvalidID("category"))
		return
	}

	var req models.UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, apperrors.BadRequest(err.Error()))
		return
	}

	category, err := ac.categoryRepo.Update(c.Request.Context(), id, &req)
	if handleError(c, err, apperrors.Internal("failed to update category")) {
		return
	}

	c.JSON(http.StatusOK, category)
}

// DeleteCategory godoc
// @Summary Delete category
// @Description Delete a category (admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Category ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/admin/categories/{id} [delete]
func (ac *AdminController) DeleteCategory(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondError(c, apperrors.InvalidID("category"))
		return
	}

	if err := ac.categoryRepo.Delete(c.Request.Context(), id); err != nil {
		handleError(c, err, apperrors.Internal("failed to delete category"))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "category deleted"})
}

// UpdateProductStatus godoc
// @Summary Update product status
// @Description Update product status (admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Product ID"
// @Param request body object{status=string} true "Status data"
// @Success 200 {object} models.Product
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/admin/products/{id}/status [put]
func (ac *AdminController) UpdateProductStatus(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondError(c, apperrors.InvalidID("product"))
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, apperrors.BadRequest(err.Error()))
		return
	}

	updateReq := &models.UpdateProductRequest{
		Status: &req.Status,
	}

	product, err := ac.productRepo.Update(c.Request.Context(), id, updateReq)
	if handleError(c, err, apperrors.Internal("failed to update product status")) {
		return
	}

	c.JSON(http.StatusOK, product)
}

// GetAllSellers godoc
// @Summary Get all sellers
// @Description Get list of all sellers (admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.Seller
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/admin/sellers [get]
func (ac *AdminController) GetAllSellers(c *gin.Context) {
	sellers, err := ac.sellerRepo.GetAll(c.Request.Context())
	if handleError(c, err, apperrors.Internal("failed to get sellers")) {
		return
	}

	c.JSON(http.StatusOK, sellers)
}

// UpdateSellerStatus godoc
// @Summary Update seller status
// @Description Activate or deactivate a seller (admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Seller ID"
// @Param request body object{is_active=bool} true "Status data"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/admin/sellers/{id}/status [put]
func (ac *AdminController) UpdateSellerStatus(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondError(c, apperrors.InvalidID("seller"))
		return
	}

	var req struct {
		IsActive bool `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, apperrors.BadRequest(err.Error()))
		return
	}

	if err := ac.sellerRepo.UpdateStatus(c.Request.Context(), id, req.IsActive); err != nil {
		handleError(c, err, apperrors.Internal("failed to update seller status"))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "seller status updated"})
}

// GetAllOrders godoc
// @Summary Get all orders
// @Description Get list of all orders with pagination (admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Param status query string false "Filter by status"
// @Success 200 {object} models.PaginatedResponse
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/admin/orders [get]
func (ac *AdminController) GetAllOrders(c *gin.Context) {
	var pagination models.PaginationParams
	if err := c.ShouldBindQuery(&pagination); err != nil {
		pagination = models.PaginationParams{Page: 1, PageSize: models.DefaultPageSize}
	}
	if pagination.Page < 1 {
		pagination.Page = 1
	}

	status := c.Query("status")

	orders, totalItems, err := ac.orderRepo.GetAll(c.Request.Context(), &pagination, status)
	if handleError(c, err, apperrors.Internal("failed to get orders")) {
		return
	}

	response := models.PaginatedResponse{
		Data:       orders,
		Pagination: models.NewPaginationMeta(pagination.Page, pagination.GetLimit(), totalItems),
	}

	c.JSON(http.StatusOK, response)
}

// UpdateOrderStatus godoc
// @Summary Update order status
// @Description Update status of an order (admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Order ID"
// @Param request body models.UpdateOrderStatusRequest true "Status data"
// @Success 200 {object} models.Order
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/admin/orders/{id}/status [put]
func (ac *AdminController) UpdateOrderStatus(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondError(c, apperrors.InvalidID("order"))
		return
	}

	var req models.UpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, apperrors.BadRequest(err.Error()))
		return
	}

	order, err := ac.orderRepo.UpdateStatus(c.Request.Context(), id, req.Status)
	if handleError(c, err, apperrors.Internal("failed to update order status")) {
		return
	}

	c.JSON(http.StatusOK, order)
}
