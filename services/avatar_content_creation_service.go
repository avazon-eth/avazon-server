package services

import "gorm.io/gorm"

type AvatarContentCreationService struct {
	DB *gorm.DB
}

func NewAvatarContentCreationService(db *gorm.DB) *AvatarContentCreationService {
	return &AvatarContentCreationService{DB: db}
}
