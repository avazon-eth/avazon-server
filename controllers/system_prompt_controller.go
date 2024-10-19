package controllers

import (
	"avazon-api/models"
	"avazon-api/services"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type SystemPromptController struct {
	service *services.SystemPromptService
}

func NewSystemPromptController(service *services.SystemPromptService) *SystemPromptController {
	return &SystemPromptController{service: service}
}

// CreateSystemPrompt handles the creation of a new system prompt
func (ctrl *SystemPromptController) CreateSystemPrompt(c *gin.Context) {
	promptID := c.Param("prompt_id")
	if len(promptID) == 0 || strings.Contains(promptID, " ") {
		HandleError(c, errors.New("prompt_id must not have whitespace"))
		return
	}

	var prompt models.SystemPrompt
	if err := c.ShouldBind(&prompt); err != nil {
		HandleError(c, err)
		return
	}
	prompt.ID = promptID

	if err := ctrl.service.UpsertSystemPrompt(prompt); err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, prompt)
}

// GetAllSystemPrompts retrieves all system prompts
func (ctrl *SystemPromptController) GetAllSystemPrompts(c *gin.Context) {
	prompts, err := ctrl.service.GetAllSystemPrompts()
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, prompts)
}

// DeleteSystemPrompt handles the deletion of a system prompt by ID
func (ctrl *SystemPromptController) DeleteSystemPrompt(c *gin.Context) {
	promptID := c.Param("prompt_id")
	if err := ctrl.service.DeleteSystemPrompt(promptID); err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Prompt deleted successfully"})
}

// Log the usage of the system prompt
func (ctrl *SystemPromptController) UpdateSystemPromptUsage(c *gin.Context) {
	agentID := c.Param("agent_id")
	promptID := c.Param("prompt_id")

	if err := ctrl.service.UpdateSystemPromptUsage(agentID, promptID); err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Usage set successfully"})
}

// GetAllSystemPromptUsages retrieves all system prompt usages
func (ctrl *SystemPromptController) GetAllSystemPromptUsages(c *gin.Context) {
	usages, err := ctrl.service.GetAllSystemPromptUsages()
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, usages)
}

// DeleteSystemPromptUsage handles the deletion of a system prompt usage by agent ID
func (ctrl *SystemPromptController) DeleteSystemPromptUsage(c *gin.Context) {
	agentID := c.Param("agent_id")
	if err := ctrl.service.DeleteSystemPromptUsage(agentID); err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Usage deleted successfully"})
}
