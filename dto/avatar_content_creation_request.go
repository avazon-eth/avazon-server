package dto

import "fmt"

type AvatarVideoImageRequest struct {
	Prompt string `json:"prompt" binding:"required"`
}

type AvatarVideoRequest struct {
	Title  string `json:"title" binding:"required"`
	Prompt string `json:"prompt" binding:"required"`
}

type AvatarMusicRequest struct {
	Title       string `json:"title" binding:"required"`
	Duration    int    `json:"duration" binding:"required"` // 20s or 45s
	Style       string `json:"style"`
	Description string `json:"description"`
}

func (req AvatarMusicRequest) GetMusicInfo() string {
	return fmt.Sprintf("Title: %s\nDuration: %d\nStyle: %s\nDescription: %s", req.Title, req.Duration, req.Style, req.Description)
}
