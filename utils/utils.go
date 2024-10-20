package utils

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/exp/rand"
)

// return 0: not exists, 1: not uint, other: uint
func GetUserID(c *gin.Context) (string, bool) {
	// Assuming the user_id is stored in the request header or query parameter
	userId, exists := c.Get("user_id")
	if !exists {
		return "", false
	}
	userIdStr, ok := userId.(string)
	if !ok {
		return "", false
	}
	return userIdStr, true
}

func GetUintParam(c *gin.Context, param string) (uint, error) {
	paramStr := c.Param(param)
	paramUint, err := strconv.ParseUint(paramStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse %s parameter: %v", param, err)
	}
	return uint(paramUint), nil
}

func GetExtensionFromMimeType(mimeType string) (string, error) {
	// Extract file extension from MIME type
	exts, err := mime.ExtensionsByType(mimeType)
	if err != nil || len(exts) == 0 {
		return "", fmt.Errorf("failed to get file extension for MIME type: %v", mimeType)
	}
	// Select the first extension
	ext := exts[0]
	// Add a dot if the extension does not start with one
	if ext[0] != '.' {
		ext = "." + ext
	}
	if ext == ".jpe" {
		return ".jpg", nil
	}
	return ext, nil
}

// Fetches a remote file using an HTTP GET request
// -> content, contentType(mime-type), error
func GetDataFromURL(url string) ([]byte, string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, "", err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		return nil, "", fmt.Errorf("failed to get content type from response")
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	return content, contentType, nil
}

func SaveBytesToFile(filename string, data []byte) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(data)
	if err != nil {
		return err
	}

	return nil
}

func GetLocalFile(filename string) ([]byte, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return io.ReadAll(file)
}

func DeleteFile(filename string) error {
	return os.Remove(filename)
}

// GenerateBirthBetween generates a random birth date between the given age range.
func GenerateBirthBetween(ageFrom int, ageTo int) time.Time {
	if ageFrom > ageTo {
		ageFrom, ageTo = ageTo, ageFrom
	}

	// Based on the current time
	now := time.Now()

	// Calculate minimum and maximum birth dates
	minBirthDate := now.AddDate(-ageTo, 0, 0)   // Birth date for maximum age
	maxBirthDate := now.AddDate(-ageFrom, 0, 0) // Birth date for minimum age

	// Calculate the number of days between the minimum and maximum birth dates
	days := int(maxBirthDate.Sub(minBirthDate).Hours() / 24)

	// Generate a random birth date by adding random days
	randomDays := rand.Intn(days + 1) // days + 1 to include the max day
	randomBirthDate := minBirthDate.AddDate(0, 0, randomDays)

	return randomBirthDate
}

func GenUUID(prefix string) (string, error) {
	u, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}
	uuidStr := fmt.Sprintf("%s%s", prefix, u.String())
	return uuidStr, nil
}

// Separate non-JSON and JSON parts
func SeparateTextAndJSON(text string) (string, []string) {
	jsonParts := []string{}
	nonJsonPart := ""
	start := 0
	braceCount := 0

	for i, char := range text {
		if char == '{' {
			if braceCount == 0 {
				nonJsonPart += text[start:i]
			}
			start = i
			braceCount++
		} else if char == '}' {
			braceCount--
			if braceCount == 0 && start != -1 {
				jsonParts = append(jsonParts, text[start:i+1])
				start = i + 1
			}
		}
	}

	if len(jsonParts) == 0 {
		return text, []string{}
	}
	return nonJsonPart, jsonParts
}
