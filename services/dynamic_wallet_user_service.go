package services

import (
	"avazon-api/controllers/errs"
	"avazon-api/models"
	"encoding/json"
	"fmt"
	"net/http"
)

type DynamicWalletUserService struct {
	APIKey string
}

func NewDynamicWalletUserService(apiKey string) *DynamicWalletUserService {
	return &DynamicWalletUserService{APIKey: apiKey}
}

func (s *DynamicWalletUserService) GetDWUserByID(id string) (*models.User, error) {
	// GET https://app.dynamicauth.com/api/v0/users/{userId}
	url := fmt.Sprintf("https://app.dynamicauth.com/api/v0/users/%s", id)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.APIKey))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get user: %s", resp.Status)
	}

	// requires email, name, profile_image
	// Google only for now
	var email string
	var name string
	var profileImageURL string
	// start from json["user"][...]
	var jsonData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&jsonData); err != nil {
		return nil, err
	}
	var userData map[string]interface{}
	userData, ok := jsonData["user"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("user data is not a valid map")
	}
	var verifiedCredentialsData []interface{}
	verifiedCredentialsData, ok = userData["verifiedCredentials"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("verified credentials data is not a valid slice")
	}
	for _, credential := range verifiedCredentialsData {
		credentialMap, ok := credential.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("credential is not a valid map")
		}
		if credentialMap["format"] == "oauth" && credentialMap["oauth_provider"] == "google" {
			email = credentialMap["oauth_username"].(string)
			name = credentialMap["oauth_display_name"].(string)
			photos := credentialMap["oauth_account_photos"].([]interface{})
			if len(photos) > 0 {
				photoURL, ok := photos[0].(string)
				if ok {
					profileImageURL = photoURL
				}
			}
			user := models.User{
				ID:              id,
				OAuth2Provider:  "google",
				OAuth2ID:        credentialMap["oauth_account_id"].(string),
				Email:           email,
				Name:            name,
				ProfileImageURL: profileImageURL,
			}
			return &user, nil
		}
	}

	return nil, errs.ErrUnauthorized
}
