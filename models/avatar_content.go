package models

import "time"

type AvatarMusic struct {
	ID            string    `json:"id" gorm:"primaryKey;varchar(40)"`
	UserID        string    `json:"user_id" gorm:"type:varchar(255);not null"`
	User          User      `json:"user" gorm:"foreignKey:UserID;"`
	Title         string    `json:"title" gorm:"varchar(255);not null"`
	AvatarID      string    `json:"avatar_id" gorm:"varchar(40);not null"`
	Avatar        Avatar    `json:"avatar" gorm:"foreignKey:AvatarID;"`
	AlbumImageURL string    `json:"album_image_url" gorm:"varchar(255);not null"`
	MusicURL      string    `json:"music_url" gorm:"varchar(255);not null"`
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type AvatarVideo struct {
	ID                string    `json:"id" gorm:"primaryKey;varchar(40)"`
	UserID            string    `json:"user_id" gorm:"type:varchar(255);not null"`
	User              User      `json:"user" gorm:"foreignKey:UserID;"`
	Title             string    `json:"title" gorm:"varchar(255);not null"`
	AvatarID          string    `json:"avatar_id" gorm:"varchar(40);not null"`
	Avatar            Avatar    `json:"avatar" gorm:"foreignKey:AvatarID;"`
	ThumbnailImageURL string    `json:"thumbnail_image_url" gorm:"varchar(255);not null"`
	VideoContentURL   string    `json:"video_content_url" gorm:"varchar(255);not null"`
	CreatedAt         time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt         time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
