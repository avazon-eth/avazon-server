package controllers

import (
	"avazon-api/controllers/errs"
	"avazon-api/dto"
	"avazon-api/services"
	"avazon-api/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AvatarController struct {
	AvatarService *services.AvatarService
}

func NewAvatarController(avatarService *services.AvatarService) *AvatarController {
	return &AvatarController{AvatarService: avatarService}
}

func (ctrl *AvatarController) GetAvatars(c *gin.Context) {
	page, limit := GetPagingParams(c)

	avatars, err := ctrl.AvatarService.GetAvatars(page, limit)
	if err != nil {
		HandleError(c, err)
		return
	}

	total, err := ctrl.AvatarService.GetAvatarsCount()
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.PageResponse{
		Items: avatars,
		Total: total,
	})
}

func (ctrl *AvatarController) GetMyAvatars(c *gin.Context) {
	userID, ok := utils.GetUserID(c)
	if !ok {
		HandleError(c, errs.ErrInvalidJWT)
		return
	}

	page, limit := GetPagingParams(c)

	avatars, err := ctrl.AvatarService.GetMyAvatars(userID, page, limit)
	if err != nil {
		HandleError(c, err)
		return
	}

	total, err := ctrl.AvatarService.GetMyAvatarsCount(userID)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.PageResponse{
		Items: avatars,
		Total: total,
	})
}

func (ctrl *AvatarController) GetOneAvatar(c *gin.Context) {
	avatarID := c.Param("avatar_id")

	avatar, err := ctrl.AvatarService.GetOneAvatar(avatarID)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, avatar)
}

func (ctrl *AvatarController) GetAvatarContents(c *gin.Context) {
	contentType := c.Param("content_type")
	page, limit := GetPagingParams(c)
	avatarID := c.Query("avatar_id")

	var avatarIDPtr *string
	if avatarID != "" {
		avatarIDPtr = &avatarID
	}

	switch contentType {
	case "music":
		avatars, err := ctrl.AvatarService.GetAvatarMusicContents(avatarIDPtr, page, limit)
		if err != nil {
			HandleError(c, err)
			return
		}
		total, err := ctrl.AvatarService.GetAvatarMusicContentsCount(avatarIDPtr)
		if err != nil {
			HandleError(c, err)
			return
		}
		c.JSON(http.StatusOK, dto.PageResponse{
			Items: avatars,
			Total: total,
		})

	case "video":
		avatars, err := ctrl.AvatarService.GetAvatarVideoContents(avatarIDPtr, page, limit)
		if err != nil {
			HandleError(c, err)
			return
		}
		total, err := ctrl.AvatarService.GetAvatarVideoContentsCount(avatarIDPtr)
		if err != nil {
			HandleError(c, err)
			return
		}
		c.JSON(http.StatusOK, dto.PageResponse{
			Items: avatars,
			Total: total,
		})
	}
}

func (ctrl *AvatarController) GetOneAvatarContent(c *gin.Context) {
	contentType := c.Param("content_type")
	contentID := c.Param("content_id")

	switch contentType {
	case "music":
		music, err := ctrl.AvatarService.GetOneAvatarMusicContent(contentID)
		if err != nil {
			HandleError(c, err)
			return
		}
		c.JSON(http.StatusOK, music)

	case "video":
		video, err := ctrl.AvatarService.GetOneAvatarVideoContent(contentID)
		if err != nil {
			HandleError(c, err)
			return
		}
		c.JSON(http.StatusOK, video)
	default:
		HandleError(c, errs.ErrBadRequest, "music, video are only supported")
	}
}

func (ctrl *AvatarController) GetMyAvatarContents(c *gin.Context) {
	userID, ok := utils.GetUserID(c)
	if !ok {
		HandleError(c, errs.ErrInvalidJWT)
		return
	}

	contentType := c.Param("content_type")
	page, limit := GetPagingParams(c)

	switch contentType {
	case "music":
		avatars, err := ctrl.AvatarService.GetMyAvatarMusicContents(userID, page, limit)
		if err != nil {
			HandleError(c, err)
			return
		}
		total, err := ctrl.AvatarService.GetMyAvatarMusicContentsCount(userID)
		if err != nil {
			HandleError(c, err)
			return
		}
		c.JSON(http.StatusOK, dto.PageResponse{
			Items: avatars,
			Total: total,
		})

	case "video":
		avatars, err := ctrl.AvatarService.GetMyAvatarVideoContents(userID, page, limit)
		if err != nil {
			HandleError(c, err)
			return
		}
		total, err := ctrl.AvatarService.GetMyAvatarVideoContentsCount(userID)
		if err != nil {
			HandleError(c, err)
			return
		}
		c.JSON(http.StatusOK, dto.PageResponse{
			Items: avatars,
			Total: total,
		})
	}
}
