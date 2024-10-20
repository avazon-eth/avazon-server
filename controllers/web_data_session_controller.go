package controllers

import (
	"avazon-api/services"
	"avazon-api/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type WebDataSessionController struct {
	service *services.WebDataSessionService
}

func NewWebDataSessionController(service *services.WebDataSessionService) *WebDataSessionController {
	return &WebDataSessionController{service: service}
}

type PutDataRequest struct {
	SessionID string `json:"session_id"`
	Data      string `json:"data"`
}

type PutTokenRequest struct {
	Token string `json:"token"`
}

type SessionIDRequest struct {
	SessionID string `json:"session_id"`
}

type TokenKeyRequest struct {
	TokenKey string `json:"token_key"`
}

func (ctrl *WebDataSessionController) PutData(c *gin.Context) {
	var req PutDataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID, ok := utils.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}

	err := ctrl.service.PutData(userID, req.SessionID, req.Data)
	if err != nil {
		HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "data put"})
}

func (ctrl *WebDataSessionController) GetData(c *gin.Context) {
	var req SessionIDRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, ok := utils.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}

	data, err := ctrl.service.GetData(req.SessionID, userID)
	if err != nil {
		HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": data})
}

func (ctrl *WebDataSessionController) ClearData(c *gin.Context) {
	var req SessionIDRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, ok := utils.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}

	err := ctrl.service.ClearData(req.SessionID, userID)
	if err != nil {
		HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "data cleared"})
}

func (ctrl *WebDataSessionController) PutToken(c *gin.Context) {
	var req PutTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, ok := utils.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}

	session, err := ctrl.service.PutToken(userID, req.Token)
	if err != nil {
		HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, session)
}

func (ctrl *WebDataSessionController) GetToken(c *gin.Context) {
	var req TokenKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := ctrl.service.GetToken(req.TokenKey)
	if err != nil {
		HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, token)
}

func (ctrl *WebDataSessionController) ClearToken(c *gin.Context) {
	var req SessionIDRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := ctrl.service.ClearToken(req.SessionID)
	if err != nil {
		HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "token cleared"})
}
