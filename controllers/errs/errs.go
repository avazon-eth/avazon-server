package errs

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// AppError is a common error structure used throughout the application.
type AppError struct {
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
	ErrorCode  string `json:"error_code"`
}

// Error implements error.
func (a AppError) Error() string {
	return a.Message
}

// Setting error messages and status codes
var (
	ErrBadRequest          = AppError{StatusCode: http.StatusBadRequest, Message: "Bad Request", ErrorCode: "40000"}
	ErrUnauthorized        = AppError{StatusCode: http.StatusUnauthorized, Message: "Unauthorized", ErrorCode: "40100"}
	ErrForbidden           = AppError{StatusCode: http.StatusForbidden, Message: "Forbidden", ErrorCode: "40300"}
	ErrNotFound            = AppError{StatusCode: http.StatusNotFound, Message: "Resource Not Found", ErrorCode: "40400"}
	ErrConflict            = AppError{StatusCode: http.StatusConflict, Message: "Conflict", ErrorCode: "40900"}
	ErrInternalServerError = AppError{StatusCode: http.StatusInternalServerError, Message: "Internal Server Error", ErrorCode: "50000"}
	ErrInvalidStatus       = AppError{StatusCode: http.StatusBadRequest, Message: "Invalid Status", ErrorCode: "40001"}
	ErrJWTExpired          = AppError{StatusCode: http.StatusUnauthorized, Message: "JWT Expired", ErrorCode: "40101"}
	ErrInvalidJWT          = AppError{StatusCode: http.StatusUnauthorized, Message: "Invalid JWT", ErrorCode: "40102"}
	ErrRefreshMismatch     = AppError{StatusCode: http.StatusUnauthorized, Message: "Refresh Token Mismatch between access token and refresh token", ErrorCode: "40103"}
)

// SendErrorResponse handles common error responses in the Gin context.
func SendErrorResponse(c *gin.Context, appErr AppError) {
	c.JSON(appErr.StatusCode, appErr)
	c.Abort()
}
