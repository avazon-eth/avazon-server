package services

import (
	"avazon-api/controllers/errs"
	"time"

	"github.com/google/uuid"
)

// web browser data transaction between different domains
type WebDataSessionService struct {
	data      map[string]*WebDataSession  // key: session_id, value: session
	tokenData map[string]*WebTokenSession // key: session_id, value: session
}

type WebDataSession struct {
	SessionID  string // uuid
	UserID     string // user's id
	Content    string // it must be shorter than 1024 bytes
	CreatedAt  time.Time
	LastUsedAt time.Time
}

// must be deleted whenever it's read
type WebTokenSession struct {
	SessionID string // uuid
	UserID    string // user's id
	Token     string // token
	CreatedAt time.Time
}

func NewWebDataSessionService() *WebDataSessionService {
	return &WebDataSessionService{
		data: make(map[string]*WebDataSession),
	}
}

// returns session_id
func (s *WebDataSessionService) PutData(userID string, content string) (string, error) {
	sessionID := uuid.New().String()

	session := &WebDataSession{
		SessionID: sessionID,
		UserID:    userID,
		Content:   content,
		CreatedAt: time.Now(),
	}

	s.data[sessionID] = session

	return sessionID, nil
}

func (s *WebDataSessionService) GetData(sessionID string, userID string) (string, error) {
	session, ok := s.data[sessionID]
	if !ok {
		return "", errs.ErrNotFound
	}
	return session.Content, nil
}

func (s *WebDataSessionService) ClearData(sessionID string, userID string) error {
	_, ok := s.data[sessionID]
	if !ok {
		return errs.ErrNotFound
	}
	delete(s.data, sessionID)
	return nil
}

func (s *WebDataSessionService) PutToken(userID string, token string) (string, error) {
	sessionID := uuid.New().String()

	session := &WebTokenSession{
		SessionID: sessionID,
		UserID:    userID,
		Token:     token,
		CreatedAt: time.Now(),
	}

	s.tokenData[sessionID] = session

	return sessionID, nil
}

func (s *WebDataSessionService) GetToken(sessionID string, userID string) (string, error) {
	session, ok := s.tokenData[sessionID]
	if !ok {
		return "", errs.ErrNotFound
	}
	return session.Token, nil
}

func (s *WebDataSessionService) ClearToken(sessionID string, userID string) error {
	_, ok := s.tokenData[sessionID]
	if !ok {
		return errs.ErrNotFound
	}
	delete(s.tokenData, sessionID)
	return nil
}
