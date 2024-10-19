package controllers

import (
	"avazon-api/controllers/errs"
	"avazon-api/dto"
	"avazon-api/middleware"
	"avazon-api/services"
	"avazon-api/utils"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	UserService *services.UserService
}

func NewUserController(userService *services.UserService) *UserController {
	return &UserController{UserService: userService}
}

func (ctrl *UserController) OAuth2Login(c *gin.Context) {
	// oauthProvider: google only for now
	oauthProvider := c.Param("provider")
	if oauthProvider == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "OAuth provider is required"})
		return
	}
	// Header: X-OAuth2-Token
	oauthToken := c.GetHeader("X-OAuth2-Token")
	if oauthToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "OAuth token is required"})
		return
	}

	if oauthProvider == "google" {
		user, err := ctrl.UserService.GetUserByGoogleAccessToken(oauthToken)
		if err != nil {
			HandleError(c, err)
			return
		}
		if user == nil {
			HandleError(c, err)
			return
		}
		accessToken, err := middleware.GenerateJWT(user.ID, "access", 60*24*30, string(user.Role)) // 30 days
		if err != nil {
			HandleError(c, err)
			return
		}
		refreshToken, err := middleware.GenerateJWT(user.ID, "refresh", 60*24*7, "refresh") // 7 days
		if err != nil {
			HandleError(c, err)
			return
		}
		tokenDTO := &dto.Token{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		}
		c.JSON(http.StatusOK, tokenDTO)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid OAuth provider. Only Google for now!"})
	}
}

func (ctrl *UserController) GetMyInfo(c *gin.Context) {
	userID, exists := utils.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	user, err := ctrl.UserService.GetUserByID(userID)
	if err != nil {
		log.Printf("Error getting user by ID: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (ctrl *UserController) RefreshToken(c *gin.Context) {
	var tokenDTO dto.Token
	if err := c.ShouldBindJSON(&tokenDTO); err != nil {
		HandleError(c, err)
		return
	}

	accessToken, err := middleware.ValidateJWT(tokenDTO.AccessToken)
	if err != nil || !accessToken.Valid {
		log.Printf("Invalid access token: %v", err)
		HandleError(c, errs.ErrInvalidJWT)
		return
	}

	accessTokenUserID, err := middleware.GetUserIDFromTokenString(tokenDTO.AccessToken)
	if err != nil {
		log.Printf("Invalid access token: %v", err)
		HandleError(c, errs.ErrInvalidJWT)
		return
	}
	refreshToken, err := middleware.ValidateJWT(tokenDTO.RefreshToken)
	if err != nil || !refreshToken.Valid {
		log.Printf("Invalid refresh token: %v", err)
		HandleError(c, errs.ErrInvalidJWT)
		return
	}
	refreshTokenUserID, err := middleware.GetUserIDFromTokenString(tokenDTO.RefreshToken)
	if err != nil {
		log.Printf("Invalid refresh token: %v", err)
		HandleError(c, errs.ErrInvalidJWT)
		return
	}
	if accessTokenUserID != refreshTokenUserID {
		log.Printf("Refresh token mismatch: %v", err)
		HandleError(c, errs.ErrRefreshMismatch)
		return
	}

	userID, ok := utils.GetUserID(c)
	if !ok {
		log.Printf("Cannot get user ID from token: %v", err)
		HandleError(c, errs.ErrInvalidJWT)
		return
	}

	userRole, err := middleware.GetUserRoleFromTokenString(tokenDTO.AccessToken)
	if userRole == nil {
		log.Printf("Cannot get user role from token: %v", err)
		HandleError(c, errs.ErrInvalidJWT)
		return
	}
	if err != nil {
		log.Printf("Cannot get user role from token: %v", err)
		HandleError(c, errs.ErrInvalidJWT)
		return
	}

	accessJWT, err := middleware.GenerateJWT(userID, "access", 30, string(*userRole))
	if err != nil {
		log.Printf("Cannot generate access token: %v", err)
		HandleError(c, errs.ErrInternalServerError)
		return
	}

	refreshJWT, err := middleware.GenerateJWT(userID, "refresh", 60*24*7, "refresh")
	if err != nil {
		log.Printf("Cannot generate refresh token: %v", err)
		HandleError(c, errs.ErrInternalServerError)
		return
	}

	c.JSON(http.StatusOK, dto.Token{
		AccessToken:  accessJWT,
		RefreshToken: refreshJWT,
	})
}
