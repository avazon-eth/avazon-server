package tools

import (
	"avazon-api/utils"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strings"
	"sync"
	"time"
)

type VideoProducer interface {
	// Create creates a video from the image and returns the video as bytes
	Create(imageURL string, prompt string) ([]byte, error)
	// CreateAsync -> (doneCh, failCh)
	CreateAsync(imageURL string, prompt string) <-chan *VideoGenTask
	// SaveImageAsset saves the image as an asset and returns the URL
	SaveImageAsset(image []byte, imageType string) (string, error)
	GetTasks() []*VideoGenTask
}

type RunwayVideoProducer struct {
	ApiKey          string
	accessToken     string             // Short token. Renewed with each request
	ReadyTaskCh     chan *VideoGenTask // Channel for preparing tasks -> starts when canStart is true
	ProcessingTasks []*VideoGenTask    // Tasks in progress -> moved to DoneQueue upon completion
	DoneCh          chan *VideoGenTask // Channel to notify task completion
	FailCh          chan *VideoGenTask // Channel to notify task failure
	mu              sync.Mutex
}

// Data structure managed by taskQueue
type VideoGenTask struct {
	ID         string    `json:"id"`          // != "" after started
	ImageURL   string    `json:"image_url"`   // required
	Prompt     string    `json:"prompt"`      // required
	VideoURL   string    `json:"video_url"`   // != "" after done
	FailReason string    `json:"fail_reason"` // != "" if failed
	CreatedAt  time.Time `json:"created_at"`  // != zero time after created
}

func NewRunwayVideoProducer(apiKey string) *RunwayVideoProducer {
	ret := &RunwayVideoProducer{
		ApiKey:          apiKey,
		ReadyTaskCh:     make(chan *VideoGenTask),
		ProcessingTasks: make([]*VideoGenTask, 0),
		DoneCh:          make(chan *VideoGenTask),
		FailCh:          make(chan *VideoGenTask),
	}
	go ret.taskLoop()
	return ret
}

func (va *RunwayVideoProducer) Create(imageURL string, prompt string) ([]byte, error) {
	imageBytes, mimeType, err := utils.GetDataFromURL(imageURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get image bytes: %v", err)
	}

	assetFilmExtension, err := utils.GetExtensionFromMimeType(mimeType)
	if err != nil {
		return nil, fmt.Errorf("failed to get file extension from MIME type: %v", err)
	}
	assetFileName := fmt.Sprintf("%d.%s", time.Now().Unix(), assetFilmExtension)
	assetURL, err := va.uploadImageAsset(assetFileName, imageBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to upload image asset: %v", err)
	}

	// Start new task
	task := &VideoGenTask{
		ImageURL:  assetURL,
		Prompt:    prompt,
		CreatedAt: time.Now(),
	}
	va.ReadyTaskCh <- task
	for {
		select {
		case t := <-va.FailCh:
			if task.ID == t.ID {
				return nil, fmt.Errorf("failed to create video: %s", t.FailReason)
			} else {
				va.FailCh <- t
			}
		case t := <-va.DoneCh:
			if task.ID == t.ID {
				// Fetch from t.VideoURL
				resp, err := http.Get(t.VideoURL)
				if err != nil {
					return nil, fmt.Errorf("failed to fetch video: %v", err)
				}
				defer resp.Body.Close()
				if resp.StatusCode != http.StatusOK {
					return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
				}
				return io.ReadAll(resp.Body)
			} else {
				va.DoneCh <- t
			}
		default:
			// HTTP Polling
			time.Sleep(10 * time.Second)
		}
	}
}

func (va *RunwayVideoProducer) CreateAsync(imageURL string, prompt string) <-chan *VideoGenTask {
	task := &VideoGenTask{
		ImageURL:  imageURL,
		Prompt:    prompt,
		CreatedAt: time.Now(),
	}
	va.ReadyTaskCh <- task
	taskCh := make(chan *VideoGenTask)
	go func() {
		for {
			select {
			case t := <-va.FailCh:
				if task.ID == t.ID {
					taskCh <- t
				} else {
					va.FailCh <- t
				}
			case t := <-va.DoneCh:
				if task.ID == t.ID {
					taskCh <- t
				} else {
					va.DoneCh <- t
				}
			default:
				time.Sleep(10 * time.Second)
			}
		}
	}()
	return taskCh
}

func (va *RunwayVideoProducer) SaveImageAsset(image []byte, imageType string) (string, error) {
	// First of all, issue a new token
	_, err := va.issueToken()
	if err != nil {
		return "", fmt.Errorf("failed to issue token: %v", err)
	}

	// 1. Upload image to asset
	// Get current time as Epoch Time
	timeline := time.Now().Unix()
	// Extract file extension from MIME type
	exts, err := mime.ExtensionsByType(imageType)
	if err != nil || len(exts) == 0 {
		return "", fmt.Errorf("failed to get file extension for MIME type: %v", imageType)
	}
	// Select the first extension
	ext := exts[0]
	// Add if the extension does not start with a dot
	if ext[0] != '.' {
		ext = "." + ext
	}
	// Generate file name
	filename := fmt.Sprintf("%d%s", timeline, ext)
	assetID, err := va.uploadImageAssetFile(filename, image, "DATASET")
	if err != nil {
		return "", fmt.Errorf("failed to upload image asset: %v", err)
	}
	previewID, err := va.uploadImageAssetFile(filename, image, "DATASET_PREVIEW")
	if err != nil {
		return "", fmt.Errorf("failed to upload preview image asset: %v", err)
	}
	// Find json["dataset"]["url"]
	resp, err := va.post("https://api.runwayml.com/v1/datasets", map[string]interface{}{
		"fileCount":        1,
		"name":             filename,
		"uploadId":         assetID,
		"previewUploadIds": []string{previewID},
		"favorite":         false,
		"type": map[string]interface{}{
			"name":        "image",
			"type":        "image",
			"isDirectory": false,
		},
		"asTeamId":      18904146,
		"privateInTeam": true,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create dataset: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	var respData struct {
		Dataset struct {
			URL string `json:"url"`
		} `json:"dataset"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
		return "", fmt.Errorf("failed to decode response body: %v", err)
	}

	return respData.Dataset.URL, nil
}

func (va *RunwayVideoProducer) GetTasks() []*VideoGenTask {
	return va.ProcessingTasks
}

func (va *RunwayVideoProducer) get(url string) (*http.Response, error) {
	// Create GET request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Set Authorization header
	req.Header.Set("Authorization", "Bearer "+va.ApiKey)

	// Create HTTP client and execute request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %v", err)
	}

	return resp, nil
}

func (va *RunwayVideoProducer) post(url string, body map[string]interface{}) (*http.Response, error) {
	// Create request body
	reqBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %v", err)
	}

	// Create POST request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Set Authorization header
	req.Header.Set("Authorization", "Bearer "+va.ApiKey)
	req.Header.Set("Content-Type", "application/json")

	// Create HTTP client and execute request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %v", err)
	}

	return resp, nil
}

func (va *RunwayVideoProducer) issueToken() (string, error) {
	// Token issuance URL for Runway API
	tokenURL := "https://api.runwayml.com/v1/short_jwt"
	// Issue token via POST request
	resp, err := va.post(tokenURL, map[string]interface{}{})
	if err != nil {
		return "", fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Check response status code
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Extract token from response body
	var respData struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
		return "", fmt.Errorf("failed to decode response body: %v", err)
	}

	// Store issued token in RunwayVideoProducer's accessToken
	va.accessToken = respData.Token

	return respData.Token, nil
}

// Returns (object_id, error)
// dataType: "DATASET" or "DATASET_PREVIEW"
func (va *RunwayVideoProducer) uploadImageAssetFile(filename string, image []byte, dataType string) (string, error) {
	// 1. Start image upload
	resp1, err := va.post("https://api.runwayml.com/v1/uploads", map[string]interface{}{
		"filename":      filename,
		"numberOfParts": 1,
		"type":          dataType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to make request: %v", err)
	}
	defer resp1.Body.Close()
	// Check response status code
	if resp1.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp1.StatusCode)
	}
	var resp1Data struct {
		ID            string   `json:"id"`
		UploadUrls    []string `json:"uploadUrls"`
		UploadHeaders struct {
			ContentType string `json:"Content-Type"`
		} `json:"uploadHeaders"`
	}
	if err := json.NewDecoder(resp1.Body).Decode(&resp1Data); err != nil {
		return "", fmt.Errorf("failed to decode response body: %v", err)
	}

	// 2. Upload image to S3
	putReq, err := http.NewRequest("PUT", resp1Data.UploadUrls[0], bytes.NewBuffer(image))
	if err != nil {
		return "", fmt.Errorf("failed to create PUT request: %v", err)
	}

	// Set Content-Type header
	putReq.Header.Set("Content-Type", resp1Data.UploadHeaders.ContentType)

	client := &http.Client{}
	putResp, err := client.Do(putReq)
	if err != nil {
		return "", fmt.Errorf("failed to upload image: %v", err)
	}
	defer putResp.Body.Close()

	// Return error if upload fails
	if putResp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("image upload failed with status: %d", putResp.StatusCode)
	}

	// Get ETag header value
	etag := putResp.Header.Get("ETag")
	etag = strings.Trim(etag, "\"")
	if etag == "" {
		return "", fmt.Errorf("ETag not found in upload response")
	}

	// 3. Notify upload completion
	completeURL := fmt.Sprintf("https://api.runwayml.com/v1/uploads/%s/complete", resp1Data.ID)
	resp3, err := va.post(completeURL, map[string]interface{}{
		"parts": []map[string]interface{}{
			{
				"PartNumber": 1,
				"ETag":       etag,
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to complete upload: %v", err)
	}
	defer resp3.Body.Close()

	// Check response status code
	if resp3.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to complete upload with status: %d", resp3.StatusCode)
	}

	// Return object ID if upload is successfully completed
	return resp1Data.ID, nil
}

// returns url only
func (va *RunwayVideoProducer) uploadImageAsset(filename string, imageBytes []byte) (string, error) {
	datasetID, err := va.uploadImageAssetFile(filename, imageBytes, "DATASET")
	if err != nil {
		return "", fmt.Errorf("failed to upload image asset: %v", err)
	}

	datasetPreviewID, err := va.uploadImageAssetFile(filename, imageBytes, "DATASET_PREVIEW")
	if err != nil {
		return "", fmt.Errorf("failed to upload image asset: %v", err)
	}

	// Prepare the request payload
	payload := map[string]interface{}{
		"fileCount": 1,
		"name":      filename,
		"uploadId":  datasetID,
		"previewUploadIds": []string{
			datasetPreviewID,
		},
		"favorite": false,
		"type": map[string]interface{}{
			"name":        "image",
			"type":        "image",
			"isDirectory": false,
		},
		"asTeamId":      18904146,
		"privateInTeam": true,
	}

	// Send the request to create the dataset
	datasetURL := "https://api.runwayml.com/v1/datasets"
	resp, err := va.post(datasetURL, payload)
	if err != nil {
		return "", fmt.Errorf("failed to create dataset: %v", err)
	}
	defer resp.Body.Close()

	// Check response status code
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to create dataset with status: %d", resp.StatusCode)
	}

	// Decode the response to get the dataset URL
	var datasetResponse struct {
		Dataset struct {
			URL string `json:"url"`
		} `json:"dataset"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&datasetResponse); err != nil {
		return "", fmt.Errorf("failed to decode dataset response: %v", err)
	}

	// Return the dataset URL
	return datasetResponse.Dataset.URL, nil
}

func (va *RunwayVideoProducer) checkCanStartNewTask() (bool, error) {
	_, err := va.issueToken()
	if err != nil {
		return false, err
	}
	checkURL := "https://api.runwayml.com/v1/tasks/can_start?mode=explore"
	// Check if json["canStartNewTask"]["canStartNewTask"] is true
	resp, err := va.get(checkURL)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	var respData struct {
		CanStartNewTask struct {
			CanStartNewTask bool `json:"canStartNewTask"`
		} `json:"canStartNewTask"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
		return false, fmt.Errorf("failed to decode response body: %v", err)
	}
	return respData.CanStartNewTask.CanStartNewTask, nil
}

// taskLoop is a goroutine that manages tasks which are all about generating videos
func (va *RunwayVideoProducer) taskLoop() {
	for {
		select {
		// Check if new task can start
		case t := <-va.ReadyTaskCh:
			if canStart, err := va.checkCanStartNewTask(); err != nil {
				fmt.Println("failed to check if new task can start:", err)
				continue
			} else if !canStart {
				fmt.Println("cannot start new task")
				continue
			}
			go va.startTask(t)
		case <-time.After(10 * time.Second):
			va.mu.Lock()
			// Check if any task is done
			if len(va.ProcessingTasks) > 0 {
				va.checkProcessingTaskStatus()
			}
			va.mu.Unlock()
		}
	}
}

// In here, ProcessingCh <- task will be executed and managed in taskLoop()
func (va *RunwayVideoProducer) startTask(task *VideoGenTask) {
	resp, err := va.post("https://api.runwayml.com/v1/tasks", map[string]interface{}{
		"taskType": "gen3a_turbo",
		"internal": false,
		"asTeamId": 18904146,
		"options": map[string]interface{}{
			"name":           fmt.Sprintf("Gen-3 Alpha Turbo %d", time.Now().Unix()),
			"seconds":        10,
			"text_prompt":    task.Prompt,
			"flip":           true, // portrait
			"exploreMode":    true,
			"watermark":      false,
			"enhance_prompt": true,
			"keyframes": []interface{}{
				map[string]interface{}{
					"image":     task.ImageURL,
					"timestamp": 0,
				},
			},
			"resolution":         "720p",
			"image_as_end_frame": false,
			"assetGroupName":     "Generative Video",
		},
	})
	if err != nil {
		task.FailReason = fmt.Sprintf("failed to start task: %v", err)
		va.FailCh <- task
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		task.FailReason = fmt.Sprintf("unexpected status code: %d", resp.StatusCode)
		va.FailCh <- task
		return
	}
	// Use json["task"]["id"] and json["task"]["status"]
	var respData struct {
		Task struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"task"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
		task.FailReason = fmt.Sprintf("failed to decode response body: %v", err)
		va.FailCh <- task
		return
	}
	task.ID = respData.Task.ID
	if respData.Task.Status != "PENDING" {
		task.FailReason = fmt.Sprintf("unexpected status: %s", respData.Task.Status)
		va.FailCh <- task
		return
	}
	va.mu.Lock()
	va.ProcessingTasks = append(va.ProcessingTasks, task)
	va.mu.Unlock()
}

// Must be called within mutex lock
func (va *RunwayVideoProducer) checkProcessingTaskStatus() {
	_, err := va.issueToken()
	if err != nil {
		va.onSystemError(fmt.Sprintf("failed to issue token: %v", err))
		return
	}

	// GET https://api.runwayml.com/v1/assets_pending?asTeamId=18904146&privateInTeam=true
	resp, err := va.get("https://api.runwayml.com/v1/assets_pending?asTeamId=18904146&privateInTeam=true")
	if err != nil {
		// Formatting error message
		va.onSystemError(fmt.Sprintf("failed to get pending assets: %v", err))
		return
	}
	// Hope to get json["pendingAssets"] as array
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		va.onSystemError(fmt.Sprintf("unexpected status code: %d", resp.StatusCode))
		return
	}
	var respData struct {
		PendingAssets []struct {
			ID string `json:"id"`
		} `json:"pendingAssets"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
		va.onSystemError(fmt.Sprintf("failed to decode response body: %v", err))
		return
	}

	// If task's ID is not in pendingAssets, check and move it to DoneCh
	isOnPendingList := make(map[string]bool)
	for _, tID := range respData.PendingAssets {
		isOnPendingList[tID.ID] = true
	}
	for i := len(va.ProcessingTasks) - 1; i >= 0; i-- {
		task := va.ProcessingTasks[i]
		if _, exists := isOnPendingList[task.ID]; !exists {
			va.onProcessingTaskOut(task)
			va.ProcessingTasks = append(va.ProcessingTasks[:i], va.ProcessingTasks[i+1:]...)
		}
	}
}

// Must be called within mutex lock
func (va *RunwayVideoProducer) onSystemError(failReason string) {
	fmt.Println("system error:", failReason)
	for _, task := range va.ProcessingTasks {
		task.FailReason = failReason
		va.FailCh <- task
	}
	va.ProcessingTasks = make([]*VideoGenTask, 0)
}

// When checked pending list and task is not in pending list
func (va *RunwayVideoProducer) onProcessingTaskOut(task *VideoGenTask) {
	// GET https://api.runwayml.com/v1/tasks/{task_id}
	resp, err := va.get("https://api.runwayml.com/v1/tasks/" + task.ID)
	if err != nil {
		task.FailReason = fmt.Sprintf("failed to get task status: %v", err)
		va.FailCh <- task
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		task.FailReason = fmt.Sprintf("unexpected status code: %d", resp.StatusCode)
		va.FailCh <- task
		return
	}
	// Hope to get json["task"]["status"] is "SUCCEEDED", and get json["task"]["artifacts"][0]["url"]
	var respData struct {
		Task struct {
			Status    string `json:"status"`
			Artifacts []struct {
				URL string `json:"url"`
			} `json:"artifacts"`
		} `json:"task"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
		task.FailReason = fmt.Sprintf("failed to decode response body: %v", err)
		va.FailCh <- task
		return
	}
	if respData.Task.Status != "SUCCEEDED" {
		task.FailReason = fmt.Sprintf("unexpected status: %s", respData.Task.Status)
		va.FailCh <- task
		return
	}
	task.VideoURL = respData.Task.Artifacts[0].URL
	va.DoneCh <- task
}
