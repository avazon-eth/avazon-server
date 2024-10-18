package errs

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// AppError is a common error structure used throughout the application.
type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Error implements error.
func (a AppError) Error() string {
	return a.Message
}

// Setting error messages and status codes
var (
	ErrBadRequest          = AppError{Code: http.StatusBadRequest, Message: "Bad Request"}
	ErrUnauthorized        = AppError{Code: http.StatusUnauthorized, Message: "Unauthorized"}
	ErrForbidden           = AppError{Code: http.StatusForbidden, Message: "Forbidden"}
	ErrNotFound            = AppError{Code: http.StatusNotFound, Message: "Resource Not Found"}
	ErrConflict            = AppError{Code: http.StatusConflict, Message: "Conflict"}
	ErrInternalServerError = AppError{Code: http.StatusInternalServerError, Message: "Internal Server Error"}
	ErrInvalidStatus       = AppError{Code: http.StatusBadRequest, Message: "Invalid Status"}
)

// SendErrorResponse handles common error responses in the Gin context.
func SendErrorResponse(c *gin.Context, appErr AppError) {
	c.JSON(appErr.Code, gin.H{"error": appErr.Message})
	c.Abort()
}
