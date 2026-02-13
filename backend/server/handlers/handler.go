package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tmc/langchaingo/llms"

	"lucid/internal/agent"
	"lucid/internal/config"
	"lucid/internal/grounding"
	"lucid/internal/logger"
	"lucid/server/services"
)

// Handler holds dependencies for HTTP handlers.
type Handler struct {
	config           *config.Config
	dbService        *services.DatabaseService
	inferenceService *services.InferenceService
	lakebaseService  *services.LakebaseService
	groundingService *grounding.Service
	groundingReady   bool // true when grounding routes are available
	agentService     *agent.AgentService
	evolutionService *agent.EvolutionService
}

// HandlerDependencies holds all dependencies needed to create handlers.
type HandlerDependencies struct {
	Config           *config.Config
	DBService        *services.DatabaseService
	InferenceService *services.InferenceService
	LakebaseService  *services.LakebaseService
	GroundingService *grounding.Service
}

// New creates a new Handler instance from dependencies.
func New(deps *HandlerDependencies) (*Handler, error) {
	if deps.Config == nil {
		return nil, fmt.Errorf("config is required")
	}
	if deps.DBService == nil {
		return nil, fmt.Errorf("database service is required")
	}
	if deps.InferenceService == nil {
		return nil, fmt.Errorf("inference service is required")
	}

	return &Handler{
		config:           deps.Config,
		dbService:        deps.DBService,
		inferenceService: deps.InferenceService,
		lakebaseService:  deps.LakebaseService,
		groundingService: deps.GroundingService,
		groundingReady:   deps.GroundingService != nil,
	}, nil
}

// InitEvolution sets up the agent and evolution services on the handler.
// Call after the handler and lakebase service are ready.
func (h *Handler) InitEvolution() {
	if h.lakebaseService == nil || !h.lakebaseService.IsConnected() {
		return
	}
	pool := h.lakebaseService.GetPool()
	repo := h.lakebaseService.GetRepository()
	if pool == nil || repo == nil {
		return
	}
	h.agentService = agent.NewAgentService(pool, nil)
	if h.agentService == nil {
		return
	}

	// Wire LLM model into the agent service so ContextMaintainer can work
	if llmRaw := h.inferenceService.GetLLMModel(); llmRaw != nil {
		if llmModel, ok := llmRaw.(llms.Model); ok {
			h.agentService.SetLLMModel(llmModel)
			logger.L().Info("Agent service LLM model set")
		}
	}

	h.evolutionService = agent.NewEvolutionService(pool, repo, h.agentService)
}

// Close cleans up resources.
func (h *Handler) Close() {
	if h.agentService != nil {
		h.agentService.Stop()
	}
	if h.dbService != nil {
		h.dbService.Close()
	}
	if h.lakebaseService != nil {
		h.lakebaseService.Close()
	}
}

// GetSystemInfo returns system information.
func (h *Handler) GetSystemInfo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"llm": gin.H{
			"default_model":    h.config.LLM.DefaultModel,
			"available_models": h.inferenceService.GetAvailableModels(),
		},
		"react": gin.H{
			"max_iterations": h.config.React.MaxIterations,
		},
	})
}

// GetModels returns list of available models.
func (h *Handler) GetModels(c *gin.Context) {
	models := h.inferenceService.GetAvailableModels()
	c.JSON(http.StatusOK, gin.H{"models": models})
}

// SwitchModel switches the current LLM model.
func (h *Handler) SwitchModel(c *gin.Context) {
	var req struct {
		ModelID string `json:"model_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.inferenceService.SwitchModel(req.ModelID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Model switched successfully",
		"model":   req.ModelID,
	})
}
