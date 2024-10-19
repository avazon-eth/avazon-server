package services

import (
	"avazon-api/controllers/errs"
	"avazon-api/dto"
	"avazon-api/models"
	"avazon-api/tools"
	"avazon-api/utils"
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AvatarFunction string

const (
	AF_CreateImage     AvatarFunction = "create_avatar_image"
	AF_CreateCharacter AvatarFunction = "create_avatar_character"
	AF_CreateVoice     AvatarFunction = "create_avatar_voice"
)

type AvatarCreateService struct {
	AssistantCreator func() tools.Assistant
	sessions         map[string]*AvatarCreateSession
	tools            *AvatarCreateTools
	mu               sync.Mutex
}

func NewAvatarCreateService(
	db *gorm.DB,
	assistantCreator func() tools.Assistant,
	promptService *SystemPromptService,
	Painter tools.Painter,
	VoiceActor tools.VoiceActor,
	VideoProducer tools.VideoProducer,
	S3Service *S3Service,
) *AvatarCreateService {
	return &AvatarCreateService{
		AssistantCreator: assistantCreator,
		sessions:         make(map[string]*AvatarCreateSession),
		tools: &AvatarCreateTools{
			DB:            db,
			Painter:       Painter,
			VoiceActor:    VoiceActor,
			VideoProducer: VideoProducer,
			PromptService: promptService,
			S3Service:     S3Service,
		},
	}
}

type AvatarCreateSession struct {
	tools              *AvatarCreateTools // Use tools for various operations
	session            *models.AvatarCreation
	imageAssistant     tools.Assistant
	characterAssistant tools.Assistant
	voiceAssistant     tools.Assistant
	enteredAt          time.Time // when user entered the room
	updatedAt          time.Time // when user updated the room (if last update is more than 30 minutes, expire)
	mu                 sync.Mutex
}

type AvatarCreateTools struct {
	DB            *gorm.DB // DB needed for saving data
	Painter       tools.Painter
	VoiceActor    tools.VoiceActor
	VideoProducer tools.VideoProducer
	PromptService *SystemPromptService
	S3Service     *S3Service
}

func (s *AvatarCreateService) StartCreation(userID uint, req dto.AvatarCreationRequest) (models.AvatarCreation, error) {
	var existingCreation models.AvatarCreation
	s.tools.DB.Where("user_id = ? and status = ?", userID, models.AC_Processing).First(&existingCreation)
	if existingCreation.ID != "" {
		s.tools.DB.Delete(&existingCreation) // delete existing creation
	}

	avatarCreation := models.AvatarCreation{
		UserID:  userID,
		Name:    req.Name,
		Species: req.Species,
		Gender:  req.Gender,
		// Age:         req.Age,
		Language:    req.Language,
		Country:     req.Country,
		ImageStyle:  req.ImageStyle,
		Description: req.Description,
		StartedAt:   time.Now(),
		Status:      "ready",
	}
	// generate random id
	randomID := uuid.New().String()
	avatarCreation.ID = randomID

	if err := s.tools.DB.Create(&avatarCreation).Error; err != nil {
		return models.AvatarCreation{}, err
	}

	return avatarCreation, nil
}

// GetOneSession retrieves an avatar creation session by its ID.
// Parameters:
//   - avatarCreationID: The ID of the avatar creation session to retrieve.
//
// Returns:
//   - models.AvatarCreation: The retrieved avatar creation session.
//   - error: An error object if any error occurs during the retrieval process.
func (s *AvatarCreateService) GetOneSession(userID uint, avatarCreationID string) (models.AvatarCreation, error) {
	var avatarCreation models.AvatarCreation
	if err := s.tools.DB.
		Preload("ImageCreations", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC")
		}).
		Preload("CharacterCreations", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC")
		}).
		Preload("VoiceCreations", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC")
		}).
		Where("id = ? AND user_id = ?", avatarCreationID, userID).
		First(&avatarCreation).Error; err != nil {
		return models.AvatarCreation{}, err
	}
	return avatarCreation, nil
}

func (s *AvatarCreateService) CreateAvatarImage() {}

func (s *AvatarCreateService) GetCreateSessionChat(avatarCreationID string, objectType string, cursor int, size int) ([]models.AvatarCreationChat, error) {
	var chats []models.AvatarCreationChat
	if err := s.tools.DB.Where("avatar_creation_id = ? AND object_type = ? AND id < ?", avatarCreationID, objectType, cursor).
		Order("id DESC").
		Limit(size).
		Find(&chats).Error; err != nil {
		return []models.AvatarCreationChat{}, err
	}
	return chats, nil
}

// avatarID is for hashed NFT key
func (s *AvatarCreateService) CreateAvatar(userID uint, avatarCreationID string, avatarID string) (models.Avatar, error) {
	var existingAvatar models.Avatar
	s.tools.DB.Where("avatar_creation_id = ? AND user_id = ?", avatarCreationID, userID).First(&existingAvatar)
	if existingAvatar.ID != "" {
		return models.Avatar{}, errs.ErrAvatarAlreadyCreated
	}

	var avatarCreation models.AvatarCreation
	if err := s.tools.DB.Where("id = ? and user_id = ?", avatarCreationID, userID).
		Preload("ImageCreations", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC")
		}).
		Preload("CharacterCreations", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC")
		}).
		Preload("VoiceCreations", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC")
		}).
		First(&avatarCreation).Error; err != nil {
		return models.Avatar{}, err
	}

	createdImage := avatarCreation.GetCreatedImage()
	createdCharacter := avatarCreation.GetCreatedCharacter()
	createdVoice := avatarCreation.GetCreatedVoice()
	if createdImage == nil {
		return models.Avatar{}, errs.ErrImageNotCreated
	}
	if createdCharacter == nil {
		return models.Avatar{}, errs.ErrCharacterNotCreated
	}
	if createdVoice == nil {
		return models.Avatar{}, errs.ErrVoiceNotCreated
	}

	avatar := models.Avatar{
		ID:                   avatarID,
		UserID:               userID,
		AvatarCreationID:     avatarCreationID,
		Name:                 avatarCreation.Name,
		Species:              avatarCreation.Species,
		Gender:               avatarCreation.Gender,
		Language:             avatarCreation.Language,
		Country:              avatarCreation.Country,
		Description:          avatarCreation.Description,
		CreatedAt:            time.Now(),
		ProfileImageURL:      createdImage.ImageURL,
		VoiceURL:             createdVoice.VoiceURL,
		AvatarVideoURL:       nil,
		CharacterDescription: createdCharacter.Content,
	}

	if err := s.tools.DB.Create(&avatar).Error; err != nil {
		return models.Avatar{}, err
	}

	go func() {
		video, err := s.tools.VideoProducer.Create(avatar.ProfileImageURL, string(AG_AvatarChatVideoPrompt))
		if err != nil {
			log.Println(err)
		}

		fileName := fmt.Sprintf("%s_chat.mp4", avatarID)
		videoURL, err := s.tools.S3Service.UploadPublicFile(context.TODO(), fileName, video, "video/mp4")
		if err != nil {
			log.Println(err)
		}

		avatar.AvatarVideoURL = &videoURL
		s.tools.DB.Model(&avatar).Update("avatar_video_url", videoURL)
	}()

	return avatar, nil
}

func (s *AvatarCreateService) EnterSession(userID uint, sessionID string) (*AvatarCreateSession, error) {
	session, ok := s.sessions[sessionID]
	// if not exists, create new room
	if !ok {
		// find creation from DB
		var creation *models.AvatarCreation
		s.tools.DB.Where("user_id=? AND id=?", userID, sessionID).
			Preload("CharacterCreations").
			Preload("VoiceCreations").
			Preload("ImageCreations").
			First(&creation)
		if creation == nil {
			return nil, errors.New("session not found")
		}

		// create assistants
		// 1. image assistant
		imageAssistant := s.AssistantCreator()
		imagePrompt, err := s.tools.PromptService.GetSystemPrompt(AG_AvatarImageCreationChat)
		if err == nil {
			imageAssistant.SetSystemPrompt(imagePrompt)
		} else {
			log.Println("Failed to get image assistant prompt:", err)
			imageAssistant.SetSystemPrompt("You are a helpful assistant. Your'e helping user to create an avatar image.")
		}
		// set function
		imageAssistant.SetTools([]tools.OpenAITool{
			{
				Type: "function",
				Function: tools.OpenAIToolFunction{
					Name:        string(AF_CreateImage),
					Description: "You can request avatar image creation to server. You have to call this function when you think it is necessary. Summarize avatar appearance with avatar's basic information and user's chattings.",
					Parameters: tools.OpenAIFunctionParameters{
						Type: "object",
						Properties: map[string]tools.OpenAIParameter{
							"summary": {
								Type:        "string",
								Description: "Summary of current creating avatar based on avatar's basic information, and user's chattings. It MUST be shorter than 250 characters.",
							},
						},
						Required:             []string{"summary"},
						AdditionalProperties: false,
					},
				},
			},
		})

		// 2. character assistant
		characterAssistant := s.AssistantCreator()
		characterPrompt, err := s.tools.PromptService.GetSystemPrompt(AG_AvatarCharacterCreationChat)
		if err == nil {
			characterAssistant.SetSystemPrompt(characterPrompt)
		} else {
			log.Println("Failed to get character assistant prompt:", err)
			characterAssistant.SetSystemPrompt("You are a helpful assistant. Your'e helping user to create an avatar character. Especially for character personality, you have to make it more detailed and unique.")
		}
		// set function
		characterAssistant.SetTools([]tools.OpenAITool{
			{
				Type: "function",
				Function: tools.OpenAIToolFunction{
					Name:        string(AF_CreateCharacter),
					Description: "You can request avatar character creation to server. You have to call this function when you think it is necessary. Server has all chat details and information, so arguments are not needed.",
					Parameters: tools.OpenAIFunctionParameters{
						Type:                 "object",
						Properties:           map[string]tools.OpenAIParameter{},
						AdditionalProperties: false,
						Required:             []string{},
					},
				},
			},
		})

		// 3. voice assistant
		voiceAssistant := s.AssistantCreator()
		voicePrompt, err := s.tools.PromptService.GetSystemPrompt(AG_AvatarVoiceCreationChat)
		if err == nil {
			voiceAssistant.SetSystemPrompt(voicePrompt)
		} else {
			log.Println("Failed to get voice assistant prompt:", err)
			voiceAssistant.SetSystemPrompt("You are a helpful assistant. You're helping user to create an avatar voice. Do not ask user parameters directly. Inference them from user's chattings by yourself.")
		}
		// set function
		voiceAssistant.SetTools([]tools.OpenAITool{
			{
				Type: "function",
				Function: tools.OpenAIToolFunction{
					Name:        string(AF_CreateVoice),
					Description: "You can request avatar voice creation to server. You have to call this function when you think it is necessary. First, summarize avatar voice with avatar's basic information and user's chattings, and use it as input parameter in this function. Do not ask user parameters directly. Inference them from user's chattings by yourself.",
					Parameters: tools.OpenAIFunctionParameters{
						Type: "object",
						Properties: map[string]tools.OpenAIParameter{
							"summary": {
								Type:        "string",
								Description: "Summary of current creating avatar based on avatar's basic information, and user's chattings.",
							},
							"gender": {
								Type:        "string",
								Description: "Gender of avatar. It must be 'male' or 'female'.",
								Enum:        []string{"male", "female"},
							},
							"accent_strength": {
								Type:        "number",
								Description: "Accent strength of avatar voice. It has to be between 0.3 and 2.0.",
							},
							"age": {
								Type:        "string",
								Description: "Age of avatar voice. It must be 'young', 'middle_aged', or 'old'.",
								Enum:        []string{"young", "middle_aged", "old"},
							},
							"accent": {
								Type:        "string",
								Description: "Accent of avatar voice. It must be 'american', 'british', 'african', 'australian', or 'indian'.",
								Enum:        []string{"american", "british", "african", "australian", "indian"},
							},
						},
						Required:             []string{"summary", "gender", "accent_strength", "age", "accent"},
						AdditionalProperties: false,
					},
				},
			},
		})

		session = &AvatarCreateSession{
			session:            creation,
			tools:              s.tools,
			imageAssistant:     imageAssistant,
			characterAssistant: characterAssistant,
			voiceAssistant:     voiceAssistant,
		}
		s.sessions[sessionID] = session
	}

	session.enteredAt = time.Now()
	session.updatedAt = time.Now()

	return session, nil
}

func (s *AvatarCreateService) CloseSession(userID uint, sessionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, sessionID)
}

// call by controller when assistant chat is done
func (s *AvatarCreateService) SaveChat(sessionID string, role string, objectType string, content string) (models.AvatarCreationChat, error) {
	chat := models.AvatarCreationChat{
		AvatarCreationID: sessionID,
		Role:             role,
		ObjectType:       objectType,
		Content:          content,
	}
	return chat, s.tools.DB.Create(&chat).Error
}

// chat between user and image assistant
func (ss *AvatarCreateSession) HandleImageChat(userMessage string) (output chan string, done chan string, err chan error) {
	output, done, err = ss.imageAssistant.HandleAsync(userMessage)

	// go func() {
	// 	userChat := models.AvatarCreationChat{
	// 		AvatarCreationID:    ss.session.ID,
	// 		Role:                "user",
	// 		ObjectType:          "image",
	// 		Content:             userMessage,
	// 		ObjectCreatedNumber: len(ss.session.ImageCreations),
	// 	}
	// 	ss.tools.DB.Create(&userChat)
	// 	// Assistant Chat saving code will be called after output is done
	// }()

	return output, done, err
}

// chat between user and character assistant
func (ss *AvatarCreateSession) HandleCharacterChat(userMessage string) (output chan string, done chan string, err chan error) {
	output, done, err = ss.characterAssistant.HandleAsync(userMessage)

	go func() {
		userChat := models.AvatarCreationChat{
			AvatarCreationID:    ss.session.ID,
			Role:                "user",
			ObjectType:          "character",
			Content:             userMessage,
			CreatedObjectNumber: len(ss.session.CharacterCreations),
		}
		ss.tools.DB.Create(&userChat)
		// Assistant Chat saving code will be called after output is done
	}()

	return output, done, err
}

// chat between user and voice assistant
func (ss *AvatarCreateSession) HandleVoiceChat(userMessage string) (output chan string, done chan string, err chan error) {
	output, done, err = ss.voiceAssistant.HandleAsync(userMessage)

	go func() {
		userChat := models.AvatarCreationChat{
			AvatarCreationID:    ss.session.ID,
			Role:                "user",
			ObjectType:          "voice",
			Content:             userMessage,
			CreatedObjectNumber: len(ss.session.VoiceCreations),
		}
		ss.tools.DB.Create(&userChat)
		// Assistant Chat saving code will be called after output is done
	}()

	return output, done, err
}

func (ss *AvatarCreateSession) CanCreateNow(objectType string) bool {
	if objectType == "image" {
		if len(ss.session.ImageCreations) == 0 {
			return true
		}
		lastImage := ss.session.ImageCreations[len(ss.session.ImageCreations)-1]
		return lastImage.Status == models.AC_Completed || lastImage.Status == models.AC_Failed
	} else if objectType == "character" {
		if len(ss.session.CharacterCreations) == 0 {
			return true
		}
		lastCharacter := ss.session.CharacterCreations[len(ss.session.CharacterCreations)-1]
		return lastCharacter.Status == models.AC_Completed || lastCharacter.Status == models.AC_Failed
	} else if objectType == "voice" {
		if len(ss.session.VoiceCreations) == 0 {
			return true
		}
		lastVoice := ss.session.VoiceCreations[len(ss.session.VoiceCreations)-1]
		return lastVoice.Status == models.AC_Completed || lastVoice.Status == models.AC_Failed
	}
	return false
}

// create image from image chattings.
// This function is called by image assistant (Reference: https://platform.openai.com/docs/guides/function-calling)
//   - summary: the summary of avatar appearance (must be shorter than 270 characters)
func (ss *AvatarCreateSession) CreateImage(summary string) (<-chan models.AvatarImageCreation, error) {
	if !ss.CanCreateNow("image") {
		return nil, errors.New("image creation is blocked: the last image creation is not completed")
	}

	ss.mu.Lock()
	imageCreationChan := make(chan models.AvatarImageCreation)
	imageCreation := &models.AvatarImageCreation{
		AvatarCreationID: ss.session.ID,
		AvatarCreation:   *ss.session,
		Prompt:           "",
		Status:           models.AC_Ready,
	}
	ss.tools.DB.Create(&imageCreation)
	ss.session.ImageCreations = append(ss.session.ImageCreations, imageCreation)
	ss.mu.Unlock()

	go func() {
		imagePrompt := ""
		switch ss.session.ImageStyle {
		case "realistic":
			imagePrompt = "(realistic),"
		case "anime":
			imagePrompt = "(anime style),"
		case "cartoon":
			imagePrompt = "(2D animation style),"
		}

		imagePrompt += summary
		// imagePrompt += ",white background"
		// imagePrompt={style},{summary},white background
		imagePrompt, err := ss.tools.Painter.EnhancePrompt(imagePrompt)
		if err != nil {
			log.Println("Failed to enhance image prompt:", err)
			imageCreation.Status = models.AC_Failed
			imageCreation.FailedReason = err.Error()
			ss.tools.DB.Save(&imageCreation)
			imageCreationChan <- *imageCreation
			return
		}
		// image prompt done

		imageCreation.Status = models.AC_Processing
		imageCreation.Prompt = imagePrompt
		// request to painter
		imageCreationChan <- *imageCreation
		imageBytes, mimeType, err := ss.tools.Painter.Paint(imagePrompt, "", 682, 1024)
		if err != nil {
			log.Println("Failed to paint image:", err)
			imageCreation.Status = models.AC_Failed
			imageCreation.FailedReason = err.Error()
			ss.tools.DB.Save(&imageCreation)
			imageCreationChan <- *imageCreation
			return
		}

		extension, err := utils.GetExtensionFromMimeType(mimeType)
		if err != nil {
			log.Println("Failed to get extension from mime type:", err)
			imageCreation.Status = models.AC_Failed
			imageCreation.FailedReason = err.Error()
			ss.tools.DB.Save(&imageCreation)
			imageCreationChan <- *imageCreation
			return
		}
		fileName := fmt.Sprintf("%s_image_%d%s", imageCreation.AvatarCreationID, imageCreation.ID, extension)
		imageURL, err := ss.tools.S3Service.UploadPublicFile(context.TODO(), fileName, imageBytes, mimeType)
		if err != nil {
			log.Println("Failed to upload image to S3:", err)
			imageCreation.Status = models.AC_Failed
			imageCreation.FailedReason = err.Error()
			ss.tools.DB.Save(&imageCreation)
			imageCreationChan <- *imageCreation
			return
		}

		imageCreation.ImageURL = imageURL
		imageCreation.Status = models.AC_Completed
		ss.tools.DB.Save(&imageCreation)
		imageCreationChan <- *imageCreation
	}()

	return imageCreationChan, nil
}

// create character from character chattings.
// if current character creation exists, edit it.
func (ss *AvatarCreateSession) CreateCharacter() (<-chan models.AvatarCharacterCreation, error) {
	if !ss.CanCreateNow("character") {
		return nil, errors.New("character creation is blocked: the last character creation is not completed")
	}

	reqInputStr := ""
	editFlag := len(ss.session.CharacterCreations) > 0
	if editFlag {
		lastCharacter := ss.session.CharacterCreations[len(ss.session.CharacterCreations)-1]
		reqInputStr += fmt.Sprintf("[Character Prompt]\n%s\n\n", lastCharacter.Prompt)
	} else {
		reqInputStr += fmt.Sprintf("[Character Information]\n%s\n\n", ss.session.GetBasicInfo())
	}

	var characterChats []models.AvatarCreationChat
	ss.tools.DB.Where(
		"avatar_creation_id=? AND object_type=? AND object_created_number=?",
		ss.session.ID,
		"character",
		len(ss.session.CharacterCreations),
	).Find(&characterChats)
	chatsStr := "[Chattings]\n"
	for _, chat := range characterChats {
		chatsStr += fmt.Sprintf("%s: %s\n", chat.Role, chat.Content)
	}
	reqInputStr += chatsStr

	characterCreationChan := make(chan models.AvatarCharacterCreation)
	characterCreation := &models.AvatarCharacterCreation{
		AvatarCreationID: ss.session.ID,
		AvatarCreation:   *ss.session,
		Status:           models.AC_Ready,
	}
	ss.tools.DB.Create(characterCreation)
	ss.session.CharacterCreations = append(ss.session.CharacterCreations, characterCreation)

	go func() {
		characterCreation.Status = models.AC_Processing
		characterCreation.Prompt = reqInputStr
		characterCreationChan <- *characterCreation
		ss.tools.DB.Save(characterCreation)

		// generate character
		var result string
		var err error
		if editFlag {
			result, err = ss.tools.PromptService.Use(AG_AvatarCharacterEdit, reqInputStr)
		} else {
			result, err = ss.tools.PromptService.Use(AG_AvatarCharacterCreation, reqInputStr)
		}

		if err != nil {
			log.Println("Failed to create character:", err)
			characterCreation.Status = models.AC_Failed
			characterCreation.FailedReason = err.Error()
			ss.tools.DB.Save(characterCreation)
			characterCreationChan <- *characterCreation
			return
		}

		characterCreation.Content = result
		characterCreation.Status = models.AC_Completed
		ss.tools.DB.Save(characterCreation)
		characterCreationChan <- *characterCreation
	}()

	return characterCreationChan, nil
}

// create voice from voice chattings.
// This function is called by voice assistant (Reference: https://platform.openai.com/docs/guides/function-calling)
//   - summary: the summary of avatar voice
func (ss *AvatarCreateSession) CreateVoice(summary string, gender string, accentStrength string, age string, accent string) (<-chan models.AvatarVoiceCreation, error) {
	if !ss.CanCreateNow("voice") {
		return nil, errors.New("voice creation is blocked: the last voice creation is not completed")
	}

	reqInputStr := ""
	editFlag := len(ss.session.VoiceCreations) > 0
	if editFlag {
		lastVoice := ss.session.VoiceCreations[len(ss.session.VoiceCreations)-1]
		reqInputStr += fmt.Sprintf("[Voice Prompt]\n%s\n\n", lastVoice.Prompt)
	} else {
		reqInputStr += fmt.Sprintf("[Basic Information]\n%s\n\n", ss.session.GetBasicInfo())
	}

	var voiceChats []models.AvatarCreationChat
	ss.tools.DB.Where(
		"avatar_creation_id=? AND object_type=? AND object_created_number=?",
		ss.session.ID,
		"voice",
		len(ss.session.VoiceCreations),
	).Find(&voiceChats)
	chatsStr := "[Chattings]\n"
	for _, chat := range voiceChats {
		chatsStr += fmt.Sprintf("%s: %s\n", chat.Role, chat.Content)
	}
	reqInputStr += chatsStr

	voiceCreationChan := make(chan models.AvatarVoiceCreation)
	voiceCreation := &models.AvatarVoiceCreation{
		AvatarCreationID: ss.session.ID,
		AvatarCreation:   *ss.session,
		Status:           models.AC_Ready,
	}
	ss.tools.DB.Create(voiceCreation)
	ss.session.VoiceCreations = append(ss.session.VoiceCreations, voiceCreation)

	go func() {
		voiceCreation.Status = models.AC_Processing
		voiceCreationChan <- *voiceCreation
		ss.tools.DB.Save(voiceCreation)

		// 1. generate voice prompt
		var prompt string
		var err error
		if editFlag {
			prompt, err = ss.tools.PromptService.Use(AG_AvatarVoiceEdit, reqInputStr)
		} else {
			prompt, err = ss.tools.PromptService.Use(AG_AvatarVoiceCreation, reqInputStr)
		}
		if err != nil {
			log.Println("Failed to create voice:", err)
			voiceCreation.Status = models.AC_Failed
			voiceCreation.FailedReason = err.Error()
			ss.tools.DB.Save(voiceCreation)
			voiceCreationChan <- *voiceCreation
			return
		}
		voiceCreation.Prompt = prompt

		// 2. generate voice
		_, voiceId, err := ss.tools.VoiceActor.Create(voiceCreation.Prompt, models.Gender(gender), accentStrength, age, accent)
		if err != nil {
			log.Println("Failed to create voice:", err)
			voiceCreation.Status = models.AC_Failed
			voiceCreation.FailedReason = err.Error()
			ss.tools.DB.Save(voiceCreation)
			voiceCreationChan <- *voiceCreation
			return
		}

		// 3. create TTS and save to S3
		introduction, err := ss.tools.PromptService.Use(AG_AvatarIntroduce, ss.session.GetBasicInfo())
		if err != nil {
			log.Println("Failed to create introduction:", err)
			introduction = "Hello! I am your avatar. How are you?"
		}
		voiceBytes, err := ss.tools.VoiceActor.TTS(voiceId, introduction)
		if err != nil {
			log.Println("Failed to create voice:", err)
			voiceCreation.Status = models.AC_Failed
			voiceCreation.FailedReason = err.Error()
			ss.tools.DB.Save(voiceCreation)
			voiceCreationChan <- *voiceCreation
			return
		}
		fileName := fmt.Sprintf("%s_voice_%d.%s", ss.session.ID, voiceCreation.ID, "mp3")
		voiceURL, err := ss.tools.S3Service.UploadPublicFile(context.TODO(), fileName, voiceBytes, "audio/mpeg")
		if err != nil {
			log.Println("Failed to upload voice to S3:", err)
			voiceCreation.Status = models.AC_Failed
			voiceCreation.FailedReason = err.Error()
			ss.tools.DB.Save(voiceCreation)
			voiceCreationChan <- *voiceCreation
		}
		voiceCreation.VoiceURL = voiceURL
		voiceCreation.Status = models.AC_Completed
		ss.tools.DB.Save(voiceCreation)
		voiceCreationChan <- *voiceCreation
	}()

	return voiceCreationChan, nil
}

func (s *AvatarCreateService) CreateImageByRequest(userID uint, creationID string, userReq string) error {
	var avatarCreation models.AvatarCreation
	s.tools.DB.First(&avatarCreation, "id=?", creationID)
	if avatarCreation.UserID != userID {
		return errs.ErrNotFound
	}
	imageCreation := &models.AvatarImageCreation{
		AvatarCreationID: creationID,
		AvatarCreation:   avatarCreation,
		Prompt:           userReq,
		Status:           models.AC_Ready,
	}
	s.tools.DB.Create(&imageCreation)

	go func() {
		imagePrompt := userReq
		imagePrompt, err := s.tools.Painter.EnhancePrompt(imagePrompt)
		if err != nil {
			log.Println("Failed to enhance image prompt:", err)
			imageCreation.Status = models.AC_Failed
			imageCreation.FailedReason = err.Error()
			s.tools.DB.Save(&imageCreation)
			return
		}

		imageCreation.Status = models.AC_Processing
		imageCreation.Prompt = imagePrompt
		s.tools.DB.Save(&imageCreation)

		imageBytes, mimeType, err := s.tools.Painter.Paint(imagePrompt, "", 682, 1024)
		if err != nil {
			log.Println("Failed to paint image:", err)
			imageCreation.Status = models.AC_Failed
			imageCreation.FailedReason = err.Error()
			s.tools.DB.Save(&imageCreation)
			return
		}

		extension, err := utils.GetExtensionFromMimeType(mimeType)
		if err != nil {
			log.Println("Failed to get extension from mime type:", err)
			imageCreation.Status = models.AC_Failed
			imageCreation.FailedReason = err.Error()
			s.tools.DB.Save(&imageCreation)
			return
		}
		fileName := fmt.Sprintf("%s_image_%d%s", imageCreation.AvatarCreationID, imageCreation.ID, extension)
		imageURL, err := s.tools.S3Service.UploadPublicFile(context.TODO(), fileName, imageBytes, mimeType)
		if err != nil {
			log.Println("Failed to upload image to S3:", err)
			imageCreation.Status = models.AC_Failed
			imageCreation.FailedReason = err.Error()
			s.tools.DB.Save(&imageCreation)
			return
		}

		imageCreation.ImageURL = imageURL
		imageCreation.Status = models.AC_Completed
		s.tools.DB.Save(&imageCreation)
	}()

	return nil
}

func (s *AvatarCreateService) CreateCharacterByRequest(userID uint, creationID string, userReq string) error {
	var avatarCreation models.AvatarCreation
	s.tools.DB.First(&avatarCreation, "id=?", creationID)
	if avatarCreation.UserID != userID {
		return errs.ErrNotFound
	}

	characterPrompt, err := s.tools.PromptService.Use(AG_AvatarCharacterCreation, userReq)
	if err != nil {
		return err
	}

	characterCreation := &models.AvatarCharacterCreation{
		UserID:           userID,
		AvatarCreationID: creationID,
		AvatarCreation:   avatarCreation,
		Prompt:           characterPrompt,
		Content:          characterPrompt,
		Status:           models.AC_Completed,
	}
	s.tools.DB.Create(&characterCreation)

	return nil
}

func (s *AvatarCreateService) CreateVoiceByRequest(userID uint, creationID string, req dto.AvatarVoiceCreationRequest) error {
	var avatarCreation models.AvatarCreation
	s.tools.DB.First(&avatarCreation, "id=?", creationID)
	if avatarCreation.UserID != userID {
		return errs.ErrNotFound
	}

	voiceCreation := &models.AvatarVoiceCreation{
		UserID:           userID,
		AvatarCreationID: creationID,
		AvatarCreation:   avatarCreation,
		Status:           models.AC_Ready,
	}

	go func() {
		voiceCreation.Status = models.AC_Processing
		s.tools.DB.Save(voiceCreation)

		// 1. generate voice prompt
		var prompt string
		var err error
		prompt, err = s.tools.PromptService.Use(AG_AvatarVoiceCreation, req.Summary)
		if err != nil {
			log.Println("Failed to create voice:", err)
			voiceCreation.Status = models.AC_Failed
			voiceCreation.FailedReason = err.Error()
			s.tools.DB.Save(voiceCreation)
			return
		}
		voiceCreation.Prompt = prompt

		// 2. generate voice
		_, voiceId, err := s.tools.VoiceActor.Create(voiceCreation.Prompt, models.Gender(req.Gender), req.AccentStrength, req.Age, req.Accent)
		if err != nil {
			log.Println("Failed to create voice:", err)
			voiceCreation.Status = models.AC_Failed
			voiceCreation.FailedReason = err.Error()
			s.tools.DB.Save(voiceCreation)
			return
		}

		// 3. create TTS and save to S3
		introduction, err := s.tools.PromptService.Use(AG_AvatarIntroduce, avatarCreation.GetBasicInfo())
		if err != nil {
			log.Println("Failed to create introduction:", err)
			introduction = "Hello! I am your avatar. How are you?"
		}
		voiceBytes, err := s.tools.VoiceActor.TTS(voiceId, introduction)
		if err != nil {
			log.Println("Failed to create voice:", err)
			voiceCreation.Status = models.AC_Failed
			voiceCreation.FailedReason = err.Error()
			s.tools.DB.Save(voiceCreation)
			return
		}
		fileName := fmt.Sprintf("%s_voice_%d.%s", voiceCreation.AvatarCreationID, voiceCreation.ID, "mp3")
		voiceURL, err := s.tools.S3Service.UploadPublicFile(context.TODO(), fileName, voiceBytes, "audio/mpeg")
		if err != nil {
			log.Println("Failed to upload voice to S3:", err)
			voiceCreation.Status = models.AC_Failed
			voiceCreation.FailedReason = err.Error()
			s.tools.DB.Save(voiceCreation)
			return
		}
		voiceCreation.VoiceURL = voiceURL
		voiceCreation.Status = models.AC_Completed
		s.tools.DB.Save(voiceCreation)
	}()

	return nil
}

// don't use this. CreateCharacter() will handle this.
// edit character from character chattings
func (ss *AvatarCreateSession) EditCharacter() error {
	return nil
}

// don't use this. CreateVoice() will handle this.
// edit voice from voice chattings
func (ss *AvatarCreateSession) EditVoice() error {
	return nil
}

func (ss *AvatarCreateSession) ResponseAfterToolCalled(objectType string, message string, toolCallName string, toolCallArguments string, toolCallId string) (output chan string, done chan string, err chan error) {
	go func() {
		toolReq := models.AvatarCreationChat{
			AvatarCreationID:  ss.session.ID,
			Role:              "assistant",
			ObjectType:        objectType,
			Content:           "",
			ToolCallId:        toolCallId,
			ToolCallName:      toolCallName,
			ToolCallArguments: toolCallArguments,
		}
		ss.tools.DB.Create(&toolReq)
		toolChat := models.AvatarCreationChat{
			AvatarCreationID: ss.session.ID,
			Role:             "tool",
			ObjectType:       objectType,
			Content:          message,
		}
		ss.tools.DB.Create(&toolChat)
	}()
	switch objectType {
	case "image":
		return ss.imageAssistant.HandleAsync(message, "tool", toolCallId, toolCallName, toolCallArguments)
	case "character":
		return ss.characterAssistant.HandleAsync(message, "tool", toolCallId, toolCallName, toolCallArguments)
	case "voice":
		return ss.voiceAssistant.HandleAsync(message, "tool", toolCallId, toolCallName, toolCallArguments)
	default:
		return nil, nil, nil
	}
}

func (ss *AvatarCreateSession) Confirm() {
	// TODO: check all avatars are completed
	// TODO: create avatar
	ss.session.Status = models.AC_Completed
	ss.tools.DB.Save(ss.session)
}
