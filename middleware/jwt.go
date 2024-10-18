package middleware

import (
	"crypto/rsa"
	"fmt"
	"os"
	"time"

	"avazon-api/models"

	"github.com/golang-jwt/jwt/v5"
)

// Load private and public keys from PEM files
var (
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
)

// Initialize RSA keys (should be called during app initialization)
func InitKeys() error {
	// Load the private key
	privateKeyBytes, err := os.ReadFile("private_key.pem")
	if err != nil {
		return fmt.Errorf("could not read private key file: %v", err)
	}

	privateKey, err = jwt.ParseRSAPrivateKeyFromPEM(privateKeyBytes)
	if err != nil {
		return fmt.Errorf("could not parse private key: %v", err)
	}

	// Load the public key
	publicKeyBytes, err := os.ReadFile("public_key.pem")
	if err != nil {
		return fmt.Errorf("could not read public key file: %v", err)
	}

	publicKey, err = jwt.ParseRSAPublicKeyFromPEM(publicKeyBytes)
	if err != nil {
		return fmt.Errorf("could not parse public key: %v", err)
	}

	return nil
}

// GenerateJWT generates a JWT token using RSA (RS256)
//
// Parameters:
//   - userId uint: The unique identifier of the user
//   - tokenType string: The type of the token (e.g., "access", "refresh")
//   - expireMin int: The expiration time of the token in minutes
//
// Returns:
//   - string: The generated JWT token string
//   - error: An error object if any error occurs during token generation
func GenerateJWT(userId uint, tokenType string, expireMin int, scope string) (string, error) {
	// Ensure the keys are loaded
	if privateKey == nil {
		return "", fmt.Errorf("private key is not initialized")
	}

	// Create the token claims
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"sub":   userId,
		"exp":   time.Now().Add(time.Minute * time.Duration(expireMin)).Unix(),
		"type":  tokenType,
		"scope": scope,
	})

	// Sign the token using the RSA private key
	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateJWT validates a JWT token using the RSA public key
//
// Parameters:
//   - tokenString string: The JWT token string
//
// Returns:
//   - *jwt.Token: The parsed JWT token
//   - error: An error object if the token is invalid or expired
func ValidateJWT(tokenString string) (*jwt.Token, error) {
	// Ensure the keys are loaded
	if publicKey == nil {
		return nil, fmt.Errorf("public key is not initialized")
	}

	// Parse and validate the JWT token
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Check if the token signing method is RS256
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return publicKey, nil
	})
}

// Extract userId from JWT token string
func GetUserIDFromTokenString(tokenString string) (uint, error) {
	token, err := ValidateJWT(tokenString)
	if err != nil {
		return 0, err
	}
	userId, err := GetUserIDFromJWT(*token)
	if err != nil {
		return 0, err
	}
	return userId, nil
}

func GetUserRoleFromTokenString(tokenString string) (*models.UserRole, error) {
	token, err := ValidateJWT(tokenString)
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

func GetUserIDFromJWT(token jwt.Token) (uint, error) {
	// Extract claims from the token
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userId, ok := claims["sub"].(float64)
		if !ok {
			return 0, fmt.Errorf("invalid token: userId not found")
		}
		return uint(userId), nil
	}

	return 0, fmt.Errorf("invalid token")
}
