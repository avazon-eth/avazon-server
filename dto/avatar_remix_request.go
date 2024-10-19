package dto

type AvatarImageRemixRequest struct {
	Prompt string `json:"prompt" binding:"required"`
}
