package services

import (
	"avazon-api/controllers/errs"
	"avazon-api/dto"
	"avazon-api/models"
	"avazon-api/tools"
	"avazon-api/utils"
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AvatarRemixService struct {
	DB        *gorm.DB
	S3Service *S3Service
	Painter   tools.Painter
}

func NewAvatarRemixService(db *gorm.DB, s3Service *S3Service, painter tools.Painter) *AvatarRemixService {
	return &AvatarRemixService{DB: db, S3Service: s3Service, Painter: painter}
}

func (s *AvatarRemixService) onRemixImageFailed(avatarImageRemix *models.AvatarImageRemix, err error) {
	avatarImageRemix.Status = models.AR_Failed
	errorMessage := err.Error()
	avatarImageRemix.FailedReason = &errorMessage
	s.DB.Save(avatarImageRemix)
}

func (s *AvatarRemixService) updateImageRemixStatus(avatarImageRemix *models.AvatarImageRemix, status models.AvatarRemixStatus) {
	avatarImageRemix.Status = status
	s.DB.Save(avatarImageRemix)
}

func (s *AvatarRemixService) StartImageRemix(userID uint, avatarID string, request dto.AvatarImageRemixRequest) (*models.AvatarImageRemix, error) {
	var avatar models.Avatar
	if err := s.DB.Where("id = ?", avatarID).
		Preload("User").
		First(&avatar).Error; err != nil {
		return nil, err
	}

	avatarImageRemix := models.AvatarImageRemix{
		ID:         uuid.New().String(),
		UserID:     userID,
		AvatarID:   avatarID,
		Avatar:     avatar,
		UserPrompt: request.Prompt,
		Status:     models.AR_Yet,
	}
	if err := s.DB.Create(&avatarImageRemix).Error; err != nil {
		return nil, err
	}

	go func() {
		s.updateImageRemixStatus(&avatarImageRemix, models.AR_Progressing)
		avatarImageBytes, contentType, err := utils.GetDataFromURL(avatar.ProfileImageURL)
		if err != nil {
			s.onRemixImageFailed(&avatarImageRemix, err)
			return
		}

		remixImageBytes, remixContentType, err := s.Painter.ChangeStyle(avatarImageBytes, contentType, request.Prompt)
		if err != nil {
			s.onRemixImageFailed(&avatarImageRemix, err)
			return
		}

		fileExtension, err := utils.GetExtensionFromMimeType(remixContentType)
		if err != nil {
			s.onRemixImageFailed(&avatarImageRemix, err)
			return
		}
		filename := fmt.Sprintf("remix%s%s", avatarImageRemix.ID, fileExtension)
		uploadedURL, err := s.S3Service.UploadPublicFile(context.TODO(), filename, remixImageBytes, remixContentType)
		if err != nil {
			s.onRemixImageFailed(&avatarImageRemix, err)
			return
		}

		avatarImageRemix.ImageURL = &uploadedURL
		avatarImageRemix.Status = models.AR_Completed
		s.DB.Save(&avatarImageRemix)
	}()

	return &avatarImageRemix, nil
}

func (s *AvatarRemixService) GetOneImageRemix(userID uint, avatarID string, remixID string) (*models.AvatarImageRemix, error) {
	var avatarImageRemix models.AvatarImageRemix
	if err := s.DB.
		Where("id = ? AND user_id = ? AND avatar_id = ?", remixID, userID, avatarID).
		First(&avatarImageRemix).Error; err != nil {
		return nil, err
	}
	return &avatarImageRemix, nil
}

func (s *AvatarRemixService) ConfirmAvatarFromImageRemix(userID uint, avatarID string, remixID string, newAvatarID string) (*models.Avatar, error) {
	var avatarImageRemix models.AvatarImageRemix
	if err := s.DB.
		Where("id = ? AND avatar_id = ? AND user_id = ?", remixID, avatarID, userID).
		First(&avatarImageRemix).Error; err != nil {
		return nil, err
	}
	if avatarImageRemix.Status != models.AR_Completed {
		return nil, errs.ErrInvalidStatus
	}

	var originalAvatar models.Avatar
	if err := s.DB.Where("id = ?", avatarID).First(&originalAvatar).Error; err != nil {
		return nil, err
	}

	remixedAvatar := models.Avatar{
		ID:                   newAvatarID,
		UserID:               userID,
		AvatarCreationID:     originalAvatar.AvatarCreationID,
		RemixAvatarID:        &originalAvatar.ID,
		Name:                 originalAvatar.Name,
		Species:              originalAvatar.Species,
		Gender:               originalAvatar.Gender,
		Language:             originalAvatar.Language,
		Country:              originalAvatar.Country,
		Description:          originalAvatar.Description,
		ProfileImageURL:      *avatarImageRemix.ImageURL,
		VoiceURL:             originalAvatar.VoiceURL,
		AvatarVideoURL:       originalAvatar.AvatarVideoURL,
		CharacterDescription: originalAvatar.CharacterDescription,
	}
	s.DB.Create(&remixedAvatar)
	s.updateImageRemixStatus(&avatarImageRemix, models.AR_Confirmed)
	return &remixedAvatar, nil
}
