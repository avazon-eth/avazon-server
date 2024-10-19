package tools

import (
	"avazon-api/utils"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
)

type Painter interface {
	// returns (image_bytes, mime_type, error)
	Paint(prompt string, negative string, width int, height int) ([]byte, string, error)
	// paint from reference image
	// :param refImageBytes: reference image bytes (e.g. face image)
	// :param refContentType: reference image MIME type (e.g. image/jpeg)
	PaintFromReference(
		refImageBytes []byte, refContentType string, prompt string, width int, height int,
	) ([]byte, string, error)
	// enhance prompt
	EnhancePrompt(prompt string) (string, error)
	// Change style of the image
	ChangeStyle(imageBytes []byte, contentType string, prompt string) ([]byte, string, error)
}

type OpenArtRequest struct {
	Prompt              *string  `json:"prompt,omitempty"` // Changed to pointer to allow null
	IsGeneratedPrompt   bool     `json:"isGeneratedPrompt"`
	ImageNum            int      `json:"image_num"`
	Width               int      `json:"width"`
	Height              int      `json:"height"`
	Steps               int      `json:"steps"`
	CfgScale            float32  `json:"cfg_scale"`
	Seed                *string  `json:"seed,omitempty"`                  // Changed to pointer to allow null
	PromptAssistantMode *string  `json:"prompt_assistant_mode,omitempty"` // Changed to pointer to allow null
	BaseModel           *string  `json:"base_model,omitempty"`            // Changed to pointer to allow null
	Model               *string  `json:"model,omitempty"`                 // Changed to pointer to allow null
	NegativePrompt      *string  `json:"negative_prompt,omitempty"`       // Changed to pointer to allow null
	Tiling              bool     `json:"tiling"`
	Sampler             *string  `json:"sampler,omitempty"`   // Changed to pointer to allow null
	Tab                 *string  `json:"tab,omitempty"`       // Changed to pointer to allow null
	AiModel             *string  `json:"ai_model,omitempty"`  // Changed to pointer to allow null
	Mode                *string  `json:"mode,omitempty"`      // Changed to pointer to allow null
	ImageURL            *string  `json:"image_url,omitempty"` // Changed to pointer to allow null
	Strength            *float64 `json:"strength,omitempty"`  // Changed to pointer to allow null
}

type OpenArtResponse struct {
	InferenceRequestID     string `json:"inference_request_id"`
	GenerationHistoryID    string `json:"generation_history_id"`
	IsReachGenerationLimit bool   `json:"is_reach_generation_limit"`
}

type OpenArtFluxResponse struct {
	GenerationHistoryIDs []string `json:"generation_history_ids"`
	MessageBody          struct {
		Steps    int    `json:"steps"`
		Prompt   string `json:"prompt"`
		Type     string `json:"type"`
		ImageNum int    `json:"image_num"`
		UserID   string `json:"user_id"`
	} `json:"message_body"`
	IsReachGenerationLimit bool `json:"is_reach_generation_limit"`
}

type ImageItem struct {
	Seed        int    `json:"seed"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	URL         string `json:"url"`
	UpscaleMode string `json:"upscale_mode"`
	Status      string `json:"status"`
}

type ImagePlaceholderResponse struct {
	Images []ImageItem `json:"images"`
}

type OpenArtPainter struct {
	ApiKey string
}

type OpenAIPainter struct {
	APIKey string
}

func NewOpenArtPainter(apiKey string) *OpenArtPainter {
	return &OpenArtPainter{ApiKey: apiKey}
}

func NewOpenAIPainter(apiKey string) *OpenAIPainter {
	return &OpenAIPainter{APIKey: apiKey}
}

// ======================================================================================================================
// OpenAIArtist
// ======================================================================================================================

func (a *OpenAIPainter) PaintFromReference(
	refImageBytes []byte, refContentType string, prompt string, width int, height int,
) ([]byte, string, error) {
	return nil, "", fmt.Errorf("not implemented")
}

// DALL-E-3
func (a *OpenAIPainter) Paint(prompt string, _ string, width int, height int) ([]byte, string, error) {
	// Prepare request data
	requestData := map[string]interface{}{
		"model":  "dall-e-3",
		"prompt": prompt,
		"n":      1,
		"size":   fmt.Sprintf("%dx%d", width, height),
	}

	// Encode request data in JSON format
	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return nil, "", fmt.Errorf("failed to marshal request data: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/images/generations", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", a.APIKey))

	// Send request using HTTP client
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, "", fmt.Errorf("received non-200 response code: %d, body: %s", resp.StatusCode, string(body))
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse response
	var responseData struct {
		Data []struct {
			URL string `json:"url"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &responseData); err != nil {
		return nil, "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(responseData.Data) == 0 {
		return nil, "", fmt.Errorf("no image data received")
	}

	// Return the URL of the generated image
	imageURL := responseData.Data[0].URL

	// Download the image data and return as byte array
	imageResp, err := http.Get(imageURL)
	if err != nil {
		return nil, "", fmt.Errorf("failed to download image: %w", err)
	}
	defer imageResp.Body.Close()

	imageData, err := io.ReadAll(imageResp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read image data: %w", err)
	}

	return imageData, "image/png", nil
}

func (a *OpenAIPainter) EnhancePrompt(prompt string) (string, error) {
	return "", fmt.Errorf("OpenAIArtist.EnhancePrompt method not implemented")
}

func (a *OpenAIPainter) ChangeStyle(imageBytes []byte, contentType string, prompt string) ([]byte, string, error) {
	return nil, "", fmt.Errorf("OpenAIArtist.ChangeStyle method not implemented")
}

// ======================================================================================================================
// OpenArtArtist
// ======================================================================================================================

// Stable Diffusion
func (a *OpenArtPainter) Paint_old(prompt string, negative string, width int, height int) ([]byte, string, error) {
	// currentModel := "Merjic/majicMIX-realistic"
	currentModel := "DynamicWang/AWPortrait"
	// 1. Request image generation from OpenArt API
	requestData := OpenArtRequest{
		Prompt:              &prompt, // Changed to pointer to allow null
		IsGeneratedPrompt:   false,
		ImageNum:            1, // By default, generate only one image
		Width:               width,
		Height:              height,
		Steps:               25,
		CfgScale:            10,
		Seed:                nil,                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           // Changed to nil to allow null
		PromptAssistantMode: &[]string{"auto"}[0],                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          // Changed to pointer to allow null
		BaseModel:           &currentModel,                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 // Changed to pointer to allow null
		Model:               &[]string{"stable_diffusion"}[0],                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              // Changed to pointer to allow null
		NegativePrompt:      &[]string{"worst quality, bad quality, displeasing, very displeasing, lowres, bad anatomy, bad perspective, bad proportions, bad aspect ratio, bad face, long face, bad teeth, bad neck, long neck, bad arm, bad hands, bad ass, bad leg, bad feet, bad reflection, bad shadow, bad link, bad source, wrong hand, wrong feet, missing limb, missing eye, missing tooth, missing ear, missing finger, extra faces, extra eyes, extra eyebrows, extra mouth, extra tongue, extra teeth, extra ears, extra breasts, extra arms, extra hands, extra legs, extra digits, fewer digits, cropped head, cropped torso, cropped shoulders, cropped arms, cropped legs, mutation, deformed, disfigured, unfinished, chromatic aberration, text, error, jpeg artifacts, watermark, scan, scan artifacts, shadow, shade, shaded face"}[0], // Changed to pointer to allow null
		Tiling:              false,
		Sampler:             &[]string{"DPM++ 2M SDE Karras"}[0], // Changed to pointer to allow null
		Tab:                 &[]string{"generation"}[0],          // Changed to pointer to allow null
		AiModel:             &currentModel,                       // Changed to pointer to allow null
		Mode:                &[]string{"create"}[0],              // Changed to pointer to allow null
	}

	requestBody, err := json.Marshal(requestData)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create request payload: %v", err)
	}

	req, err := http.NewRequest("POST", "https://openart.ai/api/apps/create_model_hub", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, "", fmt.Errorf("failed to create HTTP request: %v", err)
	}

	// Set Cookie header
	req.Header.Set("Cookie", "__Secure-next-auth.session-token="+a.ApiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("failed to make API request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("unexpected response status: %v", resp.Status)
	}

	var openArtResp OpenArtResponse
	if err := json.NewDecoder(resp.Body).Decode(&openArtResp); err != nil {
		return nil, "", fmt.Errorf("failed to decode response: %v", err)
	}

	// 2. Periodically check the status of the generated image
	var finalImageURL string
	for {
		// Request to check the status of the image
		placeholderURL := fmt.Sprintf("https://openart.ai/api/create/image_placeholder?generation_history_id=%s", openArtResp.GenerationHistoryID)
		statusReq, err := http.NewRequest("GET", placeholderURL, nil)
		if err != nil {
			return nil, "", fmt.Errorf("failed to create status request: %v", err)
		}
		statusReq.Header.Set("Cookie", "__Secure-next-auth.session-token="+a.ApiKey)

		statusResp, err := client.Do(statusReq)
		if err != nil {
			return nil, "", fmt.Errorf("failed to check image status: %v", err)
		}
		defer statusResp.Body.Close()

		if statusResp.StatusCode != http.StatusOK {
			return nil, "", fmt.Errorf("unexpected status response: %v", statusResp.Status)
		}

		var imagePlaceholderResp ImagePlaceholderResponse
		if err := json.NewDecoder(statusResp.Body).Decode(&imagePlaceholderResp); err != nil {
			return nil, "", fmt.Errorf("failed to decode status response: %v", err)
		}

		// Check image status
		allCompleted := true
		for _, img := range imagePlaceholderResp.Images {
			if img.Status == "completed" {
				finalImageURL = img.URL // Save the URL of the first completed image
				break
			} else if img.Status == "pending" {
				allCompleted = false
			}
		}

		if allCompleted {
			break
		}

		// Wait briefly before checking status again
		time.Sleep(2 * time.Second)
	}

	if finalImageURL == "" {
		return nil, "", fmt.Errorf("no completed images found")
	}

	// Change to the original quality image URL
	finalImageURL = strings.Replace(finalImageURL, "_512.webp", "_raw.jpg", 1)

	println("finalImageURL:", finalImageURL)
	return fetchImageFromURL(finalImageURL)
}

// Stable Diffusion
func (a *OpenArtPainter) Paint(prompt string, negative string, width int, height int) ([]byte, string, error) {
	currentModel := "Flux_dev"
	// 1. Request image generation from OpenArt API
	requestData := OpenArtRequest{
		Prompt:              &prompt, // Changed to pointer to allow null
		IsGeneratedPrompt:   true,
		ImageNum:            1, // By default, generate only one image
		Width:               width,
		Height:              height,
		Steps:               28,
		CfgScale:            3.5,
		Seed:                nil,                 // Changed to nil to allow null
		PromptAssistantMode: &[]string{"off"}[0], // Changed to pointer to allow null
		BaseModel:           &currentModel,       // Changed to pointer to allow null
		// NegativePrompt:      &[]string{"worst quality, bad quality, displeasing, very displeasing, lowres, bad anatomy, bad perspective, bad proportions, bad aspect ratio, bad face, long face, bad teeth, bad neck, long neck, bad arm, bad hands, bad ass, bad leg, bad feet, bad reflection, bad shadow, bad link, bad source, wrong hand, wrong feet, missing limb, missing eye, missing tooth, missing ear, missing finger, extra faces, extra eyes, extra eyebrows, extra mouth, extra tongue, extra teeth, extra ears, extra breasts, extra arms, extra hands, extra legs, extra digits, fewer digits, cropped head, cropped torso, cropped shoulders, cropped arms, cropped legs, mutation, deformed, disfigured, unfinished, chromatic aberration, text, error, jpeg artifacts, watermark, scan, scan artifacts, shadow, shade, shaded face"}[0], // Changed to pointer to allow null
		Tiling:  false,
		Sampler: &[]string{"DPM++ 2M SDE Karras"}[0], // Changed to pointer to allow null
		AiModel: &currentModel,                       // Changed to pointer to allow null
	}

	requestBody, err := json.Marshal(requestData)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create request payload: %v", err)
	}

	req, err := http.NewRequest("POST", "https://openart.ai/api/create/flux", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, "", fmt.Errorf("failed to create HTTP request: %v", err)
	}

	// Set Cookie header
	req.Header.Set("Cookie", "__Secure-next-auth.session-token="+a.ApiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("failed to make API request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("unexpected response status: %v", resp.Status)
	}

	var openArtResp OpenArtFluxResponse
	if err := json.NewDecoder(resp.Body).Decode(&openArtResp); err != nil {
		return nil, "", fmt.Errorf("failed to decode response: %v", err)
	}

	if len(openArtResp.GenerationHistoryIDs) == 0 {
		return nil, "", fmt.Errorf("no generation history IDs found")
	}

	// 2. Periodically check the status of the generated image
	var finalImageURL string
	for {
		// Request to check the status of the image
		placeholderURL := fmt.Sprintf("https://openart.ai/api/create/image_placeholder?generation_history_id=%s", openArtResp.GenerationHistoryIDs[0])
		statusReq, err := http.NewRequest("GET", placeholderURL, nil)
		if err != nil {
			return nil, "", fmt.Errorf("failed to create status request: %v", err)
		}
		statusReq.Header.Set("Cookie", "__Secure-next-auth.session-token="+a.ApiKey)

		statusResp, err := client.Do(statusReq)
		if err != nil {
			return nil, "", fmt.Errorf("failed to check image status: %v", err)
		}
		defer statusResp.Body.Close()

		if statusResp.StatusCode != http.StatusOK {
			return nil, "", fmt.Errorf("unexpected status response: %v", statusResp.Status)
		}

		var imagePlaceholderResp ImagePlaceholderResponse
		if err := json.NewDecoder(statusResp.Body).Decode(&imagePlaceholderResp); err != nil {
			return nil, "", fmt.Errorf("failed to decode status response: %v", err)
		}

		// Check image status
		allCompleted := true
		for _, img := range imagePlaceholderResp.Images {
			if img.Status == "completed" {
				finalImageURL = img.URL // Save the URL of the first completed image
				break
			} else if img.Status == "pending" {
				allCompleted = false
			}
		}

		if allCompleted {
			break
		}

		// Wait briefly before checking status again
		time.Sleep(2 * time.Second)
	}

	if finalImageURL == "" {
		return nil, "", fmt.Errorf("no completed images found")
	}

	// Change to the original quality image URL
	finalImageURL = strings.Replace(finalImageURL, "_512.webp", "_raw.jpg", 1)

	println("finalImageURL:", finalImageURL)
	return fetchImageFromURL(finalImageURL)
}

func (a *OpenArtPainter) PaintFromReference(
	refImageBytes []byte, refContentType string, prompt string, width int, height int,
) ([]byte, string, error) {
	// 1. Upload image to OpenArt
	uploadImageURL, err := a.uploadImage(refImageBytes, refContentType)
	if err != nil {
		return nil, "", fmt.Errorf("failed to upload image: %v", err)
	}

	// 2. Request image generation from OpenArt API
	currentModel := "Flux_dev"
	requestData := OpenArtRequest{
		Prompt:              &prompt, // Changed to pointer to allow null
		IsGeneratedPrompt:   true,
		ImageNum:            1, // By default, generate only one image
		Width:               width,
		Height:              height,
		Steps:               28,
		CfgScale:            3.5,
		Seed:                nil,                 // Changed to nil to allow null
		PromptAssistantMode: &[]string{"off"}[0], // Changed to pointer to allow null
		BaseModel:           &currentModel,       // Changed to pointer to allow null
		// NegativePrompt:      &[]string{"worst quality, bad quality, displeasing, very displeasing, lowres, bad anatomy, bad perspective, bad proportions, bad aspect ratio, bad face, long face, bad teeth, bad neck, long neck, bad arm, bad hands, bad ass, bad leg, bad feet, bad reflection, bad shadow, bad link, bad source, wrong hand, wrong feet, missing limb, missing eye, missing tooth, missing ear, missing finger, extra faces, extra eyes, extra eyebrows, extra mouth, extra tongue, extra teeth, extra ears, extra breasts, extra arms, extra hands, extra legs, extra digits, fewer digits, cropped head, cropped torso, cropped shoulders, cropped arms, cropped legs, mutation, deformed, disfigured, unfinished, chromatic aberration, text, error, jpeg artifacts, watermark, scan, scan artifacts, shadow, shade, shaded face"}[0], // Changed to pointer to allow null
		Tiling:   false,
		Sampler:  &[]string{"DPM++ 2M SDE Karras"}[0], // Changed to pointer to allow null
		AiModel:  &currentModel,                       // Changed to pointer to allow null
		ImageURL: &uploadImageURL,
		Strength: &[]float64{0.8}[0],
	}

	requestBody, err := json.Marshal(requestData)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create request payload: %v", err)
	}

	req, err := http.NewRequest("POST", "https://openart.ai/api/create/flux", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, "", fmt.Errorf("failed to create HTTP request: %v", err)
	}

	// Set Cookie header
	req.Header.Set("Cookie", "__Secure-next-auth.session-token="+a.ApiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("failed to make API request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("unexpected response status: %v", resp.Status)
	}

	var openArtResp OpenArtFluxResponse
	if err := json.NewDecoder(resp.Body).Decode(&openArtResp); err != nil {
		return nil, "", fmt.Errorf("failed to decode response: %v", err)
	}

	if len(openArtResp.GenerationHistoryIDs) == 0 {
		return nil, "", fmt.Errorf("no generation history IDs found")
	}

	// 2. Periodically check the status of the generated image
	var finalImageURL string
	for {
		// Request to check the status of the image
		placeholderURL := fmt.Sprintf("https://openart.ai/api/create/image_placeholder?generation_history_id=%s", openArtResp.GenerationHistoryIDs[0])
		statusReq, err := http.NewRequest("GET", placeholderURL, nil)
		if err != nil {
			return nil, "", fmt.Errorf("failed to create status request: %v", err)
		}
		statusReq.Header.Set("Cookie", "__Secure-next-auth.session-token="+a.ApiKey)

		statusResp, err := client.Do(statusReq)
		if err != nil {
			return nil, "", fmt.Errorf("failed to check image status: %v", err)
		}
		defer statusResp.Body.Close()

		if statusResp.StatusCode != http.StatusOK {
			return nil, "", fmt.Errorf("unexpected status response: %v", statusResp.Status)
		}

		var imagePlaceholderResp ImagePlaceholderResponse
		if err := json.NewDecoder(statusResp.Body).Decode(&imagePlaceholderResp); err != nil {
			return nil, "", fmt.Errorf("failed to decode status response: %v", err)
		}

		// Check image status
		allCompleted := true
		for _, img := range imagePlaceholderResp.Images {
			if img.Status == "completed" {
				finalImageURL = img.URL // Save the URL of the first completed image
				break
			} else if img.Status == "pending" {
				allCompleted = false
			}
		}

		if allCompleted {
			break
		}

		// Wait briefly before checking status again
		time.Sleep(2 * time.Second)
	}

	if finalImageURL == "" {
		return nil, "", fmt.Errorf("no completed images found")
	}

	// Change to the original quality image URL
	finalImageURL = strings.Replace(finalImageURL, "_512.webp", "_raw.jpg", 1)

	println("finalImageURL:", finalImageURL)
	return fetchImageFromURL(finalImageURL)
}

// Function to download and return the image from the URL -> (image bytes, MIME type, error)
func fetchImageFromURL(url string) ([]byte, string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, "", fmt.Errorf("failed to fetch image: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("failed to fetch image: %v", resp.Status)
	}

	// Load the entire response body into memory
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read image body: %v", err)
	}

	// Get the MIME type from the header
	mimeType := resp.Header.Get("Content-Type")

	// Return the image as is
	return body, mimeType, nil
}

func (a *OpenArtPainter) uploadImage(imageBytes []byte, mimeType string) (string, error) {
	// 1. Upload image to OpenArt
	// POST https://openart.ai/api/media/upload_image
	// form-data key: file, value: image bytes
	fileFieldName := "file"
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	refFileExtension, err := utils.GetExtensionFromMimeType(mimeType)
	if err != nil {
		return "", fmt.Errorf("failed to get file extension from content type: %v", err)
	}
	uploadFileName := "image" + refFileExtension

	part, err := writer.CreateFormFile(fileFieldName, uploadFileName) // Use a suitable filename
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %v", err)
	}

	if _, err := part.Write(imageBytes); err != nil {
		return "", fmt.Errorf("failed to write image bytes to form file: %v", err)
	}

	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("failed to close writer: %v", err)
	}

	req, err := http.NewRequest("POST", "https://openart.ai/api/media/upload_image", body)
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %v", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Cookie", "__Secure-next-auth.session-token="+a.ApiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make API request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected response status: %v", resp.Status)
	}

	var uploadResponse struct {
		ImageURL string `json:"imageUrl"`
	} // Define an inline struct to match the API response
	if err := json.NewDecoder(resp.Body).Decode(&uploadResponse); err != nil {
		return "", fmt.Errorf("failed to decode upload response: %v", err)
	}

	return uploadResponse.ImageURL, nil
}

// enhance prompt from OpenArt API
// recommend to specify graphic style in first. and then specify the surrounding environment.
// must be shorter than 300 characters!
// ex) "(realistic) beautiful portrait of a woman, white background"
func (a *OpenArtPainter) EnhancePrompt(prompt string) (string, error) {
	// Prepare request data for prompt generation
	requestData := map[string]interface{}{
		// "colorScheme":          "",
		"prompt":               prompt,
		"num_return_sequences": 1,
		"model":                "gpt-4o-mini",
		// "style":                "photorealism",
	}

	// Encode request data in JSON format
	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal prompt request data: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", "https://openart.ai/api/common/prompt", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create prompt request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", a.ApiKey))

	// Send request using HTTP client
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send prompt request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("received non-200 response code: %d, body: %s", resp.StatusCode, string(body))
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read prompt response body: %w", err)
	}

	// Parse response
	var responseData []struct {
		Prompt string `json:"prompt"`
		ID     string `json:"id"`
	}
	if err := json.Unmarshal(body, &responseData); err != nil {
		return "", fmt.Errorf("failed to unmarshal prompt response: %w", err)
	}

	// Return the enhanced prompt from the first item in the response
	if len(responseData) > 0 {
		return responseData[0].Prompt, nil
	}

	return "", fmt.Errorf("no enhanced prompt found in response")
}

func (a *OpenArtPainter) ChangeStyle(imageBytes []byte, contentType string, prompt string) ([]byte, string, error) {
	// 1. Upload image to OpenArt
	uploadImageURL, err := a.uploadImage(imageBytes, contentType)
	if err != nil {
		return nil, "", fmt.Errorf("failed to upload image: %v", err)
	}

	// 2. Change style of the image
	// POST https://openart.ai/api/apps/create
	payload := map[string]interface{}{
		"app_name":            "creative-variations",
		"image_num":           1,
		"similarity":          1,
		"style":               "Default",
		"subject_description": prompt,
		"upload_image_url":    uploadImageURL,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, "", fmt.Errorf("failed to marshal request payload: %w", err)
	}

	req, err := http.NewRequest("POST", "https://openart.ai/api/apps/create", bytes.NewBuffer(payloadBytes))

	if err != nil {
		return nil, "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", a.ApiKey))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("failed to send style change request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, "", fmt.Errorf("received non-200 response code: %d, body: %s", resp.StatusCode, string(body))
	}

	var responseData struct {
		InferenceRequestID  string `json:"inference_request_id"`
		GenerationHistoryID string `json:"generation_history_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		return nil, "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// 3. Periodically check the status of the generated image
	var finalImageURL string
	refetchCount := 0
	for {
		// Request to check the status of the image
		placeholderURL := fmt.Sprintf("https://openart.ai/api/create/image_placeholder?generation_history_id=%s", responseData.GenerationHistoryID)
		statusReq, err := http.NewRequest("GET", placeholderURL, nil)
		if err != nil {
			return nil, "", fmt.Errorf("failed to create status request: %v", err)
		}
		statusReq.Header.Set("Cookie", "__Secure-next-auth.session-token="+a.ApiKey)

		statusResp, err := client.Do(statusReq)
		if err != nil {
			return nil, "", fmt.Errorf("failed to check image status: %v", err)
		}
		defer statusResp.Body.Close()

		if statusResp.StatusCode != http.StatusOK {
			return nil, "", fmt.Errorf("unexpected status response: %v", statusResp.Status)
		}

		var imagePlaceholderResp ImagePlaceholderResponse
		if err := json.NewDecoder(statusResp.Body).Decode(&imagePlaceholderResp); err != nil {
			return nil, "", fmt.Errorf("failed to decode status response: %v", err)
		}

		// Check image status
		if len(imagePlaceholderResp.Images) > 0 {
			completed := false
			for _, img := range imagePlaceholderResp.Images {
				if img.Status == "completed" && img.URL != "" {
					finalImageURL = img.URL // Save the URL of the first completed image
					completed = true
					break
				}
			}
			if completed {
				break
			}
		} else {
			refetchCount += 1
			if refetchCount > 3 {
				return nil, "", fmt.Errorf("no images found in result")
			}
		}

		// Wait briefly before checking status again
		time.Sleep(2 * time.Second)
	}

	imageBytes, mimeType, err := fetchImageFromURL(finalImageURL)
	if err != nil {
		return nil, "", fmt.Errorf("failed to fetch image from URL: %v", err)
	}

	return imageBytes, mimeType, nil
}
