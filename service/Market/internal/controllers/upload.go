package controllers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Zifeldev/marketback/service/Market/internal/logger"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	MaxFileSize = 5 * 1024 * 1024 // 5MB
)

var allowedImageExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".webp": true,
}

type UploadController struct {
	uploadDir string
	baseURL   string
}

func NewUploadController(uploadDir, baseURL string) (*UploadController, error) {
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create upload directory: %w", err)
	}
	return &UploadController{
		uploadDir: uploadDir,
		baseURL:   baseURL,
	}, nil
}

// UploadImage godoc
// @Summary Upload product image
// @Description Upload an image file for a product
// @Tags upload
// @Accept multipart/form-data
// @Produce json
// @Param file formance_data file true "Image file"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/upload/image [post]
func (uc *UploadController) UploadImage(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		logger.GetLogger().WithField("err", err).Warn("no file provided")
		c.JSON(http.StatusBadRequest, gin.H{"error": "no file provided"})
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedImageExtensions[ext] {
		logger.GetLogger().WithField("ext", ext).Warn("invalid file type")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file type. Allowed: jpg, jpeg, png, gif, webp"})
		return
	}

	if file.Size > MaxFileSize {
		logger.GetLogger().WithField("size", file.Size).Warn("file too large")
		c.JSON(http.StatusBadRequest, gin.H{"error": "file too large. Max size: 5MB"})
		return
	}

	filename := fmt.Sprintf("%s_%d%s", uuid.New().String(), time.Now().Unix(), ext)

	savePath := filepath.Join(uc.uploadDir, filename)

	if err := c.SaveUploadedFile(file, savePath); err != nil {
		logger.GetLogger().WithField("err", err).Error("failed to save file")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save file"})
		return
	}

	imageURL := fmt.Sprintf("%s/uploads/%s", uc.baseURL, filename)

	c.JSON(http.StatusOK, gin.H{
		"url":      imageURL,
		"filename": filename,
	})
}

// DeleteImage godoc
// @Summary Delete uploaded image
// @Description Delete an uploaded image file
// @Tags upload
// @Produce json
// @Param filename path string true "Filename"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/upload/image/{filename} [delete]
func (uc *UploadController) DeleteImage(c *gin.Context) {
	filename := c.Param("filename")
	if filename == "" {
		logger.GetLogger().Warn("filename required")
		c.JSON(http.StatusBadRequest, gin.H{"error": "filename required"})
		return
	}

	if strings.Contains(filename, "..") || strings.Contains(filename, "/") {
		logger.GetLogger().WithField("filename", filename).Warn("invalid filename")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid filename"})
		return
	}

	filePath := filepath.Join(uc.uploadDir, filename)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		logger.GetLogger().WithField("filename", filename).Warn("file not found")
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	if err := os.Remove(filePath); err != nil {
		logger.GetLogger().WithField("err", err).Error("failed to delete file")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete file"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "file deleted"})
}
