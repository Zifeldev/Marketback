package controllers

import (
	"net/http"
	"strconv"

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

func (sc *SellerController) RegisterSeller(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req models.CreateSellerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	seller, err := sc.sellerRepo.Create(c.Request.Context(), userID.(int), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, seller)
}

func (sc *SellerController) GetSellerProfile(c *gin.Context) {
	userID, _ := c.Get("user_id")

	seller, err := sc.sellerRepo.GetByUserID(c.Request.Context(), userID.(int))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "seller profile not found"})
		return
	}

	c.JSON(http.StatusOK, seller)
}

func (sc *SellerController) UpdateSellerProfile(c *gin.Context) {
	userID, _ := c.Get("user_id")

	seller, err := sc.sellerRepo.GetByUserID(c.Request.Context(), userID.(int))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "seller profile not found"})
		return
	}

	var req models.UpdateSellerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updatedSeller, err := sc.sellerRepo.Update(c.Request.Context(), seller.ID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedSeller)
}

func (sc *SellerController) CreateProduct(c *gin.Context) {
	userID, _ := c.Get("user_id")

	seller, err := sc.sellerRepo.GetByUserID(c.Request.Context(), userID.(int))
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "seller profile not found"})
		return
	}

	var req models.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	product, err := sc.productRepo.Create(c.Request.Context(), seller.ID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, product)
}

func (sc *SellerController) GetSellerProducts(c *gin.Context) {
	userID, _ := c.Get("user_id")

	seller, err := sc.sellerRepo.GetByUserID(c.Request.Context(), userID.(int))
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "seller profile not found"})
		return
	}

	products, err := sc.productRepo.GetBySellerID(c.Request.Context(), seller.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, products)
}

func (sc *SellerController) UpdateProduct(c *gin.Context) {
	userID, _ := c.Get("user_id")
	productID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product ID"})
		return
	}

	seller, err := sc.sellerRepo.GetByUserID(c.Request.Context(), userID.(int))
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "seller profile not found"})
		return
	}

	product, err := sc.productRepo.GetByID(c.Request.Context(), productID)
	if err != nil || product.SellerID != seller.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "product not found or access denied"})
		return
	}

	var req models.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updatedProduct, err := sc.productRepo.Update(c.Request.Context(), productID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedProduct)
}

func (sc *SellerController) DeleteProduct(c *gin.Context) {
	userID, _ := c.Get("user_id")
	productID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product ID"})
		return
	}

	seller, err := sc.sellerRepo.GetByUserID(c.Request.Context(), userID.(int))
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "seller profile not found"})
		return
	}

	product, err := sc.productRepo.GetByID(c.Request.Context(), productID)
	if err != nil || product.SellerID != seller.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "product not found or access denied"})
		return
	}

	if err := sc.productRepo.Delete(c.Request.Context(), productID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "product deleted"})
}
