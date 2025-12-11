package controllers

import (
	"github.com/Zifeldev/marketback/service/Market/internal/apperrors"
	"github.com/Zifeldev/marketback/service/Market/internal/logger"
	"github.com/gin-gonic/gin"
)

// ErrorResponse represents the standard error response structure
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// respondError responds with an AppError
func respondError(c *gin.Context, err *apperrors.AppError) {
	logger.GetLogger().WithFields(map[string]interface{}{
		"code":    err.Code,
		"message": err.Message,
		"path":    c.Request.URL.Path,
	}).Warn("Request failed")

	c.JSON(err.HTTPStatus, ErrorResponse{
		Code:    err.Code,
		Message: err.Message,
	})
}

// handleError checks error type and responds accordingly
// Returns true if error was handled
func handleError(c *gin.Context, err error, fallbackErr *apperrors.AppError) bool {
	if err == nil {
		return false
	}

	// Check if it's already an AppError
	if appErr := apperrors.GetAppError(err); appErr != nil {
		respondError(c, appErr)
		return true
	}

	// Use fallback error with original error message for logging
	logger.GetLogger().WithField("err", err).Error(fallbackErr.Message)
	respondError(c, fallbackErr)
	return true
}
