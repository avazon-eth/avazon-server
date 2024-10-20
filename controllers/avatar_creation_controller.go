package controllers

import (
	"avazon-api/controllers/errs"
	"avazon-api/dto"
	"avazon-api/middleware"
	"avazon-api/models"
	"avazon-api/services"
	"avazon-api/utils"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type AvatarCreationController struct {
	AvatarCreationService *services.AvatarCreateService
}

func NewAvatarCreationController(avatarCreationService *services.AvatarCreateService) *AvatarCreationController {
	return &AvatarCreationController{AvatarCreationService: avatarCreationService}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		allowedOrigins := []string{
			"http://localhost:8081",
			"https://gid.cast-ing.kr",
			"https://staging.d9xje8vs9f8su.amplifyapp.com",
			"https://avazon.cast-ing.kr",
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
	userID, ok := utils.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	var req dto.AvatarCreationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		HandleError(c, err)
		return
	}

	avatarCreation, err := ctrl.AvatarCreationService.StartCreation(userID, req)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, avatarCreation)
}

// used for tracking how the session is going
func (ctrl *AvatarCreationController) GetOneSession(c *gin.Context) {
	avatarCreationID := c.Param("creation_id")
	userID, ok := utils.GetUserID(c)
	if !ok {
		HandleError(c, errs.ErrUnauthorized)
		return
	}
	avatarCreation, err := ctrl.AvatarCreationService.GetOneSession(userID, avatarCreationID)
	if err != nil {
		HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, avatarCreation)
}

func (ctrl *AvatarCreationController) CreateAvatarImage(c *gin.Context) {
	creationID := c.Param("creation_id")
	userID, ok := utils.GetUserID(c)
	if !ok {
		HandleError(c, errs.ErrUnauthorized)
		return
	}
	var req dto.AvatarImageCreationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		HandleError(c, err)
		return
	}
	err := ctrl.AvatarCreationService.CreateImageByRequest(userID, creationID, req.Summary)
	if err != nil {
		HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Image creation started"})
}

func (ctrl *AvatarCreationController) CreateAvatarCharacter(c *gin.Context) {
	creationID := c.Param("creation_id")
	userID, ok := utils.GetUserID(c)
	if !ok {
		HandleError(c, errs.ErrUnauthorized)
		return
	}
	var req dto.AvatarCharacterCreationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		HandleError(c, err)
		return
	}
	err := ctrl.AvatarCreationService.CreateCharacterByRequest(userID, creationID, req.Summary)
	if err != nil {
		HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Character creation started"})
}

func (ctrl *AvatarCreationController) CreateAvatarVoice(c *gin.Context) {
	creationID := c.Param("creation_id")
	userID, ok := utils.GetUserID(c)
	if !ok {
		HandleError(c, errs.ErrUnauthorized)
		return
	}
	var req dto.AvatarVoiceCreationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		HandleError(c, err)
		return
	}
	err := ctrl.AvatarCreationService.CreateVoiceByRequest(userID, creationID, req)
	if err != nil {
		HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Voice creation started"})
}

func (ctrl *AvatarCreationController) CreateAvatar(c *gin.Context) {
	avatarCreationID := c.Param("creation_id")
	userID, ok := utils.GetUserID(c)
	if !ok {
		HandleError(c, errs.ErrUnauthorized)
		return
	}
	avatarID := c.Query("avatar_id")
	if avatarID == "" {
		HandleError(c, errs.ErrBadRequest)
		return
	}

	avatar, err := ctrl.AvatarCreationService.CreateAvatar(userID, avatarCreationID, avatarID)
	if err != nil {
		HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, avatar)
}

// ========================================================
// Avatar Create Session
// ========================================================

type AvatarCreateRequest struct {
	ObjectType string `json:"object_type"` // image, character, voice, all
	Event      string `json:"event"`       // chat, close, confirm
	Content    string `json:"content"`
}

type AvatarCreateResponse struct {
	ObjectType string `json:"object_type"` // image, character, voice, all
	Event      string `json:"event"`       // chat, chunk, creation, close, error
	Content    string `json:"content"`
}

// GET /avatar/create/:id/enter/
// websocket upgrade here
func (ctrl *AvatarCreationController) EnterSession(c *gin.Context) {
	avatarCreationID := c.Param("creation_id")
	// userID, ok := utils.GetUserID(c)
	// if !ok {
	// 	HandleError(c, errs.ErrUnauthorized)
	// 	return
	// }

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("WebSocket Upgrade Error:", err)
		return
	}
	defer conn.Close()

	accessTokenBody := &struct {
		AccessToken string `json:"access_token"`
	}{}
	if err := conn.ReadJSON(accessTokenBody); err != nil {
		log.Println("Error reading access token:", err)
		return
	}
	userID, err := middleware.GetUserIDFromTokenString(accessTokenBody.AccessToken)
	if err != nil {
		log.Println("Invalid access token:", err)
		return
	}

	session, err := ctrl.AvatarCreationService.EnterSession(userID, avatarCreationID)
	if err != nil {
		log.Println("Wrong session ID or unauthorized access")
		conn.WriteMessage(websocket.TextMessage, []byte("Invalid State"))
		conn.Close()
		return
	}
	log.Println("Starting session for avatar creation ID:", avatarCreationID)

	// message := map[string]string{"status": "OK"}
	// jsonMessage, _ := json.Marshal(message)
	// conn.WriteMessage(websocket.TextMessage, jsonMessage)

	// receive message from client
	clientMessageChan := make(chan AvatarCreateRequest)
	closeChan := make(chan struct{})
	go func() {
		defer close(clientMessageChan)
		defer close(closeChan)
		for {
			var req AvatarCreateRequest
			err := conn.ReadJSON(&req)
			if err != nil {
				log.Println("Error reading message:", err)
				closeChan <- struct{}{}
				return
			}
			clientMessageChan <- req
		}
	}()

	imageChatMu := sync.Mutex{}
	characterChatMu := sync.Mutex{}
	voiceChatMu := sync.Mutex{}
	for {
		select {
		case req := <-clientMessageChan:
			if req.Event == "chat" {
				objectType := req.ObjectType
				// save user chat
				if req.Content != "" {
					chat, err := ctrl.AvatarCreationService.SaveChat(avatarCreationID, "user", objectType, req.Content)
					if err != nil {
						log.Println("Error saving chat:", err)
						return
					}
					jsonChat, err := json.Marshal(chat)
					if err != nil {
						log.Println("Error marshalling chat to JSON:", err)
						return
					}
					conn.WriteJSON(AvatarCreateResponse{Event: "chat", ObjectType: objectType, Content: string(jsonChat)})
				}

				var outputChan chan string
				var doneChan chan string
				var errorChan chan error
				switch objectType {
				case "image":
					imageChatMu.Lock()
					outputChan, doneChan, errorChan = session.HandleImageChat(req.Content)
				case "character":
					characterChatMu.Lock()
					outputChan, doneChan, errorChan = session.HandleCharacterChat(req.Content)
				case "voice":
					voiceChatMu.Lock()
					outputChan, doneChan, errorChan = session.HandleVoiceChat(req.Content)
				default:
					log.Println("Invalid object type:", objectType)
					continue
				}
				// handle chat response
				go func() {
					switch objectType {
					case "image":
						defer imageChatMu.Unlock()
					case "character":
						defer characterChatMu.Unlock()
					case "voice":
						defer voiceChatMu.Unlock()
					}

					functionName := ""
					functionHandled := false
					toolCallId := ""
					for {
						select {
						case output := <-outputChan:
							if strings.HasPrefix(output, "function:") {
								// Handle function call
								// You can add your logic here to process the function call
								log.Println("Function call detected:", output)
								if functionHandled {
									log.Println("Function call already exists:", functionName)
									continue
								}
								fCallJsonStr := strings.TrimPrefix(output, "function:")
								fCallJson := map[string]string{}
								err := json.Unmarshal([]byte(fCallJsonStr), &fCallJson)
								if err != nil {
									log.Println("Error unmarshalling function call:", err)
									continue
								}
								functionName = fCallJson["name"]
								toolCallId = fCallJson["id"]
							} else {
								conn.WriteJSON(AvatarCreateResponse{Event: "chunk", ObjectType: objectType, Content: output})
							}
						case message := <-doneChan:
							nonJsonStr, jsonStrList := utils.SeparateTextAndJSON(message)

							// handle function call
							if functionName != "" && len(jsonStrList) > 0 {
								if !session.CanCreateNow(objectType) {
									continue // pass this turn
								}

								switch functionName {
								case "create_avatar_image":
									var summary string
									var result map[string]interface{}
									if err := json.Unmarshal([]byte(jsonStrList[0]), &result); err != nil {
										log.Println("Error parsing message to JSON:", message, err)
										errorChan <- err
										continue
									}
									summary = result["summary"].(string)
									imageChan, err := session.CreateImage(summary)
									if err != nil {
										log.Println("Error creating image:", err)
										conn.WriteJSON(AvatarCreateResponse{Event: "error", ObjectType: objectType, Content: err.Error()})
										return
									}
									// handle image creation
									go func() {
										for image := range imageChan {
											imageJson, err := json.Marshal(image)
											if err != nil {
												log.Println("Error marshalling image to JSON:", err)
												return
											}
											conn.WriteJSON(AvatarCreateResponse{Event: "creation", ObjectType: objectType, Content: string(imageJson)})
											if image.Status == models.AC_Completed || image.Status == models.AC_Failed {
												return
											}
										}
									}()
								case "create_avatar_character":
									characterChan, err := session.CreateCharacter()
									if err != nil {
										log.Println("Error creating character:", err)
										return
									}
									// handle character creation
									go func() {
										for character := range characterChan {
											characterJson, err := json.Marshal(character)
											if err != nil {
												log.Println("Error marshalling character to JSON:", err)
												return
											}
											conn.WriteJSON(AvatarCreateResponse{Event: "creation", ObjectType: objectType, Content: string(characterJson)})
											if character.Status == models.AC_Completed || character.Status == models.AC_Failed {
												return
											}
										}
									}()
								case "create_avatar_voice":
									var summary string
									var gender string
									var accentStrength string
									var age string
									var accent string
									var result map[string]interface{}
									if err := json.Unmarshal([]byte(jsonStrList[0]), &result); err != nil {
										log.Println("Error parsing message to JSON:", message, err)
										return
									}
									summary = result["summary"].(string)
									gender = result["gender"].(string)
									accentStrength = fmt.Sprintf("%f", result["accent_strength"].(float64))
									age = result["age"].(string)
									accent = result["accent"].(string)
									voiceChan, err := session.CreateVoice(summary, gender, accentStrength, age, accent)
									if err != nil {
										log.Println("Error creating voice:", err)
										errorChan <- err
										continue
									}
									// handle voice creation
									go func() {
										for voice := range voiceChan {
											voiceJson, err := json.Marshal(voice)
											if err != nil {
												log.Println("Error marshalling voice to JSON:", err)
												errorChan <- err
												return
											}
											conn.WriteJSON(AvatarCreateResponse{Event: "creation", ObjectType: objectType, Content: string(voiceJson)})
											if voice.Status == models.AC_Completed || voice.Status == models.AC_Failed {
												return
											}
										}
									}()
								} // end of switch functionName
								outputChan, doneChan, errorChan = session.ResponseAfterToolCalled(objectType, "creation started", functionName, jsonStrList[0], toolCallId)
								if outputChan == nil || doneChan == nil || errorChan == nil {
									log.Println("Error creating output channel:", err)
									errorChan <- err
									continue
								}
								functionName = "" // reset function name (if not reset, it will be used again here)
								toolCallId = ""
								functionHandled = true
							}
							if nonJsonStr != "" { // handle chat response
								chat, err := ctrl.AvatarCreationService.SaveChat(avatarCreationID, "assistant", objectType, nonJsonStr)
								if err != nil {
									log.Println("Error saving chat:", err)
									errorChan <- err
									continue
								}
								jsonChat, err := json.Marshal(chat)
								if err != nil {
									log.Println("Error marshalling chat to JSON:", err)
									errorChan <- err
									continue
								}
								conn.WriteJSON(AvatarCreateResponse{Event: "chat", ObjectType: objectType, Content: string(jsonChat)})
							}
						case err := <-errorChan:
							if err != nil {
								conn.WriteJSON(AvatarCreateResponse{Event: "error", ObjectType: objectType, Content: err.Error()})
							}
							return
						}
					}
				}()
			} else if req.Event == "confirm" {
				session.Confirm()
			} else if req.Event == "close" {
				ctrl.AvatarCreationService.CloseSession(userID, avatarCreationID)
				conn.WriteJSON(AvatarCreateResponse{ObjectType: "all", Event: "close", Content: "session closed"})
				return
			}
		case <-closeChan:
			ctrl.AvatarCreationService.CloseSession(userID, avatarCreationID)
			conn.WriteJSON(AvatarCreateResponse{ObjectType: "all", Event: "close", Content: "session closed"})
			return
		}
	}
}
