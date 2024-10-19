package controllers

import (
	"avazon-api/controllers/errs"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

func HandleError(c *gin.Context, err error, message ...string) {
	fmt.Printf("Controller returns error: %v\n", err)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		errs.SendErrorResponse(c, errs.ErrNotFound)
	} else if errors.Is(err, io.EOF) {
		errs.SendErrorResponse(c, errs.ErrBadRequest)
	} else if errors.Is(err, errs.ErrInvalidStatus) {
		errs.SendErrorResponse(c, errs.ErrInvalidStatus)
	} else if appErr, ok := err.(errs.AppError); ok {
		// If the error is of type AppError
		errs.SendErrorResponse(c, appErr)
	} else if validationErr, ok := err.(validator.ValidationErrors); ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
	} else {
		errs.SendErrorResponse(c, errs.ErrInternalServerError)
	}
}

func GetPagingParams(c *gin.Context) (page int, limit int) {
	page, err := strconv.Atoi(c.Query("page"))
	if err != nil {
		page = 0
	}
	limit, err = strconv.Atoi(c.Query("limit"))
	if err != nil {
		limit = 20
	}
	return
}
