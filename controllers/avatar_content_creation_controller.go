package controllers

import (
	"avazon-api/services"

	"github.com/gin-gonic/gin"
)

type AvatarContentCreationController struct {
	AvatarContentCreationService *services.AvatarContentCreationService
}

func NewAvatarContentCreationController(avatarContentCreationService *services.AvatarContentCreationService) *AvatarContentCreationController {
	return &AvatarContentCreationController{AvatarContentCreationService: avatarContentCreationService}
}

// ========== Music Creation ==========

func (ctrl *AvatarContentCreationController) StartMusicCreation(c *gin.Context) {
}

func (ctrl *AvatarContentCreationController) GetMusicCreations(c *gin.Context) {
}

func (ctrl *AvatarContentCreationController) GetOneMusicCreation(c *gin.Context) {
}

func (ctrl *AvatarContentCreationController) ConfirmAvatarMusic(c *gin.Context) {

}

// ========== Video Creation ==========

func (ctrl *AvatarContentCreationController) StartVideoCreation(c *gin.Context) {

}

func (ctrl *AvatarContentCreationController) GetVideoCreations(c *gin.Context) {

}

func (ctrl *AvatarContentCreationController) GetAllVideoCreations(c *gin.Context) {

}

func (ctrl *AvatarContentCreationController) GetOneVideoCreation(c *gin.Context) {

}

func (ctrl *AvatarContentCreationController) ConfirmAvatarVideo(c *gin.Context) {

}
