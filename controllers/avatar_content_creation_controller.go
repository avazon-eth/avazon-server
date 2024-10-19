package controllers

import (
	"avazon-api/controllers/errs"
	"avazon-api/dto"
	"avazon-api/services"
	"avazon-api/utils"
	"net/http"

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
	userID, ok := utils.GetUserID(c)
	if !ok {
		HandleError(c, errs.ErrUnauthorized)
		return
	}

	var request dto.AvatarMusicRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		HandleError(c, errs.ErrBadRequest)
		return
	}

	musicCreation, err := ctrl.AvatarContentCreationService.CreateAvatarMusicImage(userID, c.Param("avatar_id"), request)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, musicCreation)
}

// ?avatar_id(optional)&page&limit
func (ctrl *AvatarContentCreationController) GetMusicCreations(c *gin.Context) {
	userID, ok := utils.GetUserID(c)
	if !ok {
		HandleError(c, errs.ErrUnauthorized)
		return
	}
	page, limit := GetPagingParams(c)
	avatarID := c.Param("avatar_id")
	var avatarIDPtr *string
	if avatarID != "" {
		avatarIDPtr = &avatarID
	}

	musicCreations, err := ctrl.AvatarContentCreationService.GetAvatarMusicCreations(userID, avatarIDPtr, page, limit)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, musicCreations)
}

func (ctrl *AvatarContentCreationController) GetOneMusicCreation(c *gin.Context) {
	userID, ok := utils.GetUserID(c)
	if !ok {
		HandleError(c, errs.ErrUnauthorized)
		return
	}

	musicCreation, err := ctrl.AvatarContentCreationService.GetAvatarMusicCreation(userID, c.Param("creation_id"))
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, musicCreation)
}

func (ctrl *AvatarContentCreationController) ConfirmAvatarMusic(c *gin.Context) {
	userID, ok := utils.GetUserID(c)
	if !ok {
		HandleError(c, errs.ErrUnauthorized)
		return
	}
	creationID := c.Param("creation_id")
	if creationID == "" {
		HandleError(c, errs.ErrBadRequest, "creation_id is required")
		return
	}

	contentID := c.Query("content_id")
	if contentID == "" {
		HandleError(c, errs.ErrBadRequest, "content_id is required")
		return
	}

	music, err := ctrl.AvatarContentCreationService.ConfirmAvatarMusic(userID, creationID, contentID)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, music)
}

// ========== Video Creation ==========

func (ctrl *AvatarContentCreationController) StartVideoImageCreation(c *gin.Context) {
	userID, ok := utils.GetUserID(c)
	if !ok {
		HandleError(c, errs.ErrUnauthorized)
		return
	}

	var request dto.AvatarVideoImageRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		HandleError(c, errs.ErrBadRequest)
		return
	}

	videoCreation, err := ctrl.AvatarContentCreationService.CreateAvatarVideoImage(userID, c.Param("avatar_id"), request)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, videoCreation)
}

func (ctrl *AvatarContentCreationController) GetVideoCreation(c *gin.Context) {
	userID, ok := utils.GetUserID(c)
	if !ok {
		HandleError(c, errs.ErrUnauthorized)
		return
	}
	page, limit := GetPagingParams(c)
	avatarID := c.Param("avatar_id")
	var avatarIDPtr *string
	if avatarID != "" {
		avatarIDPtr = &avatarID
	}

	videoCreations, err := ctrl.AvatarContentCreationService.GetAvatarVideoCreations(userID, avatarIDPtr, page, limit)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, videoCreations)
}

func (ctrl *AvatarContentCreationController) GetAllVideoCreations(c *gin.Context) {
	userID, ok := utils.GetUserID(c)
	if !ok {
		HandleError(c, errs.ErrUnauthorized)
		return
	}

	videoCreation, err := ctrl.AvatarContentCreationService.GetAvatarVideoCreation(userID, c.Param("creation_id"))
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, videoCreation)
}

func (ctrl *AvatarContentCreationController) StartVideoCreationFromImage(c *gin.Context) {
	userID, ok := utils.GetUserID(c)
	if !ok {
		HandleError(c, errs.ErrUnauthorized)
		return
	}

	avatarID := c.Param("avatar_id")
	if avatarID == "" {
		HandleError(c, errs.ErrBadRequest, "avatar_id is required")
		return
	}
	creationID := c.Param("creation_id")
	if creationID == "" {
		HandleError(c, errs.ErrBadRequest, "creation_id is required")
		return
	}

	var request dto.AvatarVideoRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		HandleError(c, errs.ErrBadRequest)
		return
	}

	videoCreation, err := ctrl.AvatarContentCreationService.CreateAvatarVideoFromImage(userID, avatarID, creationID, request)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, videoCreation)
}

func (ctrl *AvatarContentCreationController) GetOneVideoCreation(c *gin.Context) {
	userID, ok := utils.GetUserID(c)
	if !ok {
		HandleError(c, errs.ErrUnauthorized)
		return
	}

	videoCreation, err := ctrl.AvatarContentCreationService.GetAvatarVideoCreation(userID, c.Param("creation_id"))
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, videoCreation)
}

func (ctrl *AvatarContentCreationController) ConfirmAvatarVideo(c *gin.Context) {
	userID, ok := utils.GetUserID(c)
	if !ok {
		HandleError(c, errs.ErrUnauthorized)
		return
	}
	creationID := c.Param("creation_id")
	if creationID == "" {
		HandleError(c, errs.ErrBadRequest, "creation_id is required")
		return
	}

	contentID := c.Query("content_id")
	if contentID == "" {
		HandleError(c, errs.ErrBadRequest, "content_id is required")
		return
	}

	video, err := ctrl.AvatarContentCreationService.ConfirmAvatarVideo(userID, creationID, contentID)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, video)
}
