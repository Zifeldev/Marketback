package controllers

import (
	"net/http"
	"strconv"

	"github.com/Zifeldev/marketback/service/Market/internal/logger"
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
		logger.GetLogger().WithField("err", err).Error("CreateCategory: invalid request body")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	category, err := ac.categoryRepo.Create(c.Request.Context(), &req)
	if err != nil {
		logger.GetLogger().WithField("err", err).Error("CreateCategory: failed to create category")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		logger.GetLogger().WithField("err", err).Error("UpdateCategory: invalid category ID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category ID"})
		return
	}

	var req models.UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.GetLogger().WithField("err", err).Error("UpdateCategory: invalid request body")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	category, err := ac.categoryRepo.Update(c.Request.Context(), id, &req)
	if err != nil {
		logger.GetLogger().WithField("err", err).Error("UpdateCategory: failed to update category")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		logger.GetLogger().WithField("err", err).Error("DeleteCategory: invalid category ID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category ID"})
		return
	}

	if err := ac.categoryRepo.Delete(c.Request.Context(), id); err != nil {
		logger.GetLogger().WithField("err", err).Error("DeleteCategory: failed to delete category")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		logger.GetLogger().WithField("err", err).Error("UpdateProductStatus: invalid product ID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product ID"})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.GetLogger().WithField("err", err).Error("UpdateProductStatus: invalid request body")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updateReq := &models.UpdateProductRequest{
		Status: &req.Status,
	}

	product, err := ac.productRepo.Update(c.Request.Context(), id, updateReq)
	if err != nil {
		logger.GetLogger().WithField("err", err).Error("UpdateProductStatus: failed to update product status")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
	if err != nil {
		logger.GetLogger().WithField("err", err).Error("GetAllSellers: failed to get sellers")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		logger.GetLogger().WithField("err", err).Error("UpdateSellerStatus: invalid seller ID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid seller ID"})
		return
	}

	var req struct {
		IsActive bool `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.GetLogger().WithField("err", err).Error("UpdateSellerStatus: invalid request body")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := ac.sellerRepo.UpdateStatus(c.Request.Context(), id, req.IsActive); err != nil {
		logger.GetLogger().WithField("err", err).Error("UpdateSellerStatus: failed to update seller status")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "seller status updated"})
}

// GetAllOrders godoc
// @Summary Get all orders
// @Description Get list of all orders (admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.Order
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/admin/orders [get]
func (ac *AdminController) GetAllOrders(c *gin.Context) {
	orders, err := ac.orderRepo.GetAll(c.Request.Context())
	if err != nil {
		logger.GetLogger().WithField("err", err).Error("GetAllOrders: failed to get orders")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, orders)
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
		logger.GetLogger().WithField("err", err).Error("UpdateOrderStatus: invalid order ID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order ID"})
		return
	}

	var req models.UpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.GetLogger().WithField("err", err).Error("UpdateOrderStatus: invalid request body")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	order, err := ac.orderRepo.UpdateStatus(c.Request.Context(), id, req.Status)
	if err != nil {
		logger.GetLogger().WithField("err", err).Error("UpdateOrderStatus: failed to update order status")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, order)
}
