package controllers

import (
	"avazon-api/controllers/errs"
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

// func (ctrl *UserController) OAuth2Login(c *gin.Context) {
// 	// oauthProvider: google only for now
// 	oauthProvider := c.Param("provider")
// 	if oauthProvider == "" {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "OAuth provider is required"})
// 		return
// 	}
// 	// Header: X-OAuth2-Auth
// 	// oauthCode := c.GetHeader("X-OAuth2-Auth")
// 	// if oauthCode == "" {
// 	// 	c.JSON(http.StatusBadRequest, gin.H{"error": "OAuth code is required"})
// 	// 	return
// 	// }

// 	if oauthProvider == "google" {
// 		googleCode := c.Query("code")
// 		if googleCode == "" {
// 			HandleError(c, errs.ErrBadRequest, "Google authorization code is required")
// 			return
// 		}
// 		googleAccessToken, err := ctrl.UserService.GetUserGoogleAccessTokenByAuthorizationCode(googleCode)
// 		if err != nil {
// 			HandleError(c, err)
// 			return
// 		}
// 		user, err := ctrl.UserService.GetUserByGoogleAccessToken(googleAccessToken)
// 		if err != nil {
// 			HandleError(c, err)
// 			return
// 		}
// 		accessToken, err := middleware.GenerateJWT(user.ID, "access", 60*24*30, string(user.Role)) // 30 days
// 		if err != nil {
// 			HandleError(c, err)
// 			return
// 		}
// 		refreshToken, err := middleware.GenerateJWT(user.ID, "refresh", 60*24*7, "refresh") // 7 days
// 		if err != nil {
// 			HandleError(c, err)
// 			return
// 		}
// 		tokenDTO := &dto.Token{
// 			AccessToken:  accessToken,
// 			RefreshToken: refreshToken,
// 		}
// 		c.JSON(http.StatusOK, tokenDTO)
// 	} else {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid OAuth provider. Only Google for now!"})
// 	}
// }

func (ctrl *UserController) GetMyInfo(c *gin.Context) {
	userID, exists := utils.GetUserID(c)
	if !exists {
		HandleError(c, errs.ErrInvalidJWT)
		return
	}

	user, err := ctrl.UserService.GetUserByIDCreateIfNotExists(userID)
	if err != nil {
		log.Printf("Error getting user by ID: %v", err)
		HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, user)
}

// func (ctrl *UserController) RefreshToken(c *gin.Context) {
// 	var tokenDTO dto.Token
// 	if err := c.ShouldBindJSON(&tokenDTO); err != nil {
// 		HandleError(c, err)
// 		return
// 	}

// 	accessTokenPayload, err := middleware.ExtractPayload(tokenDTO.AccessToken)
// 	accessTokenType, ok := accessTokenPayload["type"].(string)
// 	if !ok || accessTokenType != "access" {
// 		log.Printf("Invalid access token: %v", err)
// 		HandleError(c, errs.ErrInvalidJWT)
// 		return
// 	}
// 	accessTokenUserID, ok := accessTokenPayload["sub"].(string)
// 	if !ok {
// 		log.Printf("Invalid access token: %v", err)
// 		HandleError(c, errs.ErrInvalidJWT)
// 		return
// 	}
// 	refreshToken, err := middleware.ValidateJWT(tokenDTO.RefreshToken)
// 	if err != nil || !refreshToken.Valid {
// 		log.Printf("Invalid refresh token: %v", err)
// 		HandleError(c, errs.ErrInvalidJWT)
// 		return
// 	}
// 	refreshTokenUserID, err := middleware.GetUserIDFromTokenString(tokenDTO.RefreshToken)
// 	if err != nil {
// 		log.Printf("Invalid refresh token: %v", err)
// 		HandleError(c, errs.ErrInvalidJWT)
// 		return
// 	}
// 	if accessTokenUserID != refreshTokenUserID {
// 		log.Printf("Refresh token mismatch: %v", err)
// 		HandleError(c, errs.ErrRefreshMismatch)
// 		return
// 	}

// 	userID, ok := utils.GetUserID(c)
// 	if !ok {
// 		log.Printf("Cannot get user ID from token: %v", err)
// 		HandleError(c, errs.ErrInvalidJWT)
// 		return
// 	}

// 	userRole, ok := accessTokenPayload["scope"].(string)
// 	if !ok || userRole == "" || userRole == "refresh" {
// 		log.Printf("Cannot get user role from token: %v", err)
// 		HandleError(c, errs.ErrInvalidJWT)
// 		return
// 	}

// 	accessJWT, err := middleware.GenerateJWT(userID, "access", 60*24*30, userRole)
// 	if err != nil {
// 		log.Printf("Cannot generate access token: %v", err)
// 		HandleError(c, errs.ErrInternalServerError)
// 		return
// 	}

// 	refreshJWT, err := middleware.GenerateJWT(userID, "refresh", 60*24*7, "refresh")
// 	if err != nil {
// 		log.Printf("Cannot generate refresh token: %v", err)
// 		HandleError(c, errs.ErrInternalServerError)
// 		return
// 	}

// 	c.JSON(http.StatusOK, dto.Token{
// 		AccessToken:  accessJWT,
// 		RefreshToken: refreshJWT,
// 	})
// }
