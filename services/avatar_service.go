package services

import (
	"avazon-api/models"

	"gorm.io/gorm"
)

type AvatarService struct {
	DB *gorm.DB
}

func NewAvatarService(db *gorm.DB) *AvatarService {
	return &AvatarService{DB: db}
}

func (s *AvatarService) GetAvatars(page int, limit int) ([]models.Avatar, error) {
	var avatars []models.Avatar
	if err := s.DB.
		Model(&models.Avatar{}).
		Limit(limit).
		Offset(page * limit).
		Order("created_at DESC").
		Find(&avatars).Error; err != nil {
		return nil, err
	}
	return avatars, nil
}

func (s *AvatarService) GetMyAvatars(userID uint, page int, limit int) ([]models.Avatar, error) {
	var avatars []models.Avatar
	if err := s.DB.
		Model(&models.Avatar{}).
		Where("user_id = ?", userID).
		Limit(limit).
		Offset(page * limit).
		Order("created_at DESC").
		Find(&avatars).Error; err != nil {
		return nil, err
	}
	return avatars, nil
}

func (s *AvatarService) GetAvatarsCount() (int64, error) {
	var count int64
	if err := s.DB.Model(&models.Avatar{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (s *AvatarService) GetMyAvatarsCount(userID uint) (int64, error) {
	var count int64
	if err := s.DB.Model(&models.Avatar{}).Where("user_id = ?", userID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (s *AvatarService) GetOneAvatar(avatarID string) (*models.Avatar, error) {
	var avatar models.Avatar
	if err := s.DB.
		Preload("User").
		Where("id = ?", avatarID).First(&avatar).Error; err != nil {
		return nil, err
	}
	return &avatar, nil
}

func (s *AvatarService) GetAvatarMusicContents(avatarID *string, page int, limit int) ([]models.AvatarMusic, error) {
	var avatarMusicContents []models.AvatarMusic

	q := s.DB.Model(&models.AvatarMusic{}).Preload("User").Preload("Avatar")

	if avatarID != nil {
		q = q.Where("avatar_id = ?", avatarID)
	}

	if err := q.
		Limit(limit).
		Offset(page * limit).
		Order("created_at DESC").
		Find(&avatarMusicContents).Error; err != nil {
		return nil, err
	}
	return avatarMusicContents, nil
}

func (s *AvatarService) GetOneAvatarMusicContent(musicContentID string) (*models.AvatarMusic, error) {
	var avatarMusicContent models.AvatarMusic
	if err := s.DB.
		Preload("User").
		Preload("Avatar").
		Where("id = ?", musicContentID).First(&avatarMusicContent).Error; err != nil {
		return nil, err
	}
	return &avatarMusicContent, nil
}

func (s *AvatarService) GetAvatarVideoContents(avatarID *string, page int, limit int) ([]models.AvatarVideo, error) {
	var avatarVideoContents []models.AvatarVideo

	q := s.DB.Model(&models.AvatarVideo{}).Preload("User").Preload("Avatar")

	if avatarID != nil {
		q = q.Where("avatar_id = ?", avatarID)
	}

	if err := q.
		Limit(limit).
		Offset(page * limit).
		Order("created_at DESC").
		Find(&avatarVideoContents).Error; err != nil {
		return nil, err
	}
	return avatarVideoContents, nil
}

func (s *AvatarService) GetOneAvatarVideoContent(videoContentID string) (*models.AvatarVideo, error) {
	var avatarVideoContent models.AvatarVideo
	if err := s.DB.
		Preload("User").
		Preload("Avatar").
		Where("id = ?", videoContentID).First(&avatarVideoContent).Error; err != nil {
		return nil, err
	}
	return &avatarVideoContent, nil
}

func (s *AvatarService) GetAvatarMusicContentsCount(avatarID *string) (int64, error) {
	var count int64

	q := s.DB.Model(&models.AvatarMusic{}).Preload("User").Preload("Avatar")

	if avatarID != nil {
		q = q.Where("avatar_id = ?", avatarID)
	}

	if err := q.Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

func (s *AvatarService) GetAvatarVideoContentsCount(avatarID *string) (int64, error) {
	var count int64

	q := s.DB.Model(&models.AvatarVideo{}).Preload("User").Preload("Avatar")

	if avatarID != nil {
		q = q.Where("avatar_id = ?", avatarID)
	}

	if err := q.Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

func (s *AvatarService) GetMyAvatarMusicContents(userID uint, page int, limit int) ([]models.AvatarMusic, error) {
	var avatarMusicContents []models.AvatarMusic
	if err := s.DB.
		Model(&models.AvatarMusic{}).
		Preload("User").
		Preload("Avatar").
		Where("user_id = ?", userID).
		Limit(limit).
		Offset(page * limit).
		Order("created_at DESC").
		Find(&avatarMusicContents).Error; err != nil {
		return nil, err
	}
	return avatarMusicContents, nil
}

func (s *AvatarService) GetMyAvatarVideoContents(userID uint, page int, limit int) ([]models.AvatarVideo, error) {
	var avatarVideoContents []models.AvatarVideo
	if err := s.DB.
		Model(&models.AvatarVideo{}).
		Preload("User").
		Preload("Avatar").
		Where("user_id = ?", userID).
		Limit(limit).
		Offset(page * limit).
		Order("created_at DESC").
		Find(&avatarVideoContents).Error; err != nil {
		return nil, err
	}
	return avatarVideoContents, nil
}

func (s *AvatarService) GetMyAvatarMusicContentsCount(userID uint) (int64, error) {
	var count int64
	if err := s.DB.Model(&models.AvatarMusic{}).
		Preload("User").
		Preload("Avatar").
		Where("user_id = ?", userID).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (s *AvatarService) GetMyAvatarVideoContentsCount(userID uint) (int64, error) {
	var count int64
	if err := s.DB.Model(&models.AvatarVideo{}).
		Preload("User").
		Preload("Avatar").
		Where("user_id = ?", userID).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
