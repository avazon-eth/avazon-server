package controllers

import (
	"avazon-api/services"

	"github.com/gin-gonic/gin"
)

type WebSessionController struct {
	service *services.WebDataSessionService
}

func NewWebSessionController(service *services.WebDataSessionService) *WebSessionController {
	return &WebSessionController{service: service}
}

func (ctrl *WebSessionController) EnterToSession(c *gin.Context) {

}
