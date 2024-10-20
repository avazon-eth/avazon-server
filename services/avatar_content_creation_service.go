package services

import (
	"avazon-api/controllers/errs"
	"avazon-api/dto"
	"avazon-api/models"
	"avazon-api/tools"
	"avazon-api/utils"
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
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

// called when video creation failed while progressing
func (s *AvatarContentCreationService) onVideoFailed(avatarVideo *models.AvatarVideoContentCreation, reason string) {
	avatarVideo.Status = models.ACC_Failed
	avatarVideo.FailedReason = &reason
	if err := s.DB.Model(&avatarVideo).Updates(avatarVideo).Error; err != nil {
		log.Printf("Error updating avatar video status: %v", err)
		return
	}
}

// called when music creation failed while progressing
func (s *AvatarContentCreationService) onMusicFailed(avatarMusic *models.AvatarMusicContentCreation, reason string) {
	avatarMusic.Status = models.ACC_Failed
	avatarMusic.FailedReason = &reason
	if err := s.DB.Model(&avatarMusic).Updates(avatarMusic).Error; err != nil {
		log.Printf("Error updating avatar music status: %v", err)
		return
	}
}

func (s *AvatarContentCreationService) CreateAvatarVideoImage(userID string, avatarID string, request dto.AvatarVideoImageRequest) (*models.AvatarVideoContentCreation, error) {
	var avatar models.Avatar

	if err := s.DB.Where("id = ?", avatarID).First(&avatar).Error; err != nil {
		log.Printf("Error fetching avatar: %v", err)
		return nil, err
	}

	imageURL := avatar.ProfileImageURL
	imageBytes, mimeType, err := utils.GetDataFromURL(imageURL)
	if err != nil {
		log.Printf("Error getting data from URL: %v", err)
		return nil, err
	}

	avatarVideo := &models.AvatarVideoContentCreation{
		ID:          uuid.New().String(),
		UserID:      userID,
		AvatarID:    avatarID,
		Avatar:      avatar,
		ImagePrompt: request.Prompt,
		Status:      models.ACC_Yet,
	}
	if err := s.DB.Create(avatarVideo).Error; err != nil {
		log.Printf("Error creating avatar video: %v", err)
		return nil, err
	}

	go func() {
		avatarVideo.Status = models.ACC_ImageProgressing
		if err := s.DB.Model(&avatarVideo).Updates(avatarVideo).Error; err != nil {
			avatarVideo.Status = models.ACC_Failed
			reason := err.Error()
			avatarVideo.FailedReason = &reason
			log.Printf("Error updating avatar video status to progressing: %v", err)
			return
		}

		newImageBytes, newImageMimeType, err := s.VideoImagePainter.PaintFromReference(imageBytes, mimeType, request.Prompt, 672, 1024)
		if err != nil {
			s.onVideoFailed(avatarVideo, err.Error())
			log.Printf("Error painting video image: %v", err)
			return
		}
		newImageExtension, err := utils.GetExtensionFromMimeType(newImageMimeType)
		if err != nil {
			s.onVideoFailed(avatarVideo, err.Error())
			log.Printf("Error getting extension from MIME type: %v", err)
			return
		}
		fileName := fmt.Sprintf("%s%s", uuid.New().String(), newImageExtension)

		// upload to S3
		uploadedURL, err := s.S3Service.UploadPublicFile(
			context.TODO(),
			fileName,
			newImageBytes,
			newImageMimeType,
		)
		if err != nil {
			s.onVideoFailed(avatarVideo, err.Error())
			log.Printf("Error uploading file to S3: %v", err)
			return
		}

		avatarVideo.ThumbnailImageURL = &uploadedURL
		if err := s.DB.Model(&avatarVideo).Updates(&avatarVideo).Error; err != nil {
			s.onVideoFailed(avatarVideo, err.Error())
			log.Printf("Error updating avatar video with thumbnail URL: %v", err)
			return
		}

		avatarVideo.Status = models.ACC_ImageCompleted
		if err := s.DB.Model(&avatarVideo).Updates(avatarVideo).Error; err != nil {
			s.onVideoFailed(avatarVideo, err.Error())
			log.Printf("Error updating avatar video status to completed: %v", err)
			return
		}
	}()

	return avatarVideo, nil
}

func (s *AvatarContentCreationService) CreateAvatarVideoFromImage(userID string, avatarID string, videoID string, request dto.AvatarVideoRequest) (*models.AvatarVideoContentCreation, error) {
	var avatarVideo *models.AvatarVideoContentCreation
	if err := s.DB.
		Where("id = ? AND avatar_id = ? AND user_id = ?", videoID, avatarID, userID).
		First(&avatarVideo).Error; err != nil {
		log.Printf("Error fetching avatar video: %v", err)
		return nil, err
	}

	if avatarVideo.Status == models.ACC_Yet {
		return nil, errs.ErrContentCreationNotStartedYet
	}
	if avatarVideo.Status == models.ACC_ImageProgressing {
		return nil, errs.ErrImageNotCompleted
	}
	if avatarVideo.Status == models.ACC_ContentProgressing {
		return nil, errs.ErrContentNotCompleted
	}
	if avatarVideo.Status == models.ACC_Confirmed {
		return nil, errs.ErrContentCreationAlreadyCompleted
	}

	go func() {
		avatarVideo.VideoPrompt = request.Prompt
		avatarVideo.Status = models.ACC_ContentProgressing
		if err := s.DB.Model(&avatarVideo).Updates(avatarVideo).Error; err != nil {
			s.onVideoFailed(avatarVideo, err.Error())
			log.Printf("Error updating avatar video status to content progressing: %v", err)
			return
		}

		videoBytes, err := s.VideoProducer.Create(*avatarVideo.ThumbnailImageURL, avatarVideo.VideoPrompt)
		if err != nil {
			s.onVideoFailed(avatarVideo, err.Error())
			log.Printf("Error creating video: %v", err)
			return
		}

		fileName := fmt.Sprintf("%s.%s", uuid.New().String(), "mp4")

		uploadedURL, err := s.S3Service.UploadPublicFile(
			context.TODO(),
			fileName,
			videoBytes,
			"video/mp4",
		)
		if err != nil {
			s.onVideoFailed(avatarVideo, err.Error())
			log.Printf("Error uploading video to S3: %v", err)
			return
		}

		avatarVideo.VideoContentURL = &uploadedURL
		if err := s.DB.Model(&avatarVideo).Updates(avatarVideo).Error; err != nil {
			s.onVideoFailed(avatarVideo, err.Error())
			log.Printf("Error updating avatar video with video URL: %v", err)
			return
		}

		avatarVideo.Status = models.ACC_ContentCompleted
		if err := s.DB.Model(&avatarVideo).Updates(avatarVideo).Error; err != nil {
			s.onVideoFailed(avatarVideo, err.Error())
			log.Printf("Error updating avatar video status to content completed: %v", err)
			return
		}
	}()

	return avatarVideo, nil
}

func (s *AvatarContentCreationService) CreateAvatarMusicImage(userID string, avatarID string, request dto.AvatarMusicRequest) (*models.AvatarMusicContentCreation, error) {
	// Fetch the avatar using the provided avatarID
	var avatar models.Avatar
	if err := s.DB.Where("id = ?", avatarID).First(&avatar).Error; err != nil {
		log.Printf("Error fetching avatar: %v", err)
		return nil, err
	}

	musicSummary, err := s.PromptService.Use(AG_MusicSummarizer, request.GetMusicInfo())
	if err != nil {
		log.Printf("Error summarizing music: %v", err)
		return nil, errs.ErrInternalServerError
	}

	avatarMusic := &models.AvatarMusicContentCreation{
		ID:                   uuid.New().String(),
		UserID:               userID,
		Title:                request.Title,
		Style:                request.Style,
		AvatarID:             avatarID,
		Avatar:               avatar,
		GeneratedMusicPrompt: &musicSummary,
		Status:               models.ACC_Yet,
	}
	if err := s.DB.Create(avatarMusic).Error; err != nil {
		log.Printf("Error creating avatar music: %v", err)
		return nil, err
	}

	go func() {
		avatarMusic.Status = models.ACC_ImageProgressing
		if err := s.DB.Model(&avatarMusic).Updates(avatarMusic).Error; err != nil {
			s.onMusicFailed(avatarMusic, err.Error())
			log.Printf("Error updating avatar music status to image progressing: %v", err)
			return
		}

		imagePrompt, err := s.PromptService.Use(AG_MusicImagePromptCreation, *avatarMusic.GeneratedMusicPrompt)
		if err != nil {
			s.onMusicFailed(avatarMusic, err.Error())
			log.Printf("Error creating image prompt: %v", err)
			return
		}

		avatarMusic.GeneratedImagePrompt = &imagePrompt
		if err := s.DB.Model(&avatarMusic).Updates(avatarMusic).Error; err != nil {
			s.onMusicFailed(avatarMusic, err.Error())
			log.Printf("Error updating avatar music with generated image prompt: %v", err)
			return
		}

		imageBytes, mimeType, err := s.AlbumImagePainter.Paint(imagePrompt, "", 1024, 1024)
		if err != nil {
			s.onMusicFailed(avatarMusic, err.Error())
			log.Printf("Error painting album image: %v", err)
			return
		}

		fileExtension, err := utils.GetExtensionFromMimeType(mimeType)
		if err != nil {
			s.onMusicFailed(avatarMusic, err.Error())
			log.Printf("Error getting extension from MIME type: %v", err)
			return
		}
		fileName := fmt.Sprintf("%s%s", uuid.New().String(), fileExtension)

		uploadedURL, err := s.S3Service.UploadPublicFile(
			context.TODO(),
			fileName,
			imageBytes,
			mimeType,
		)
		if err != nil {
			s.onMusicFailed(avatarMusic, err.Error())
			log.Printf("Error uploading album image to S3: %v", err)
			return
		}

		avatarMusic.AlbumImageURL = &uploadedURL
		if err := s.DB.Model(&avatarMusic).Updates(avatarMusic).Error; err != nil {
			s.onMusicFailed(avatarMusic, err.Error())
			log.Printf("Error updating avatar music with album image URL: %v", err)
			return
		}

		avatarMusic.Status = models.ACC_ImageCompleted
		if err := s.DB.Model(&avatarMusic).Updates(avatarMusic).Error; err != nil {
			s.onMusicFailed(avatarMusic, err.Error())
			log.Printf("Error updating avatar music status to image completed: %v", err)
			return
		}

		// TODO: May be separated into two steps
		s.CreateAvatarMusic(userID, avatarID, avatarMusic.ID)
	}()

	return avatarMusic, nil
}

func (s *AvatarContentCreationService) RegenerateAvatarMusicImage(userID string, avatarID string, musicID string) (*models.AvatarMusicContentCreation, error) {
	var mc *models.AvatarMusicContentCreation
	if err := s.DB.Where("id = ? AND avatar_id = ? AND user_id = ?", musicID, avatarID, userID).First(&mc).Error; err != nil {
		log.Printf("Error fetching avatar music: %v", err)
		return nil, err
	}

	if mc.Status == models.ACC_ImageProgressing {
		return nil, errs.ErrImageNotCompleted
	}
	if mc.Status == models.ACC_Confirmed {
		return nil, errs.ErrContentCreationAlreadyCompleted
	}

	go func() {
		mc.AlbumImageURL = nil
		mc.Status = models.ACC_ImageProgressing
		if err := s.DB.Model(&mc).Updates(mc).Error; err != nil {
			s.onMusicFailed(mc, err.Error())
			log.Printf("Error updating avatar music status to image progressing: %v", err)
			return
		}

		imagePrompt, err := s.PromptService.Use(AG_MusicImagePromptCreation, *mc.GeneratedMusicPrompt)
		if err != nil {
			s.onMusicFailed(mc, err.Error())
			log.Printf("Error creating image prompt: %v", err)
			return
		}

		mc.GeneratedImagePrompt = &imagePrompt
		if err := s.DB.Model(&mc).Updates(mc).Error; err != nil {
			s.onMusicFailed(mc, err.Error())
			log.Printf("Error updating avatar music with generated image prompt: %v", err)
			return
		}

		imageBytes, mimeType, err := s.AlbumImagePainter.Paint(imagePrompt, "", 1024, 1024)
		if err != nil {
			s.onMusicFailed(mc, err.Error())
			log.Printf("Error painting album image: %v", err)
			return
		}

		fileExtension, err := utils.GetExtensionFromMimeType(mimeType)
		if err != nil {
			s.onMusicFailed(mc, err.Error())
			log.Printf("Error getting extension from MIME type: %v", err)
			return
		}
		fileName := fmt.Sprintf("%s%s", uuid.New().String(), fileExtension)

		uploadedURL, err := s.S3Service.UploadPublicFile(
			context.TODO(),
			fileName,
			imageBytes,
			mimeType,
		)
		if err != nil {
			s.onMusicFailed(mc, err.Error())
			log.Printf("Error uploading album image to S3: %v", err)
			return
		}

		mc.AlbumImageURL = &uploadedURL
		if err := s.DB.Model(&mc).Updates(mc).Error; err != nil {
			s.onMusicFailed(mc, err.Error())
			log.Printf("Error updating avatar music with album image URL: %v", err)
			return
		}

		if mc.MusicURL != nil {
			mc.Status = models.ACC_ContentCompleted
		} else {
			mc.Status = models.ACC_ImageCompleted
		}
		if err := s.DB.Model(&mc).Updates(mc).Error; err != nil {
			s.onMusicFailed(mc, err.Error())
			log.Printf("Error updating avatar music status to image completed: %v", err)
			return
		}
	}()

	return mc, nil
}

func (s *AvatarContentCreationService) CreateAvatarMusic(userID string, avatarID string, musicID string) (*models.AvatarMusicContentCreation, error) {
	var avatarMusic *models.AvatarMusicContentCreation
	if err := s.DB.Where("id = ? AND avatar_id = ? AND user_id = ?", musicID, avatarID, userID).First(&avatarMusic).Error; err != nil {
		log.Printf("Error fetching avatar music: %v", err)
		return nil, err
	}

	if avatarMusic.Status == models.ACC_Yet {
		return nil, errs.ErrContentCreationNotStartedYet
	}
	if avatarMusic.Status == models.ACC_ImageProgressing {
		return nil, errs.ErrImageNotCompleted
	}
	if avatarMusic.Status == models.ACC_ContentProgressing {
		return nil, errs.ErrContentNotCompleted
	}
	if avatarMusic.Status == models.ACC_Confirmed {
		return nil, errs.ErrContentCreationAlreadyCompleted
	}

	go func() {
		avatarMusic.Status = models.ACC_ContentProgressing
		if err := s.DB.Model(&avatarMusic).Updates(avatarMusic).Error; err != nil {
			s.onMusicFailed(avatarMusic, err.Error())
			log.Printf("Error updating avatar music status to content progressing: %v", err)
			return
		}

		musicPrompt, err := s.PromptService.Use(AG_MusicPromptCreation, *avatarMusic.GeneratedMusicPrompt)
		if err != nil {
			s.onMusicFailed(avatarMusic, err.Error())
			log.Printf("Error creating music prompt: %v", err)
			return
		}
		avatarMusic.GeneratedMusicPrompt = &musicPrompt
		if err := s.DB.Model(&avatarMusic).Updates(avatarMusic).Error; err != nil {
			s.onMusicFailed(avatarMusic, err.Error())
			log.Printf("Error updating avatar music with generated music prompt: %v", err)
			return
		}

		musicBytes, err := s.MusicProducer.Produce(musicPrompt, avatarMusic.Title, avatarMusic.Style)
		if err != nil {
			s.onMusicFailed(avatarMusic, err.Error())
			log.Printf("Error producing music: %v", err)
			return
		}

		fileName := fmt.Sprintf("%s.%s", uuid.New().String(), "mp3")

		uploadedURL, err := s.S3Service.UploadPublicFile(
			context.TODO(),
			fileName,
			musicBytes,
			"audio/mpeg",
		)

		if err != nil {
			s.onMusicFailed(avatarMusic, err.Error())
			log.Printf("Error uploading music to S3: %v", err)
			return
		}

		avatarMusic.MusicURL = &uploadedURL
		if err := s.DB.Model(&avatarMusic).Updates(avatarMusic).Error; err != nil {
			s.onMusicFailed(avatarMusic, err.Error())
			log.Printf("Error updating avatar music with music URL: %v", err)
			return
		}

		avatarMusic.Status = models.ACC_ContentCompleted
		if err := s.DB.Model(&avatarMusic).Updates(avatarMusic).Error; err != nil {
			s.onMusicFailed(avatarMusic, err.Error())
			log.Printf("Error updating avatar music status to content completed: %v", err)
			return
		}
	}()

	return avatarMusic, nil
}

func (s *AvatarContentCreationService) GetAvatarMusicCreations(userID string, avatarID *string, page int, limit int) ([]*models.AvatarMusicContentCreation, error) {
	var musicCreations []*models.AvatarMusicContentCreation

	q := s.DB.Where("user_id = ?", userID).
		Offset(page * limit).
		Limit(limit).
		Order("created_at DESC")
	if avatarID != nil {
		q = q.Where("avatar_id = ?", avatarID)
	}

	if err := q.Find(&musicCreations).Error; err != nil {
		log.Printf("Error fetching avatar music creations: %v", err)
		return nil, err
	}
	return musicCreations, nil
}

func (s *AvatarContentCreationService) GetAvatarVideoCreations(userID string, avatarID *string, page int, limit int) ([]*models.AvatarVideoContentCreation, error) {
	var videoCreations []*models.AvatarVideoContentCreation

	q := s.DB.Where("user_id = ?", userID).
		Offset(page * limit).
		Limit(limit).
		Order("created_at DESC")
	if avatarID != nil {
		q = q.Where("avatar_id = ?", avatarID)
	}

	if err := q.Find(&videoCreations).Error; err != nil {
		log.Printf("Error fetching avatar video creations: %v", err)
		return nil, err
	}
	return videoCreations, nil
}

func (s *AvatarContentCreationService) GetAvatarMusicCreation(userID string, musicCreationID string) (*models.AvatarMusicContentCreation, error) {
	var music models.AvatarMusicContentCreation
	if err := s.DB.
		Where("id = ? AND user_id = ?", musicCreationID, userID).
		First(&music).Error; err != nil {
		log.Printf("Error fetching avatar music details: %v", err)
		return nil, err
	}
	return &music, nil
}

func (s *AvatarContentCreationService) GetAvatarVideoCreation(userID string, videoCreationID string) (*models.AvatarVideoContentCreation, error) {
	var video models.AvatarVideoContentCreation
	if err := s.DB.
		Where("id = ? AND user_id = ?", videoCreationID, userID).
		First(&video).Error; err != nil {
		log.Printf("Error fetching avatar video details: %v", err)
		return nil, err
	}
	return &video, nil
}

func (s *AvatarContentCreationService) ConfirmAvatarMusic(userID string, musicCreationID string, contentID string) (*models.AvatarMusic, error) {
	var AvatarMusicContentCreation *models.AvatarMusicContentCreation
	if err := s.DB.
		Where("id = ? AND user_id = ?", musicCreationID, userID).
		First(&AvatarMusicContentCreation).Error; err != nil {
		log.Printf("Error fetching avatar music creation: %v", err)
		return nil, err
	}

	if AvatarMusicContentCreation.Status != models.ACC_ContentCompleted {
		return nil, errs.ErrContentNotCompleted
	}

	AvatarMusicContentCreation.Status = models.ACC_Confirmed
	if err := s.DB.Model(&AvatarMusicContentCreation).Updates(AvatarMusicContentCreation).Error; err != nil {
		log.Printf("Error updating avatar music creation status to completed: %v", err)
		return nil, err
	}

	avatarMusic := models.AvatarMusic{
		ID:            contentID,
		UserID:        userID,
		User:          AvatarMusicContentCreation.User,
		Title:         AvatarMusicContentCreation.Title,
		AvatarID:      AvatarMusicContentCreation.AvatarID,
		Avatar:        AvatarMusicContentCreation.Avatar,
		AlbumImageURL: *AvatarMusicContentCreation.AlbumImageURL,
		MusicURL:      *AvatarMusicContentCreation.MusicURL,
	}

	if err := s.DB.Create(&avatarMusic).Error; err != nil {
		log.Printf("Error creating avatar music: %v", err)
		return nil, err
	}

	return &avatarMusic, nil
}

func (s *AvatarContentCreationService) ConfirmAvatarVideo(userID string, videoCreationID string, contentID string) (*models.AvatarVideo, error) {
	var AvatarVideoContentCreation *models.AvatarVideoContentCreation
	if err := s.DB.
		Where("id = ? AND user_id = ?", videoCreationID, userID).
		First(&AvatarVideoContentCreation).Error; err != nil {
		log.Printf("Error fetching avatar video creation: %v", err)
		return nil, err
	}

	if AvatarVideoContentCreation.Status == models.ACC_Confirmed {
		return nil, errs.ErrContentCreationAlreadyCompleted
	}
	if AvatarVideoContentCreation.Status != models.ACC_ContentCompleted {
		return nil, errs.ErrContentNotCompleted
	}

	AvatarVideoContentCreation.Status = models.ACC_Confirmed
	if err := s.DB.Model(&AvatarVideoContentCreation).Updates(AvatarVideoContentCreation).Error; err != nil {
		log.Printf("Error updating avatar video creation status to completed: %v", err)
		return nil, err
	}

	avatarVideo := models.AvatarVideo{
		ID:                contentID,
		UserID:            userID,
		User:              AvatarVideoContentCreation.User,
		AvatarID:          AvatarVideoContentCreation.AvatarID,
		Avatar:            AvatarVideoContentCreation.Avatar,
		ThumbnailImageURL: *AvatarVideoContentCreation.ThumbnailImageURL,
		VideoContentURL:   *AvatarVideoContentCreation.VideoContentURL,
	}

	if err := s.DB.Create(&avatarVideo).Error; err != nil {
		log.Printf("Error creating avatar video: %v", err)
		return nil, err
	}

	return &avatarVideo, nil
}
