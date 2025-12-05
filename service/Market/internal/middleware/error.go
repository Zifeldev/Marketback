package middleware

import (
	"net/http"

	"github.com/Zifeldev/marketback/service/Market/internal/apperrors"
	"github.com/gin-gonic/gin"
)

// ErrorResponse represents API error response
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ErrorHandler is a middleware that handles errors and maps them to HTTP responses
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err

			if appErr := apperrors.GetAppError(err); appErr != nil {
				c.JSON(appErr.HTTPStatus, ErrorResponse{
					Code:    appErr.Code,
					Message: appErr.Message,
				})
				return
			}

			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Code:    apperrors.CodeInternalError,
				Message: "internal server error",
			})
		}
	}
}

// HandleError handles any error and responds appropriately
func HandleError(c *gin.Context, err error) {
	if appErr := apperrors.GetAppError(err); appErr != nil {
		c.JSON(appErr.HTTPStatus, ErrorResponse{
			Code:    appErr.Code,
			Message: appErr.Message,
		})
		return
	}

	c.JSON(http.StatusInternalServerError, ErrorResponse{
		Code:    apperrors.CodeInternalError,
		Message: err.Error(),
	})
}
