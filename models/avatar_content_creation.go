package models

import "time"

type AvatarContentCreationStatus string

// yet -> image_progressing -> image_completed
// -> content_progressing -> content_completed -> confirmed
// if error occurs, status -> failed
const (
	ACC_Yet                AvatarContentCreationStatus = "yet"
	ACC_ImageProgressing   AvatarContentCreationStatus = "image_progressing"
	ACC_ImageCompleted     AvatarContentCreationStatus = "image_completed"
	ACC_ContentProgressing AvatarContentCreationStatus = "content_progressing"
	ACC_ContentCompleted   AvatarContentCreationStatus = "content_completed"
	ACC_Confirmed          AvatarContentCreationStatus = "confirmed"
	ACC_Failed             AvatarContentCreationStatus = "failed"
)

type AvatarMusicContentCreation struct {
	ID                   string                      `json:"id" gorm:"primaryKey;varchar(36)"` // UUID
	UserID               uint                        `json:"user_id"`
	User                 User                        `json:"-" gorm:"foreignKey:UserID"`
	AvatarID             string                      `json:"avatar_id" gorm:"varchar(36);not null"`
	Avatar               Avatar                      `json:"avatar" gorm:"foreignKey:AvatarID;"`
	Title                string                      `json:"title" gorm:"type:varchar(255)"`
	Style                string                      `json:"style" gorm:"type:varchar(300)"`
	Description          string                      `json:"description" gorm:"type:varchar(3000)"`
	GeneratedImagePrompt *string                     `json:"generated_image_prompt" gorm:"type:text"`
	GeneratedMusicPrompt *string                     `json:"generated_music_prompt" gorm:"type:text"`
	AlbumImageURL        *string                     `json:"album_image_url" gorm:"type:varchar(255)"`
	MusicURL             *string                     `json:"music_url" gorm:"type:varchar(255)"`
	Status               AvatarContentCreationStatus `json:"status" gorm:"varchar(20);not null"`
	FailedReason         *string                     `json:"failed_reason" gorm:"type:varchar(255)"`
	CreatedAt            time.Time                   `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt            time.Time                   `json:"updated_at" gorm:"autoUpdateTime"`
}

type AvatarVideoContentCreation struct {
	ID       string `json:"id" gorm:"primaryKey;varchar(36)"` // UUID
	UserID   uint   `json:"user_id"`
	User     User   `json:"-" gorm:"foreignKey:UserID"`
	AvatarID string `json:"avatar_id" gorm:"varchar(36);not null"`
	Avatar   Avatar `json:"avatar" gorm:"foreignKey:AvatarID;"`
	// step 1: image prompt
	ImagePrompt       string  `json:"image_prompt" gorm:"varchar(255);not null"`
	ThumbnailImageURL *string `json:"thumbnail_image_url" gorm:"varchar(255);"`
	// step 2: content prompt
	VideoPrompt     string                      `json:"video_prompt" gorm:"varchar(255);not null"`
	VideoContentURL *string                     `json:"video_content_url" gorm:"varchar(255);"`
	Status          AvatarContentCreationStatus `json:"status" gorm:"varchar(20);not null"`
	FailedReason    *string                     `json:"failed_reason" gorm:"type:varchar(255)"`
	CreatedAt       time.Time                   `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time                   `json:"updated_at" gorm:"autoUpdateTime"`
}
