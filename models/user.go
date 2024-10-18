package models

import "time"

type UserRole string

const (
	UserRoleUser  UserRole = "user"
	UserRoleAdmin UserRole = "admin"
)

type User struct {
	ID              uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	Username        string    `json:"username" gorm:"not null"`
	Email           string    `json:"email" gorm:"not null;unique"`
	ProfileImageURL string    `json:"profile_image_url"`
	Password        string    `json:"-"`                                             // not null if not using oauth2
	OAuth2Provider  string    `json:"oauth2_provider" gorm:"column:oauth2_provider"` // google
	OAuth2ID        string    `json:"-" gorm:"column:oauth2_id"`
	Role            UserRole  `json:"role" gorm:"not null"`
	CreatedAt       time.Time `json:"created_at"`
	EditedAt        time.Time `json:"-" gorm:"autoUpdateTime"`
}
