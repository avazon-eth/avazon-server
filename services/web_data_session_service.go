package services

import (
	"avazon-api/controllers/errs"
	"avazon-api/models"
	"time"

	"github.com/google/uuid"
)

// web browser data transaction between different domains
type WebDataSessionService struct {
	data      map[string]*models.WebDataSession  // key: session_id, value: session
	tokenData map[string]*models.WebTokenSession // key: session_id, value: session
}

func NewWebDataSessionService() *WebDataSessionService {
	return &WebDataSessionService{
		data:      make(map[string]*models.WebDataSession),
		tokenData: make(map[string]*models.WebTokenSession),
	}
}

// returns session_id
func (s *WebDataSessionService) PutData(userID string, sessionID string, content string) error {
	session, ok := s.data[sessionID]
	if ok && session.UserID == userID {
		session.Content = content
		session.LastUsedAt = time.Now()
		return nil
	}
	return errs.ErrNotFound
}

func (s *WebDataSessionService) GetData(sessionID string, userID string) (string, error) {
	session, ok := s.data[sessionID]
	if !ok {
		return "", errs.ErrNotFound
	}
	if session.UserID != userID {
		return "", errs.ErrUnauthorized
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

func (s *WebDataSessionService) PutToken(userID string, token string) (models.WebTokenSession, error) {
	session := &models.WebTokenSession{
		TokenKey:      uuid.New().String(),
		DataSessionID: uuid.New().String(),
		UserID:        userID,
		Token:         token,
		CreatedAt:     time.Now(),
	}

	s.tokenData[session.TokenKey] = session
	s.data[session.DataSessionID] = &models.WebDataSession{
		SessionID: session.DataSessionID,
		UserID:    userID,
		CreatedAt: time.Now(),
		Content:   "",
	}

	return *session, nil
}

func (s *WebDataSessionService) GetToken(tokenKey string) (models.WebTokenSession, error) {
	session, ok := s.tokenData[tokenKey]
	if !ok {
		return models.WebTokenSession{}, errs.ErrNotFound
	}
	sessionCopy := *session
	delete(s.tokenData, tokenKey)
	return sessionCopy, nil
}

func (s *WebDataSessionService) ClearToken(tokenKey string) error {
	_, ok := s.tokenData[tokenKey]
	if !ok {
		return errs.ErrNotFound
	}
	delete(s.tokenData, tokenKey)
	return nil
}
