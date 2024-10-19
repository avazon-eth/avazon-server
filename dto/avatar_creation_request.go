package dto

import "avazon-api/models"

type AvatarImageCreationRequest struct {
	Summary string `json:"summary" binding:"required,notempty"` // ex) Summary of the image
}

type AvatarCharacterCreationRequest struct {
	Summary string `json:"summary" binding:"required,notempty"` // ex) Summary of the character
}

type AvatarVoiceCreationRequest struct {
	Summary        string        `json:"summary" binding:"required,notempty"`         // ex) Summary of the voice
	Gender         models.Gender `json:"gender" binding:"required,notempty"`          // ex) male, female, mixed, etc.
	AccentStrength string        `json:"accent_strength" binding:"required,notempty"` // ex) strong, moderate, light
	Age            string        `json:"age" binding:"required,notempty"`             // ex) young, middle-aged, old
	Accent         string        `json:"accent" binding:"required,notempty"`          // ex) American, British, etc.
}

type AvatarCreationRequest struct {
	Name        string `json:"name" binding:"required,notempty"`        // ex) John Doe
	Species     string `json:"species" binding:"required,notempty"`     // ex) human, alien, robot, etc.
	Gender      string `json:"gender" binding:"required,notempty"`      // ex) male, female, mixed, etc.
	Age         int    `json:"age" binding:"required,gte=1"`            // Age must be greater than 0
	Language    string `json:"language" binding:"required,notempty"`    // ex) English
	Country     string `json:"country" binding:"required,notempty"`     // ex) United States
	ImageStyle  string `json:"image_style" binding:"required,notempty"` // cartoon, realistic
	Description string `json:"description"`                             // optional
}
