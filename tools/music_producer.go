package tools

import (
	"avazon-api/controllers/errs"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type MusicProducer interface {
	Produce(title string, style string, description string) ([]byte, error) // returns .mp3
}

type JENAIProducer struct {
	ApiKey string
}

type JENAIRequest struct {
	Prompt        string `json:"prompt"`
	Duration      int    `json:"duration"`
	Format        string `json:"format"`
	FadeOutLength int    `json:"fadeOutLength"`
}

type JENAIGeneratingTask struct {
	ID         string `json:"id"`
	URL        string `json:"url"`
	FailReason string `json:"failReason"`
}

func NewJENAIProducer(apiKey string) *JENAIProducer {
	return &JENAIProducer{ApiKey: apiKey}
}

// Produce 메서드: 음악 생성
func (mp *JENAIProducer) Produce(title string, style string, description string) ([]byte, error) {
	// 1. Generate Music
	generateURL := "https://app.jenmusic.ai/api/v1/public/track/generate"

	prompt := fmt.Sprintf("title: %s, style: %s, description: %s", title, style, description)

	// Generate Music 요청 생성
	generateReqData := JENAIRequest{
		Prompt:        prompt,
		Duration:      45,
		Format:        "mp3",
		FadeOutLength: 0,
	}

	// Generate Music 요청 전송
	generateReqDataJSON, err := json.Marshal(generateReqData)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", generateURL, bytes.NewBuffer(generateReqDataJSON))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+mp.ApiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	// find json["data"][0]["id"]
	var genResp struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&genResp); err != nil {
		return nil, err
	}
	if len(genResp.Data) == 0 {
		return nil, errs.ErrNotFound
	}

	genMusicID := genResp.Data[0].ID
	genTask := &JENAIGeneratingTask{ID: genMusicID}
	taskCh := make(chan *JENAIGeneratingTask)

	go func(task *JENAIGeneratingTask) {
		for {
			time.Sleep(5 * time.Second)
			// GET https://app.jenmusic.ai/api/v1/public/generation_status/{{objectId}}
			statusURL := "https://app.jenmusic.ai/api/v1/public/generation_status/" + genMusicID
			statusReq, err := http.NewRequest("GET", statusURL, nil)
			if err != nil {
				genTask.FailReason = err.Error()
				taskCh <- genTask
				break
			}
			statusReq.Header.Add("Authorization", "Bearer "+mp.ApiKey)
			client := &http.Client{}
			resp, err := client.Do(statusReq)
			if err != nil {
				genTask.FailReason = err.Error()
				taskCh <- genTask
				break
			}
			defer resp.Body.Close()
			// find json["data"]["status"] (should be in "generating", "validating", "validated")
			var statusResp struct {
				Data struct {
					Status string `json:"status"`
					URL    string `json:"url"`
				} `json:"data"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&statusResp); err != nil {
				genTask.FailReason = err.Error()
				taskCh <- genTask
				break
			}
			if statusResp.Data.Status == "generating" || statusResp.Data.Status == "validating" {
				continue
			} else if statusResp.Data.Status == "validated" {
				genTask.URL = statusResp.Data.URL
				taskCh <- genTask
				break
			} else {
				genTask.FailReason = "unexpected status: " + statusResp.Data.Status
				taskCh <- genTask
				break
			}
		}
	}(genTask)
	t := <-taskCh
	if t.FailReason != "" {
		fmt.Printf("failed to generate music: %v", t.FailReason)
		return nil, errs.ErrInternalServerError
	}

	// GET t.URL
	resp, err = client.Get(t.URL)
	if err != nil {
		fmt.Printf("failed to get generated music: %v", err)
		return nil, errs.ErrInternalServerError
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("unexpected response status: %v", resp.Status)
		return nil, errs.ErrInternalServerError
	}

	musicBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("failed to read music data: %v", err)
		return nil, errs.ErrInternalServerError
	}
	return musicBytes, nil
}
