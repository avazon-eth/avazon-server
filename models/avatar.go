package models

import "time"

type Avatar struct {
	ID               string         `json:"id" gorm:"primary_key;type:varchar(255)"`
	UserID           uint           `json:"user_id"`
	User             User           `json:"-" gorm:"foreignKey:UserID"`
	AvatarCreationID string         `json:"-" gorm:"type:varchar(255)"`
	AvatarCreation   AvatarCreation `json:"-" gorm:"foreignKey:AvatarCreationID"`
	RemixAvatarID    string         `json:"remix_avatar_id" gorm:"type:varchar(255)"`
	Name             string         `json:"name" gorm:"type:varchar(100)"`
	Species          string         `json:"species" gorm:"type:varchar(30)"`
	Gender           string         `json:"gender" gorm:"type:varchar(10)"`
	// Age                  int            `json:"age" gorm:"not null"`
	Language             string    `json:"language" gorm:"type:varchar(30)"`
	Country              string    `json:"country" gorm:"type:varchar(30)"`
	Description          string    `json:"description" gorm:"type:varchar(1000)"`
	CreatedAt            time.Time `json:"created_at"`
	ProfileImageURL      string    `json:"profile_image_url"`
	VoiceURL             string    `json:"voice_url"`
	AvatarVideoURL       *string   `json:"avatar_video_url"` // mock: used for test realtime chatting
	CharacterDescription string    `json:"character_description"`
}

type AvatarRemixImage struct {
	ID        int       `json:"id" gorm:"primary_key;auto_increment"`
	UserID    uint      `json:"user_id"`
	User      User      `json:"-" gorm:"foreignKey:UserID"`
	AvatarID  string    `json:"avatar_id" gorm:"not null;foreignKey:AvatarCreationID;constraint:OnDelete:CASCADE"`
	Avatar    Avatar    `json:"-" gorm:"foreignKey:AvatarCreationID;constraint:OnDelete:CASCADE"`
	Prompt    string    `json:"prompt" gorm:"type:varchar(3000)"`
	ImageURL  string    `json:"image_url" gorm:"not null"`
	CreatedAt time.Time `json:"created_at"`
}
