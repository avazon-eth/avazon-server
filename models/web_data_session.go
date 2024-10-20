package models

import "time"

type WebDataSession struct {
	SessionID  string    `json:"-"`       // uuid
	UserID     string    `json:"-"`       // user's id
	Content    string    `json:"content"` // it must be shorter than 1024 bytes
	CreatedAt  time.Time `json:"created_at"`
	LastUsedAt time.Time `json:"last_used_at"`
}

// must be deleted whenever it's read
type WebTokenSession struct {
	TokenKey      string    `json:"token_key"`       // uuid
	DataSessionID string    `json:"data_session_id"` // uuid
	UserID        string    `json:"-"`               // user's id
	Token         string    `json:"token"`           // token
	CreatedAt     time.Time `json:"created_at"`
}
