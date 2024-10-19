package controllers

import (
	"avazon-api/dto"
	"avazon-api/services"
	"avazon-api/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AvatarRemixController struct {
	AvatarRemixService *services.AvatarRemixService
}

func NewAvatarRemixController(avatarRemixService *services.AvatarRemixService) *AvatarRemixController {
	return &AvatarRemixController{AvatarRemixService: avatarRemixService}
}

func (ctrl *AvatarRemixController) StartImageRemix(c *gin.Context) {
	avatarID := c.Param("avatar_id")
	if avatarID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Avatar ID is required"})
		return
	}

	userID, ok := utils.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req dto.AvatarImageRemixRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Prompt == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Prompt is required"})
		return
	}

	remix, err := ctrl.AvatarRemixService.StartImageRemix(userID, avatarID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, remix)
}

func (ctrl *AvatarRemixController) GetOneImageRemix(c *gin.Context) {
	userID, ok := utils.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	avatarID := c.Param("avatar_id")
	if avatarID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Avatar ID is required"})
		return
	}

	remixID := c.Param("remix_id")
	if remixID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Remix ID is required"})
		return
	}

	remix, err := ctrl.AvatarRemixService.GetOneImageRemix(userID, avatarID, remixID)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, remix)
}

func (ctrl *AvatarRemixController) ConfirmImageRemix(c *gin.Context) {
	userID, ok := utils.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	avatarID := c.Param("avatar_id")
	if avatarID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Avatar ID is required"})
		return
	}

	remixID := c.Param("remix_id")
	if remixID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Remix ID is required"})
		return
	}

	newAvatarID := c.Query("avatar_id")
	if newAvatarID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "New Avatar ID is required"})
		return
	}

	remix, err := ctrl.AvatarRemixService.ConfirmAvatarFromImageRemix(userID, avatarID, remixID, newAvatarID)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, remix)
}
