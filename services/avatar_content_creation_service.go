package services

import (
	"avazon-api/tools"

	"gorm.io/gorm"
)

type AvatarContentCreationService struct {
	DB                *gorm.DB
	S3Service         *S3Service
	PromptService     *SystemPromptService
	AlbumImagePainter tools.Painter
	VideoImagePainter tools.Painter
	MusicProducer     tools.MusicProducer
	VideoProducer     tools.VideoProducer
}

func NewAvatarContentCreationService(
	db *gorm.DB,
	s3Service *S3Service,
	promptService *SystemPromptService,
	albumImagePainter tools.Painter,
	videoImagePainter tools.Painter,
	musicProducer tools.MusicProducer,
	videoProducer tools.VideoProducer,
) *AvatarContentCreationService {
	return &AvatarContentCreationService{
		DB:                db,
		S3Service:         s3Service,
		PromptService:     promptService,
		AlbumImagePainter: albumImagePainter,
		VideoImagePainter: videoImagePainter,
		MusicProducer:     musicProducer,
		VideoProducer:     videoProducer,
	}
}
