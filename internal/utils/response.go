package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func SuccessResponse(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    data,
		Message: message,
	})
}

func ErrorResponse(c *gin.Context, message string, err error) {
	c.JSON(http.StatusBadRequest, Response{
		Success: false,
		Error:   err.Error(),
		Message: message,
	})
}

func ValidationErrorResponse(c *gin.Context, message string, validationErrors map[string]string) {
	c.JSON(http.StatusBadRequest, Response{
		Success: false,
		Error:   "Validation error",
		Message: message,
		Data:    validationErrors,
	})
}
