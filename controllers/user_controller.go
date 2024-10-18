package controllers

import (
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
		accessToken, err := middleware.GenerateJWT(user.ID, "access", 30, string(user.Role)) // 30 minutes
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

}
