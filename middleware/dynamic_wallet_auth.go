package middleware

import (
	"avazon-api/models"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

var dynamicWalletKey *rsa.PublicKey

// Initialize RSA keys (should be called during app initialization)
func InitDynamicWalletKey() error {
	// Load the public key
	dwBytes, err := os.ReadFile("dw_key.pem")
	if err != nil {
		return fmt.Errorf("could not read public key file: %v", err)
	}

	dynamicWalletKey, err = jwt.ParseRSAPublicKeyFromPEM(dwBytes)
	if err != nil {
		return fmt.Errorf("could not parse public key: %v", err)
	}

	return nil
}

func ValidateDynamicWalletJWT(tokenString string) (*jwt.Token, error) {
	// Ensure the keys are loaded
	if dynamicWalletKey == nil {
		return nil, fmt.Errorf("public key is not initialized")
	}

	// Parse and validate the JWT token
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Check if the token signing method is RS256
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			if envID, ok := claims["environment_id"].(string); !ok || envID != os.Getenv("DW_ENVIRONMENT") {
				return nil, fmt.Errorf("invalid environment_id: %v", envID)
			}
		}

		// Validate the audience (aud) claim
		// if claims, ok := token.Claims.(jwt.MapClaims); ok {
		// 	if aud, ok := claims["aud"]; !ok || aud != "http://localhost:5173" {
		// 		return nil, fmt.Errorf("invalid audience: %v", aud)
		// 	}
		// }

		return dynamicWalletKey, nil
	})
}

func ExtractJWT(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return publicKey, nil
	})

	if err != nil && !errors.Is(err, jwt.ErrTokenExpired) {
		return nil, err
	}
	return token, nil
}

func ExtractPayload(jwtToken string) (map[string]interface{}, error) {
	parts := strings.Split(jwtToken, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid JWT token format")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("error decoding payload: %v", err)
	}

	var payloadData map[string]interface{}
	if err := json.Unmarshal(payload, &payloadData); err != nil {
		return nil, fmt.Errorf("error unmarshalling payload: %v", err)
	}

	return payloadData, nil
}

// Extract userId from JWT token string
func GetUserIDFromTokenString(tokenString string) (string, error) {
	token, err := ValidateDynamicWalletJWT(tokenString)
	if err != nil {
		return "", err
	}
	userId, err := GetUserIDFromJWT(*token)
	if err != nil {
		return "", err
	}
	return userId, nil
}

func GetUserRoleFromTokenString(tokenString string) (*models.UserRole, error) {
	token, err := ValidateDynamicWalletJWT(tokenString)
	if err != nil {
		return nil, err
	}
	role, ok := token.Claims.(jwt.MapClaims)["scope"]
	if !ok {
		return nil, fmt.Errorf("invalid token: role not found")
	}
	roleString, ok := role.(string)
	if !ok {
		return nil, fmt.Errorf("invalid token: role not found")
	}
	userRole := models.UserRole(roleString)
	return &userRole, nil
}

func GetUserIDFromJWT(token jwt.Token) (string, error) {
	// Extract claims from the token
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userId, ok := claims["sub"].(string)
		if !ok {
			return "", fmt.Errorf("invalid token: userId not found")
		}
		return userId, nil
	}

	return "", fmt.Errorf("invalid token")
}
