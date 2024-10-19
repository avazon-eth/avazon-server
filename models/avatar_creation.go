package models

import (
	"fmt"
	"time"
)

type AvatarCreationStatus string

const (
	AC_Ready      AvatarCreationStatus = "ready"
	AC_Processing AvatarCreationStatus = "processing"
	AC_Completed  AvatarCreationStatus = "completed"
	AC_Failed     AvatarCreationStatus = "failed"
)

type AvatarCreation struct {
	ID                 string                     `json:"id" gorm:"primary_key;type:varchar(40);"` // UUID
	UserID             uint                       `json:"user_id"`
	User               User                       `json:"-" gorm:"foreignKey:UserID"`
	ImageCreations     []*AvatarImageCreation     `json:"images"`
	CharacterCreations []*AvatarCharacterCreation `json:"characters"`
	VoiceCreations     []*AvatarVoiceCreation     `json:"voices"`
	Name               string                     `json:"name" gorm:"type:varchar(50);not null"`
	Species            string                     `json:"species" gorm:"type:varchar(30);not null"`       // human, alien, robot, etc.
	Gender             string                     `json:"gender" gorm:"type:varchar(10);not null"`        // male, female, mixed, etc.
	Language           string                     `json:"language" gorm:"type:varchar(30);not null"`      // English default (will be changed to multi-language later)
	Country            string                     `json:"country" gorm:"type:varchar(30);not null"`       // free format (United States, Mars, etc.)
	Description        string                     `json:"description" gorm:"type:varchar(1000);not null"` // optional
	ImageStyle         string                     `json:"image_style" gorm:"type:varchar(20);not null"`   // realistic, anime, cartoon
	StartedAt          time.Time                  `json:"started_at" gorm:"not null"`                     // need to be updated when the avatar creation is started
	CompletedAt        time.Time                  `json:"completed_at"`                                   // need to be updated when the avatar creation is completed
	Status             AvatarCreationStatus       `json:"status" gorm:"type:varchar(20);not null"`        // ready, processing, completed, failed
}

type AvatarCreationChat struct {
	ID                  int            `json:"id" gorm:"primary_key;auto_increment"`
	AvatarCreationID    string         `json:"avatar_creation_id" gorm:"not null;foreignKey:AvatarCreationID;constraint:OnDelete:CASCADE"`
	AvatarCreation      AvatarCreation `json:"-" gorm:"foreignKey:AvatarCreationID;constraint:OnDelete:CASCADE"`
	Role                string         `json:"role" gorm:"type:varchar(20);not null"`        // user, assistant, tool
	ObjectType          string         `json:"object_type" gorm:"type:varchar(20);not null"` // image, character, voice
	CreatedAt           time.Time      `json:"created_at"`
	Content             string         `json:"content" gorm:"type:varchar(3000)"`
	CreatedObjectNumber int            `json:"created_object_number" gorm:"default:0"` // current number of objects created
	ToolCallId          string         `json:"tool_call_id" gorm:"type:varchar(255)"`
	ToolCallName        string         `json:"tool_call_name" gorm:"type:varchar(100)"`
	ToolCallArguments   string         `json:"tool_call_arguments" gorm:"type:varchar(1000)"`
}

type AvatarImageCreation struct {
	ID               int                  `json:"id" gorm:"primary_key;auto_increment"`
	UserID           uint                 `json:"user_id"`
	User             User                 `json:"-" gorm:"foreignKey:UserID"`
	AvatarCreationID string               `json:"avatar_creation_id" gorm:"not null;foreignKey:AvatarCreationID;constraint:OnDelete:CASCADE"`
	AvatarCreation   AvatarCreation       `json:"-" gorm:"foreignKey:AvatarCreationID;constraint:OnDelete:CASCADE"`
	Prompt           string               `json:"prompt" gorm:"type:varchar(3000)"`
	ImageURL         string               `json:"image_url" gorm:"not null"`
	Status           AvatarCreationStatus `json:"status"`
	FailedReason     string               `json:"failed_reason"` // reason for failure
	CreatedAt        time.Time            `json:"created_at"`
}

type AvatarCharacterCreation struct {
	ID               int                  `json:"id" gorm:"primary_key;auto_increment"`
	UserID           uint                 `json:"user_id"`
	User             User                 `json:"-" gorm:"foreignKey:UserID"`
	AvatarCreationID string               `json:"avatar_creation_id" gorm:"not null;foreignKey:AvatarCreationID;constraint:OnDelete:CASCADE"`
	AvatarCreation   AvatarCreation       `json:"-" gorm:"foreignKey:AvatarCreationID;constraint:OnDelete:CASCADE"`
	Prompt           string               `json:"prompt" gorm:"type:text"`           // input prompt
	Content          string               `json:"content" gorm:"type:varchar(3000)"` // result here
	Status           AvatarCreationStatus `json:"status"`
	FailedReason     string               `json:"failed_reason"` // reason for failure
	CreatedAt        time.Time            `json:"created_at"`
}

type AvatarVoiceCreation struct {
	ID               int                  `json:"id" gorm:"primary_key;auto_increment"`
	UserID           uint                 `json:"user_id"`
	User             User                 `json:"-" gorm:"foreignKey:UserID"`
	AvatarCreationID string               `json:"avatar_creation_id" gorm:"not null;foreignKey:AvatarCreationID;constraint:OnDelete:CASCADE"`
	AvatarCreation   AvatarCreation       `json:"-" gorm:"foreignKey:AvatarCreationID;constraint:OnDelete:CASCADE"`
	Prompt           string               `json:"prompt" gorm:"type:varchar(3000)"`
	VoiceURL         string               `json:"voice_url" gorm:"not null"`
	Status           AvatarCreationStatus `json:"status"`
	FailedReason     string               `json:"failed_reason"` // reason for failure
	CreatedAt        time.Time            `json:"created_at"`
}

func (ac *AvatarCreation) GetBasicInfo() string {
	basicInfo := fmt.Sprintf("Name: %s, Species: %s, Gender: %s, Language: %s, Country: %s, Description: %s", ac.Name, ac.Species, ac.Gender, ac.Language, ac.Country, ac.Description)
	if len(ac.CharacterCreations) > 0 {
		basicInfo += "\nCharacter: " + ac.CharacterCreations[len(ac.CharacterCreations)-1].Content
	}
	return basicInfo
}

func (ac *AvatarCreation) GetCreatedImage() *AvatarImageCreation {
	if len(ac.ImageCreations) == 0 {
		return nil
	}
	var newestImage *AvatarImageCreation
	for i := 0; i < len(ac.ImageCreations); i++ {
		if ac.ImageCreations[i].Status == AC_Completed {
			if newestImage == nil || ac.ImageCreations[i].ID > newestImage.ID {
				newestImage = ac.ImageCreations[i]
			}
		}
	}
	return newestImage
}

func (ac *AvatarCreation) GetCreatedCharacter() *AvatarCharacterCreation {
	if len(ac.CharacterCreations) == 0 {
		return nil
	}
	var newestCharacter *AvatarCharacterCreation
	for i := 0; i < len(ac.CharacterCreations); i++ {
		if ac.CharacterCreations[i].Status == AC_Completed {
			if newestCharacter == nil || ac.CharacterCreations[i].ID > newestCharacter.ID {
				newestCharacter = ac.CharacterCreations[i]
			}
		}
	}
	return newestCharacter
}

func (ac *AvatarCreation) GetCreatedVoice() *AvatarVoiceCreation {
	if len(ac.VoiceCreations) == 0 {
		return nil
	}
	var newestVoice *AvatarVoiceCreation
	for i := 0; i < len(ac.VoiceCreations); i++ {
		if ac.VoiceCreations[i].Status == AC_Completed {
			if newestVoice == nil || ac.VoiceCreations[i].ID > newestVoice.ID {
				newestVoice = ac.VoiceCreations[i]
			}
		}
	}
	return newestVoice
}
