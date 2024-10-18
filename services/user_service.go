package services

import (
	"avazon-api/controllers/errs"
	"avazon-api/models"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"gorm.io/gorm"
)

type UserService struct {
	DB *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{DB: db}
}

// gets user info by google access token. If not exists, create new user and return user info.
func (s *UserService) GetUserByGoogleAccessToken(accessToken string) (*models.User, error) {
	// Google OAuth2 API endpoint
	googleAPIURL := "https://www.googleapis.com/oauth2/v3/userinfo"

	// Create HTTP client
	client := &http.Client{}

	// Send request to Google API
	req, err := http.NewRequest("GET", googleAPIURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+accessToken)

	// Get response
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Parse JSON
	var userInfo map[string]interface{}
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, err
	}

	// Check user info and generate token
	if email, ok := userInfo["email"].(string); ok {
		user := models.User{}
		s.DB.Where("oauth2_provider = ? AND oauth2_id = ?", "google", userInfo["sub"]).First(&user)
		if user.ID == 0 { // if not exists, create new user
			user = models.User{
				Username:        userInfo["name"].(string),
				Email:           email,
				ProfileImageURL: userInfo["picture"].(string),
				OAuth2Provider:  "google",
				OAuth2ID:        userInfo["sub"].(string),
				Role:            "user",
			}
			s.DB.Create(&user)
		}
		return &user, nil
	} else {
		log.Printf("we cannot find user info from google api: %v", userInfo)
		return nil, errs.ErrInternalServerError
	}
}

func (s *UserService) GetUserByID(userID uint) (*models.User, error) {
	var user models.User
	if err := s.DB.First(&user, userID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
