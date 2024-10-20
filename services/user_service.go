package services

import (
	"avazon-api/models"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"gorm.io/gorm"
)

type UserService struct {
	DB *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{DB: db}
}
func (s *UserService) GetUserGoogleAccessTokenByAuthorizationCode(code string) (string, error) {
	// Google OAuth2 token endpoint
	tokenURL := "https://oauth2.googleapis.com/token"

	// Prepare the request body
	reqBody := fmt.Sprintf("code=%s&client_id=%s&client_secret=%s&redirect_uri=%s&grant_type=authorization_code",
		code,
		os.Getenv("GOOGLE_CLIENT_ID"),
		os.Getenv("GOOGLE_CLIENT_SECRET"),
		os.Getenv("GOOGLE_REDIRECT_URI"),
	)

	// Create HTTP client
	client := &http.Client{}

	// Create a new POST request
	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(reqBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Check for errors
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get access token: %s", resp.Status)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Parse the JSON response
	var tokenResponse struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(body, &tokenResponse); err != nil {
		return "", err
	}

	return tokenResponse.AccessToken, nil
}

// gets user info by google access token. If not exists, create new user and return user info.
// func (s *UserService) GetUserByGoogleAccessToken(accessToken string) (*models.User, error) {
// 	// Google OAuth2 API endpoint
// 	googleAPIURL := "https://www.googleapis.com/oauth2/v3/userinfo"

// 	// Create HTTP client
// 	client := &http.Client{}

// 	// Send request to Google API
// 	req, err := http.NewRequest("GET", googleAPIURL, nil)
// 	if err != nil {
// 		return nil, err
// 	}
// 	req.Header.Add("Authorization", "Bearer "+accessToken)

// 	// Get response
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode >= 400 && resp.StatusCode < 500 {
// 		return nil, errs.ErrOAuthTokenInvalid
// 	}

// 	// Read response body
// 	body, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Parse JSON
// 	var userInfo map[string]interface{}
// 	if err := json.Unmarshal(body, &userInfo); err != nil {
// 		return nil, err
// 	}

// 	// Check user info and generate token
// 	if email, ok := userInfo["email"].(string); ok {
// 		user := models.User{}
// 		s.DB.Where("oauth2_provider = ? AND oauth2_id = ?", "google", userInfo["sub"]).First(&user)
// 		if user.ID == "" {
// 			user = models.User{
// 				Username:        userInfo["name"].(string),
// 				Email:           email,
// 				ProfileImageURL: userInfo["picture"].(string),
// 				OAuth2Provider:  "google",
// 				OAuth2ID:        userInfo["sub"].(string),
// 				Role:            "user",
// 			}
// 			s.DB.Create(&user)
// 		}
// 		return &user, nil
// 	} else {
// 		log.Printf("we cannot find user info from google api: %v", userInfo)
// 		return nil, errs.ErrInternalServerError
// 	}
// }

func (s *UserService) GetUserByID(userID string) (*models.User, error) {
	var user models.User
	if err := s.DB.Where("id = ?", userID).Find(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *UserService) GetUserByIDCreateIfNotExists(userID string) (*models.User, error) {
	var user models.User
	if err := s.DB.
		Where("id = ?", userID).
		FirstOrCreate(&user, models.User{ID: userID, Role: "user"}).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
