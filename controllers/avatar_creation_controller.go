package controllers

import (
	"avazon-api/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type AvatarCreationController struct {
	avatarCreationService *services.AvatarCreateService
}

func NewAvatarCreationController(avatarCreationService *services.AvatarCreateService) *AvatarCreationController {
	return &AvatarCreationController{avatarCreationService: avatarCreationService}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		allowedOrigins := []string{
			"http://localhost:8081",
			"https://gid.cast-ing.kr",
		}

		origin := r.Header.Get("Origin")
		for _, o := range allowedOrigins {
			if o == origin {
				return true
			}
		}
		return false
	},
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
