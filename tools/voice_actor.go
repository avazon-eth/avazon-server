package tools

import (
	"avazon-api/models"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

type VoiceActor interface {
	// Generate voice -> Return the model's provider and voice_id in order
	// returns (provider, voice_id, error)
	Create(prompt string, gender models.Gender, args ...string) (string, string, error)
	// Create voice
	TTS(voiceId string, text string) ([]byte, error)
	// Create TTS stream using voiceId
	TTSStream(voiceId string, text string) (io.ReadCloser, error)
}

type ElevenLabsVoiceActor struct {
	ApiKey string
}

type ElevenLabsVoiceGenerateReq struct {
	Accent         string        `json:"accent"`          // american, british, african, australian, indian
	AccentStrength float64       `json:"accent_strength"` // Must be between 0.3 and 2.0
	Age            string        `json:"age"`             // young, middle_aged, old
	Gender         models.Gender `json:"gender"`          // female, male
	Text           string        `json:"text"`            // prompt content (100 ~ 1000 characters)
}

type ElevenLabsVoiceCreateReq struct {
	VoiceName        string `json:"voice_name"`
	VoiceDescription string `json:"voice_description"`
	GeneratedVoiceID string `json:"generated_voice_id"`
}

type ElevenLabsVoiceCreateRes struct {
	VoiceID string `json:"voice_id"`
}

func NewElevenLabsVoiceActor(apiKey string) *ElevenLabsVoiceActor {
	return &ElevenLabsVoiceActor{
		ApiKey: apiKey,
	}
}

// Create method: Generate and save voice
// args[0]: accent_strength, args[1]: age, args[2]: accent
func (va *ElevenLabsVoiceActor) Create(prompt string, gender models.Gender, args ...string) (string, string, error) {
	// 1. Generate Voice
	generateURL := "https://api.elevenlabs.io/v1/voice-generation/generate-voice"

	accentStrength := 1.0
	age := "young"
	accent := "american"
	if len(args) > 0 {
		accentStrength, _ = strconv.ParseFloat(args[0], 64)
	}
	if len(args) > 1 {
		age = args[1]
	}
	if len(args) > 2 {
		accent = args[2]
	}

	// Create Generate Voice request
	generateReqData := ElevenLabsVoiceGenerateReq{
		Accent:         accent,
		AccentStrength: accentStrength,
		Age:            age,
		Gender:         gender,
		Text:           prompt,
	}
	jsonData, err := json.Marshal(generateReqData)
	if err != nil {
		fmt.Printf("failed to marshal generate request data: %v", err)
		return "", "", fmt.Errorf("failed to marshal generate request data: %v", err)
	}

	req, err := http.NewRequest("POST", generateURL, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("failed to create generate HTTP request: %v", err)
		return "", "", fmt.Errorf("failed to create generate HTTP request: %v", err)
	}
	req.Header.Set("XI-API-KEY", va.ApiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("failed to make generate API request: %v", err)
		return "", "", fmt.Errorf("failed to make generate API request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("unexpected response status: %v %s", resp.Status, body)
		return "", "", fmt.Errorf("unexpected response status: %v, body: %s", resp.Status, string(body))
	}

	// Retrieve generated_voice_id from response headers
	generatedVoiceID := resp.Header.Get("generated_voice_id")
	if generatedVoiceID == "" {
		fmt.Printf("generated_voice_id not found in response headers")
		return "", "", fmt.Errorf("generated_voice_id not found in response headers")
	}

	// 2. Create Voice
	createURL := "https://api.elevenlabs.io/v1/voice-generation/create-voice"

	saveReqData := ElevenLabsVoiceCreateReq{
		VoiceName:        generatedVoiceID,
		VoiceDescription: "Generated voice from API request",
		GeneratedVoiceID: generatedVoiceID,
	}

	jsonSaveData, err := json.Marshal(saveReqData)
	if err != nil {
		fmt.Printf("failed to marshal save request data: %v", err)
		return "", "", fmt.Errorf("failed to marshal save request data: %v", err)
	}

	saveReq, err := http.NewRequest("POST", createURL, bytes.NewBuffer(jsonSaveData))
	if err != nil {
		fmt.Printf("failed to create save HTTP request: %v", err)
		return "", "", fmt.Errorf("failed to create save HTTP request: %v", err)
	}

	saveReq.Header.Set("XI-API-KEY", va.ApiKey)
	saveReq.Header.Set("Content-Type", "application/json")

	saveResp, err := client.Do(saveReq)
	if err != nil {
		fmt.Printf("failed to make save API request: %v", err)
		return "", "", fmt.Errorf("failed to make save API request: %v", err)
	}
	defer saveResp.Body.Close()

	if saveResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(saveResp.Body)
		fmt.Printf("unexpected response status: %v %s", saveResp.Status, body)
		return "", "", fmt.Errorf("unexpected response status: %v, body: %s", saveResp.Status, string(body))
	}

	var saveRes ElevenLabsVoiceCreateRes
	if err := json.NewDecoder(saveResp.Body).Decode(&saveRes); err != nil {
		fmt.Printf("failed to decode save response: %v", err)
		return "", "", fmt.Errorf("failed to decode save response: %v", err)
	}

	return "elevenlabs", saveRes.VoiceID, nil
}

// TTS method: Convert text to speech and return binary data in MP3 format
func (va *ElevenLabsVoiceActor) TTS(voiceId string, text string) ([]byte, error) {
	ttsURL := fmt.Sprintf("https://api.elevenlabs.io/v1/text-to-speech/%s", voiceId)

	// Create TTS request
	ttsReqData := map[string]string{
		"text":     text,
		"model_id": "eleven_multilingual_v2",
	}

	jsonTTSData, err := json.Marshal(ttsReqData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal TTS request data: %v", err)
	}

	req, err := http.NewRequest("POST", ttsURL, bytes.NewBuffer(jsonTTSData))
	if err != nil {
		return nil, fmt.Errorf("failed to create TTS HTTP request: %v", err)
	}

	req.Header.Set("XI-API-KEY", va.ApiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make TTS API request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("unexpected response status: %v %s", resp.Status, body)
		return nil, fmt.Errorf("unexpected response status: %v, body: %s", resp.Status, string(body))
	}

	// Read voice data (binary data in MP3 format)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read TTS response body: %v", err)
	}

	// Return MP3 data to the caller
	return body, nil
}

// TTSStream method: Create TTS stream
func (va *ElevenLabsVoiceActor) TTSStream(voiceID string, text string) (io.ReadCloser, error) {
	ttsURL := fmt.Sprintf("https://api.elevenlabs.io/v1/text-to-speech/%s/stream", voiceID)

	// Create TTS request
	ttsReqData := map[string]string{
		"text":     text,
		"model_id": "eleven_multilingual_v2",
	}

	jsonTTSData, err := json.Marshal(ttsReqData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal TTS request data: %v", err)
	}

	req, err := http.NewRequest("POST", ttsURL, bytes.NewBuffer(jsonTTSData))
	if err != nil {
		return nil, fmt.Errorf("failed to create TTS HTTP request: %v", err)
	}

	req.Header.Set("XI-API-KEY", va.ApiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make TTS API request: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("unexpected response status: %v %s", resp.Status, body)
		return nil, fmt.Errorf("unexpected response status: %v, body: %s", resp.Status, string(body))
	}

	return resp.Body, nil
}
