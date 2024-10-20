package models

import "time"

type UserRole string

const (
	UserRoleUser  UserRole = "user"
	UserRoleAdmin UserRole = "admin"
)

type User struct {
	ID              string    `json:"id" gorm:"primaryKey;type:varchar(255)"`
	Email           string    `json:"email" gorm:"not null;unique;type:varchar(255)"`
	Name            string    `json:"name" gorm:"not null;type:varchar(255)"`
	ProfileImageURL string    `json:"profile_image_url" gorm:"type:varchar(255)"`
	OAuth2Provider  string    `json:"oauth2_provider" gorm:"column:oauth2_provider;varchar(255)"` // google
	OAuth2ID        string    `json:"-" gorm:"column:oauth2_id;type:varchar(255)"`
	Role            UserRole  `json:"role" gorm:"not null;type:varchar(255)"`
	CreatedAt       time.Time `json:"created_at"`
	EditedAt        time.Time `json:"-" gorm:"autoUpdateTime"`
}
