package models

import "time"

type SystemPrompt struct {
	ID                string              `json:"id" gorm:"primaryKey;varchar(100)"` // name of the prompt
	SystemPromptUsage []SystemPromptUsage `json:"-" gorm:"foreignKey:PromptID;OnDelete:set null"`
	Prompt            string              `json:"prompt" binding:"required" gorm:"not null;size:10000"`
	UsedFor           string              `json:"used_for"`
	Category          string              `json:"category"`
	CreatedAt         time.Time           `json:"created_at"`
	EditedAt          time.Time           `json:"edited_at" gorm:"autoUpdateTime"`
}

// SystemPromptUsage is a model that represents the usage of a system prompt by an agent
// for example) {"agent_id": "create_idol", "prompt_id": "create_idol_v1"}
type SystemPromptUsage struct {
	AgentID   string       `json:"agent_id" gorm:"primaryKey;not null"`
	PromptID  string       `json:"prompt_id" binding:"omitempty"`
	Prompt    SystemPrompt `json:"prompt" gorm:"foreignKey:PromptID;references:ID" binding:"omitempty"`
	CreatedAt time.Time    `json:"created_at"`
	EditedAt  time.Time    `json:"edited_at" gorm:"autoUpdateTime"`
}
