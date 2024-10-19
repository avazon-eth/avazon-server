package controllers

import (
	"avazon-api/services"

	"github.com/gin-gonic/gin"
)

type AvatarCreationController struct {
	avatarCreationService *services.AvatarCreateService
}

func NewAvatarCreationController(avatarCreationService *services.AvatarCreateService) *AvatarCreationController {
	return &AvatarCreationController{avatarCreationService: avatarCreationService}
}

// First of all, create a new avatar creation session
// if already there is existing session, close that session and create a new one
func (ctrl *AvatarCreationController) StartCreation(c *gin.Context) {
}

// WebSocket exchange happens here
func (ctrl *AvatarCreationController) EnterSession(c *gin.Context) {
}

// used for tracking how the session is going
func (ctrl *AvatarCreationController) GetOneSession(c *gin.Context) {
}

// receives avatar_id which will be used for avatar (NFT hash value)
func (ctrl *AvatarCreationController) CreateAvatar(c *gin.Context) {
}
