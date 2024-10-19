package services

import "gorm.io/gorm"

type AvatarFunction string

const (
	AF_CreateImage     AvatarFunction = "create_avatar_image"
	AF_CreateCharacter AvatarFunction = "create_avatar_character"
	AF_CreateVoice     AvatarFunction = "create_avatar_voice"
)

type AvatarCreationService struct {
	DB *gorm.DB
}

func NewAvatarCreationService(db *gorm.DB) *AvatarCreationService {
	return &AvatarCreationService{DB: db}
}
