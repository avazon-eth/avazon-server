package models

type AvatarRemixStatus string

const (
	AR_Yet         AvatarRemixStatus = "yet"
	AR_Progressing AvatarRemixStatus = "progressing"
	AR_Completed   AvatarRemixStatus = "completed"
	AR_Confirmed   AvatarRemixStatus = "confirmed"
	AR_Failed      AvatarRemixStatus = "failed"
)

type AvatarImageRemix struct {
	ID           string            `json:"id" gorm:"primary_key;type:varchar(36);not null"`
	UserID       string            `json:"user_id" gorm:"type:varchar(255);not null"`
	User         User              `json:"user" gorm:"foreignKey:UserID"`
	AvatarID     string            `json:"avatar_id" gorm:"type:varchar(255);not null"`
	Avatar       Avatar            `json:"avatar" gorm:"foreignKey:AvatarID"`
	UserPrompt   string            `json:"user_prompt" gorm:"type:varchar(1000);not null"`
	Status       AvatarRemixStatus `json:"status" gorm:"type:varchar(20);not null"`
	FailedReason *string           `json:"failed_reason" gorm:"type:varchar(255);"`
	ImageURL     *string           `json:"image_url" gorm:"type:varchar(255);"`
}
