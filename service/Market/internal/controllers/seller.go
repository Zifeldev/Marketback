package controllers

import (
	"net/http"
	"strconv"

	"github.com/Zifeldev/marketback/service/Market/internal/apperrors"
	"github.com/Zifeldev/marketback/service/Market/internal/models"
	"github.com/Zifeldev/marketback/service/Market/internal/repository"
	"github.com/gin-gonic/gin"
)

type SellerController struct {
	sellerRepo  *repository.SellerRepository
	productRepo *repository.ProductRepository
}

func NewSellerController(sellerRepo *repository.SellerRepository, productRepo *repository.ProductRepository) *SellerController {
	return &SellerController{
		sellerRepo:  sellerRepo,
		productRepo: productRepo,
	}
}

// RegisterSeller godoc
// @Summary Register seller profile
// @Description Create a seller profile for the authenticated user
// @Tags seller
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.CreateSellerRequest true "Seller data"
// @Success 201 {object} models.Seller
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/seller/register [post]
func (sc *SellerController) RegisterSeller(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req models.CreateSellerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, apperrors.BadRequest(err.Error()))
		return
	}

	seller, err := sc.sellerRepo.Create(c.Request.Context(), userID.(int), &req)
	if handleError(c, err, apperrors.Internal("failed to create seller")) {
		return
	}

	c.JSON(http.StatusCreated, seller)
}

// GetSellerProfile godoc
// @Summary Get seller profile
// @Description Get current user's seller profile
// @Tags seller
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.Seller
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/seller/profile [get]
func (sc *SellerController) GetSellerProfile(c *gin.Context) {
	userID, _ := c.Get("user_id")

	seller, err := sc.sellerRepo.GetByUserID(c.Request.Context(), userID.(int))
	if handleError(c, err, apperrors.NotFound("seller profile not found")) {
		return
	}

	c.JSON(http.StatusOK, seller)
}

// UpdateSellerProfile godoc
// @Summary Update seller profile
// @Description Update current user's seller profile
// @Tags seller
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.UpdateSellerRequest true "Update data"
// @Success 200 {object} models.Seller
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/seller/profile [put]
func (sc *SellerController) UpdateSellerProfile(c *gin.Context) {
	userID, _ := c.Get("user_id")

	seller, err := sc.sellerRepo.GetByUserID(c.Request.Context(), userID.(int))
	if handleError(c, err, apperrors.NotFound("seller profile not found")) {
		return
	}

	var req models.UpdateSellerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, apperrors.BadRequest(err.Error()))
		return
	}

	updatedSeller, err := sc.sellerRepo.Update(c.Request.Context(), seller.ID, &req)
	if handleError(c, err, apperrors.Internal("failed to update seller")) {
		return
	}

	c.JSON(http.StatusOK, updatedSeller)
}

// CreateProduct godoc
// @Summary Create product
// @Description Create a new product for seller
// @Tags seller
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.CreateProductRequest true "Product data"
// @Success 201 {object} models.Product
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/seller/products [post]
func (sc *SellerController) CreateProduct(c *gin.Context) {
	userID, _ := c.Get("user_id")

	seller, err := sc.sellerRepo.GetByUserID(c.Request.Context(), userID.(int))
	if handleError(c, err, apperrors.Forbidden("seller profile not found")) {
		return
	}

	var req models.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, apperrors.BadRequest(err.Error()))
		return
	}

	product, err := sc.productRepo.Create(c.Request.Context(), seller.ID, &req)
	if handleError(c, err, apperrors.Internal("failed to create product")) {
		return
	}

	c.JSON(http.StatusCreated, product)
}

// GetSellerProducts godoc
// @Summary Get seller products
// @Description Get all products for current seller
// @Tags seller
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.Product
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/seller/products [get]
func (sc *SellerController) GetSellerProducts(c *gin.Context) {
	userID, _ := c.Get("user_id")

	seller, err := sc.sellerRepo.GetByUserID(c.Request.Context(), userID.(int))
	if handleError(c, err, apperrors.Forbidden("seller profile not found")) {
		return
	}

	products, err := sc.productRepo.GetBySellerID(c.Request.Context(), seller.ID)
	if handleError(c, err, apperrors.Internal("failed to get products")) {
		return
	}

	c.JSON(http.StatusOK, products)
}

// UpdateProduct godoc
// @Summary Update product
// @Description Update seller's product
// @Tags seller
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Product ID"
// @Param request body models.UpdateProductRequest true "Update data"
// @Success 200 {object} models.Product
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/seller/products/{id} [put]
func (sc *SellerController) UpdateProduct(c *gin.Context) {
	userID, _ := c.Get("user_id")
	productID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondError(c, apperrors.InvalidID("product"))
		return
	}

	seller, err := sc.sellerRepo.GetByUserID(c.Request.Context(), userID.(int))
	if handleError(c, err, apperrors.Forbidden("seller profile not found")) {
		return
	}

	product, err := sc.productRepo.GetByID(c.Request.Context(), productID)
	if err != nil || product.SellerID != seller.ID {
		respondError(c, apperrors.Forbidden("product not found or access denied"))
		return
	}

	var req models.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, apperrors.BadRequest(err.Error()))
		return
	}

	updatedProduct, err := sc.productRepo.Update(c.Request.Context(), productID, &req)
	if handleError(c, err, apperrors.Internal("failed to update product")) {
		return
	}

	c.JSON(http.StatusOK, updatedProduct)
}

// DeleteProduct godoc
// @Summary Delete product
// @Description Delete seller's product
// @Tags seller
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Product ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/seller/products/{id} [delete]
func (sc *SellerController) DeleteProduct(c *gin.Context) {
	userID, _ := c.Get("user_id")
	productID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondError(c, apperrors.InvalidID("product"))
		return
	}

	seller, err := sc.sellerRepo.GetByUserID(c.Request.Context(), userID.(int))
	if handleError(c, err, apperrors.Forbidden("seller profile not found")) {
		return
	}

	product, err := sc.productRepo.GetByID(c.Request.Context(), productID)
	if err != nil || product.SellerID != seller.ID {
		respondError(c, apperrors.Forbidden("product not found or access denied"))
		return
	}

	if err := sc.productRepo.Delete(c.Request.Context(), productID); err != nil {
		handleError(c, err, apperrors.Internal("failed to delete product"))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "product deleted"})
}
