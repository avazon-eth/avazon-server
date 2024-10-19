package services

import (
	"avazon-api/models"
	"avazon-api/tools"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type SystemPromptService struct {
	DB              *gorm.DB
	CreateAssistant func() tools.Assistant
}

func NewSystemPromptService(
	db *gorm.DB,
	createAssistant func() tools.Assistant,
) *SystemPromptService {
	return &SystemPromptService{
		DB:              db,
		CreateAssistant: createAssistant,
	}
}

type Agent string

const (
	AG_AvatarImageCreationChat     Agent = "avatar_image_create_chat"
	AG_AvatarCharacterCreationChat Agent = "avatar_character_create_chat"
	AG_AvatarVoiceCreationChat     Agent = "avatar_voice_create_chat"
	AG_AvatarCharacterCreation     Agent = "avatar_character_create"
	AG_AvatarCharacterEdit         Agent = "avatar_character_edit"
	AG_AvatarVoiceCreation         Agent = "avatar_voice_create"
	AG_AvatarVoiceEdit             Agent = "avatar_voice_edit"
	AG_AvatarIntroduce             Agent = "avatar_introduce"
	AG_AvatarChatVideoPrompt       Agent = "avatar_chat_video_prompt"
	// content creation
	AG_MusicSummarizer          Agent = "music_summarizer"
	AG_MusicImagePromptCreation Agent = "music_image_prompt_create"
	AG_MusicPromptCreation      Agent = "music_create"
)

// usually used by other service components
func (s *SystemPromptService) GetSystemPrompt(promptAgent Agent) (string, error) {
	var promptUsage models.SystemPromptUsage
	if err := s.DB.Preload("Prompt").First(&promptUsage, "agent_id = ?", promptAgent).Error; err != nil {
		return "", fmt.Errorf("system prompt not found: %w", err)
	}
	return promptUsage.Prompt.Prompt, nil
}

// usually used by other service components
func (s *SystemPromptService) Use(agent Agent, input string) (string, error) {
	var systemPromptUsage models.SystemPromptUsage
	result := s.DB.Where("agent_id = ?", agent).Preload("Prompt").First(&systemPromptUsage)
	if result.Error != nil {
		return "", fmt.Errorf("system prompt not found for agent: %s, error: %w", agent, result.Error)
	}
	assistant := s.CreateAssistant()
	assistant.SetSystemPrompt(systemPromptUsage.Prompt.Prompt)
	return assistant.Handle(input)
}

// checks if a system prompt exists in the database and either creates or updates it accordingly.
func (s *SystemPromptService) UpsertSystemPrompt(prompt models.SystemPrompt) error {
	var existingPrompt models.SystemPrompt
	// Check for an existing system prompt by its ID
	result := s.DB.First(&existingPrompt, "id = ?", prompt.ID)

	// If there is an error and it's not a "not found" error, return the error
	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return fmt.Errorf("failed to check existing system prompt: %w", result.Error)
	}

	// If no existing prompt is found, create a new one
	if result.RowsAffected == 0 {
		if err := s.DB.Create(&prompt).Error; err != nil {
			return fmt.Errorf("failed to create system prompt: %w", err)
		}
	} else {
		// If an existing prompt is found, update it
		if err := s.DB.Save(&prompt).Error; err != nil {
			return fmt.Errorf("failed to update system prompt: %w", err)
		}
	}
	// Return nil if the operation was successful
	return nil
}

func (s *SystemPromptService) GetAllSystemPrompts() ([]models.SystemPrompt, error) {
	var prompts []models.SystemPrompt
	if err := s.DB.Find(&prompts).Error; err != nil {
		return nil, fmt.Errorf("system prompt not found: %w", err)
	}
	return prompts, nil
}

func (s *SystemPromptService) DeleteSystemPrompt(id string) error {
	if err := s.DB.Where("prompt_id = ?", id).Delete(&models.SystemPromptUsage{}).Error; err != nil {
		return err
	}
	if err := s.DB.Where("id = ?", id).Delete(&models.SystemPrompt{}).Error; err != nil {
		return err
	}
	return nil
}

func (s *SystemPromptService) UpdateSystemPromptUsage(agentID string, promptID string) error {
	var usage models.SystemPromptUsage
	result := s.DB.Where("agent_id = ?", agentID).First(&usage)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		// If not found, create a new usage entry
		usage = models.SystemPromptUsage{
			AgentID:  agentID,
			PromptID: promptID, // Assuming promptID is provided or needs to be set
		}
		if err := s.DB.Create(&usage).Error; err != nil {
			return fmt.Errorf("failed to create system prompt usage for agent: %s, error: %w", agentID, err)
		}
	} else if result.Error != nil {
		return fmt.Errorf("system prompt usage not found for agent: %s, error: %w", agentID, result.Error)
	}

	usage.PromptID = promptID
	if err := s.DB.Save(&usage).Error; err != nil {
		return fmt.Errorf("failed to update system prompt usage: %w", err)
	}
	return nil
}

func (s *SystemPromptService) GetAllSystemPromptUsages() ([]models.SystemPromptUsage, error) {
	var usages []models.SystemPromptUsage
	if err := s.DB.Preload("Prompt").Find(&usages).Error; err != nil {
		return nil, fmt.Errorf("system prompt usage not found: %w", err)
	}
	return usages, nil
}

func (s *SystemPromptService) DeleteSystemPromptUsage(agentID string) error {
	if err := s.DB.Delete(&models.SystemPromptUsage{}, "agent_id = ?", agentID).Error; err != nil {
		return fmt.Errorf("failed to delete system prompt usage for agent: %s, error: %w", agentID, err)
	}
	return nil
}
