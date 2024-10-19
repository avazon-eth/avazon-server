package tools

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

// Simple assistant like ChatGPT
type Assistant interface {
	// Set base system prompt
	SetSystemPrompt(string)
	// Handle user input and return response
	Handle(userInput string) (string, error)
	// Handle user input asynchronously and return response
	HandleAsync(userInput string, args ...string) (output chan string, done chan string, err chan error)
	SetTools(tools []OpenAITool)
	Init(messages []Message)
}

// OpenAIAssistant struct definition
type OpenAIAssistant struct {
	messages []Message
	tools    []OpenAITool
	apiKey   string
	model    string
}

// Define struct for API request
type OpenAIRequest struct {
	Model    string       `json:"model"`
	Messages []Message    `json:"messages"`
	Stream   bool         `json:"stream"`
	Tools    []OpenAITool `json:"tools,omitempty"`
}

type OpenAITool struct {
	Type     string             `json:"type"`
	Function OpenAIToolFunction `json:"function"`
}

type OpenAIToolFunction struct {
	Name        string                   `json:"name"`
	Description string                   `json:"description"`
	Parameters  OpenAIFunctionParameters `json:"parameters"`
}

type OpenAIFunctionParameters struct {
	Type                 string                     `json:"type"` // object
	Properties           map[string]OpenAIParameter `json:"properties"`
	Required             []string                   `json:"required"`
	AdditionalProperties bool                       `json:"additionalProperties"` // maybe false
}

type OpenAIParameter struct {
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Enum        []string `json:"enum,omitempty"`
}

type Message struct {
	Role       string        `json:"role"`
	Content    string        `json:"content"`
	ToolCalls  []ToolCallRes `json:"tool_calls,omitempty"`
	ToolCallId string        `json:"tool_call_id,omitempty"` // if Role is tool
}

// Define struct for API response
type OpenAIResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int      `json:"created"`
	Choices []Choice `json:"choices"`
}

// call from assistant
type ToolCallRes struct {
	ID       string            `json:"id"`
	Type     string            `json:"type"`
	Function OpenAIFunctionRes `json:"function"`
}

type OpenAIFunctionRes struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// Delta struct represents partial message delta in streaming response
type Delta struct {
	Type      string        `json:"type,omitempty"`
	Role      string        `json:"role,omitempty"`
	Content   string        `json:"content,omitempty"`
	ToolCalls []ToolCallRes `json:"tool_calls,omitempty"`
}

type Choice struct {
	Message      Message `json:"message"`
	Delta        Delta   `json:"delta"`
	FinishReason string  `json:"finish_reason"`
	Index        int     `json:"index"`
}

// NewOpenAIAssistant creates a new OpenAIAssistant
func NewOpenAIAssistant(apiKey string, model string) *OpenAIAssistant {
	return &OpenAIAssistant{
		messages: []Message{},
		tools:    []OpenAITool{},
		apiKey:   apiKey,
		model:    model,
	}
}

func NewOpenAIAssistantWithMessages(apiKey string, model string, messages []Message) *OpenAIAssistant {
	return &OpenAIAssistant{
		messages: messages,
		tools:    []OpenAITool{},
		apiKey:   apiKey,
		model:    model,
	}
}

// SetSystemPrompt sets the system prompt
func (a *OpenAIAssistant) SetSystemPrompt(prompt string) {
	if len(a.messages) == 0 {
		a.messages = append(a.messages, Message{Role: "system", Content: prompt})
	} else {
		if a.messages[0].Role == "system" {
			a.messages[0].Content = prompt
		} else {
			a.messages = append([]Message{{Role: "system", Content: prompt}}, a.messages...)
		}
	}
}

// Handle handles user input synchronously and returns response
func (a *OpenAIAssistant) Handle(userInput string) (string, error) {
	a.messages = append(a.messages, Message{Role: "user", Content: userInput})

	requestBody := OpenAIRequest{
		Model:    a.model,
		Messages: a.messages,
		Stream:   false,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+a.apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Read response body using io.ReadAll
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var openAIResp OpenAIResponse
	err = json.Unmarshal(body, &openAIResp)
	if err != nil {
		return "", err
	}

	respMessage := openAIResp.Choices[0].Message.Content
	a.messages = append(a.messages, Message{Role: "assistant", Content: respMessage})

	if len(openAIResp.Choices) > 0 {
		return respMessage, nil
	}

	return "", fmt.Errorf("no response from OpenAI")
}

// HandleAsync handles user input asynchronously and returns streaming response
//   - outputChan: function call format -> "function:function_name" (no arguments for now)
//   - args[0]: user role (if not given, "user" will be used)
//   - args[1]: tool call id (if not given, no tool call will be handled)
//   - args[2]: tool call name (if not given, no tool call will be handled)
//   - args[3]: tool call arguments (if not given, no tool call will be handled)
func (a *OpenAIAssistant) HandleAsync(userInput string, args ...string) (output chan string, done chan string, err chan error) {
	outputChan := make(chan string)
	doneChan := make(chan string)
	errChan := make(chan error)

	go func() {
		defer close(outputChan)
		defer close(doneChan)
		defer close(errChan)

		userRole := "user"
		if len(args) > 0 {
			userRole = args[0]
		}
		toolCallId := ""
		if len(args) > 1 {
			toolCallId = args[1]
		}
		toolCallName := ""
		if len(args) > 2 {
			toolCallName = args[2]
		}
		toolCallArguments := ""
		if len(args) > 3 {
			toolCallArguments = args[3]
		}

		if userRole == "tool" {
			a.messages = append(a.messages, Message{
				Role:      "assistant",
				Content:   userInput,
				ToolCalls: []ToolCallRes{{ID: toolCallId, Type: "function", Function: OpenAIFunctionRes{Name: toolCallName, Arguments: toolCallArguments}}},
			})
			a.messages = append(a.messages, Message{Role: userRole, Content: userInput, ToolCallId: toolCallId})
		} else if userInput != "" {
			a.messages = append(a.messages, Message{Role: userRole, Content: userInput})
		}

		requestBody := OpenAIRequest{
			Model:    a.model,
			Messages: a.messages,
			Stream:   true,
			Tools:    a.tools,
		}

		jsonData, err := json.Marshal(requestBody)
		if err != nil {
			errChan <- err
			return
		}

		req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
		if err != nil {
			errChan <- err
			return
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+a.apiKey)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			errChan <- err
			return
		}
		defer resp.Body.Close()

		respContent := ""
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()

			// Remove "data: " prefix and extract JSON only
			if len(line) > 6 && line[:6] == "data: " {
				line = line[6:]
			}

			if line == "[DONE]" {
				break
			}

			if len(line) == 0 {
				continue
			}

			var openAIResp OpenAIResponse
			err := json.Unmarshal([]byte(line), &openAIResp)
			if err != nil {
				errChan <- fmt.Errorf("error parsing JSON: %v", err)
				log.Printf("Error parsing JSON: %v, data: %s\n", err, line)
				for scanner.Scan() {
					line := scanner.Text()
					log.Println(line)
				}
				return
			}

			// Extract content from delta and send to output
			if len(openAIResp.Choices) > 0 {
				for _, choice := range openAIResp.Choices {
					// function call
					if len(choice.Delta.ToolCalls) > 0 {
						for _, toolCall := range choice.Delta.ToolCalls {
							if toolCall.Type == "function" {
								fName := toolCall.Function.Name
								cId := toolCall.ID
								outputChan <- "function:" + "{\"id\":\"" + cId + "\",\"name\":\"" + fName + "\"}" // must be handled by the caller
							} else {
								// after function call
								// argument will be given as string, in doneChan
								respContent += toolCall.Function.Arguments // must be handled by the caller
							}
						}
					}
					// chunked response message
					content := choice.Delta.Content
					if content != "" {
						outputChan <- content
						respContent += content
					}
				}
			}
		}

		if err := scanner.Err(); err != nil {
			errChan <- fmt.Errorf("error reading stream: %v", err)
			return
		}

		doneChan <- respContent
	}()

	return outputChan, doneChan, errChan
}

func (a *OpenAIAssistant) SetTools(tools []OpenAITool) {
	a.tools = tools
}

// Init sets initial messages to the assistant
func (a *OpenAIAssistant) Init(messages []Message) {
	a.messages = messages
}
