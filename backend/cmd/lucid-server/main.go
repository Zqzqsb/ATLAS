package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/tmc/langchaingo/llms"

	"lucid/internal/adapter"
	"lucid/internal/config"
	"lucid/internal/grounding"
	"lucid/internal/llm"
	"lucid/internal/logger"
	"lucid/server/handlers"
	"lucid/server/services"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "configs/system.yaml", "Path to configuration file")
	migrateOnly := flag.Bool("migrate-only", false, "Run database auto-migration and exit")
	flag.Parse()

	// Handle migrate-only mode
	if *migrateOnly {
		runMigrateOnly(*configPath)
		return
	}

	// Load configuration (logger not yet initialized — use stdlib log)
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize structured logger
	logLevel := cfg.Server.LogLevel
	if logLevel == "" {
		logLevel = "debug"
	}
	logFormat := "text"
	if cfg.Server.Mode == "release" {
		logFormat = "json"
	}
	logger.Init(logLevel, logFormat)
	slog := logger.L()
	slog.Info("Logger initialized", "level", logLevel, "format", logFormat)

	// Set Gin mode
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize dependencies
	dbService := services.NewDatabaseService(cfg, adapter.NewAdapter)

	// Initialize LLM (optional - may fail if config not found)
	var llmModel llms.Model
	llmCfg, err := llm.FindConfig()
	if err != nil {
		slog.Warn("LLM config not found, inference features limited", "error", err)
	} else {
		modelKey := cfg.LLM.DefaultModel
		if modelKey == "" {
			modelKey = llmCfg.DefaultModel
		} else {
			// Validate that the configured key actually exists in llm_config.json
			if _, lookupErr := llmCfg.GetModel(modelKey); lookupErr != nil {
				slog.Warn("Configured default_model not found in llm_config.json, falling back",
					"configured", modelKey, "fallback", llmCfg.DefaultModel)
				modelKey = llmCfg.DefaultModel
			} else {
				llmCfg.DefaultModel = modelKey
			}
		}
		llmModel, err = llmCfg.CreateLLMByKey(modelKey)
		if err != nil {
			slog.Warn("LLM initialization failed", "model", modelKey, "error", err)
		} else {
			slog.Info("LLM initialized", "model", modelKey)
		}
	}

	inferenceEngine := services.NewInferenceEngine(llmModel, dbService)

	// Load model configs so /api/v1/models returns real model list
	if llmCfg != nil {
		inferenceEngine.LoadModelConfigs(llmCfg, cfg.LLM.DefaultModel)
	}

	// Create field suggester (optional, requires LLM)
	var fieldSuggester services.FieldSuggesterInterface
	if llmModel != nil {
		fieldSuggester = services.NewFieldSuggester(llmModel, adapter.NewAdapter, cfg)
	}

	// Create inference service
	inferenceService := services.NewInferenceService(cfg, dbService, inferenceEngine, fieldSuggester)

	// ========================================
	// Initialize Lake-Base Storage Service
	// ========================================
	var lakebaseService *services.LakebaseService
	lakebaseConfigPath := "configs/lakebase.yaml"
	if _, err := os.Stat(lakebaseConfigPath); os.IsNotExist(err) {
		// Fall back to example config so demo databases are registered out of the box
		if _, err2 := os.Stat("configs/lakebase.yaml.example"); err2 == nil {
			lakebaseConfigPath = "configs/lakebase.yaml.example"
			slog.Info("Lake-Base config not found, using example", "path", lakebaseConfigPath)
		}
	}
	if _, err := os.Stat(lakebaseConfigPath); err == nil {
		lakebaseService, err = services.NewLakebaseService(lakebaseConfigPath)
		if err != nil {
			slog.Warn("Lake-Base service initialization failed", "error", err)
		} else {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			if err := lakebaseService.Connect(ctx); err != nil {
				slog.Warn("Lake-Base connection failed", "error", err)
				lakebaseService = nil
			} else {
				slog.Info("Lake-Base storage connected")
				inferenceEngine.SetLakebaseService(lakebaseService)
			}
			cancel()
		}
	} else {
		slog.Info("Lake-Base config not found, skipping")
	}

	// ========================================
	// Startup Schema Sync
	// ========================================
	if lakebaseService != nil {
		syncCtx, syncCancel := context.WithTimeout(context.Background(), 60*time.Second)
		lakebaseService.SyncAllSchemas(syncCtx, cfg.Databases, dbService.GetAdapter)
		syncCancel()
	}

	// ========================================
	// Initialize Semantic Grounding Service
	// ========================================
	var groundingService *grounding.Service
	if lakebaseService != nil {
		vectorRepo := lakebaseService.GetVectorRepository()
		embedder := lakebaseService.GetEmbeddingProvider()

		if vectorRepo != nil && embedder != nil {
			groundingService = grounding.NewService(&grounding.ServiceConfig{
				DatasourceID: 1,
				VectorRepo:   vectorRepo,
				Embedder:     embedder,
				LLMModel:     llmModel,
				Config:       grounding.DefaultGroundingConfig(),
			})
			if llmModel != nil {
				slog.Info("Semantic Grounding initialized (full mode)")
			} else {
				slog.Info("Semantic Grounding initialized (coarse-only mode)")
			}
		} else {
			slog.Warn("Semantic Grounding skipped: missing vector repo or embedder")
		}
	}

	// Create handlers with dependencies
	h, err := handlers.New(&handlers.HandlerDependencies{
		Config:           cfg,
		DBService:        dbService,
		InferenceService: inferenceService,
		LakebaseService:  lakebaseService,
		GroundingService: groundingService,
	})
	if err != nil {
		slog.Error("Failed to initialize handlers", "error", err)
		os.Exit(1)
	}

	// Initialize Agent & Evolution services
	h.InitEvolution()
	slog.Info("Agent & Evolution services initialized")

	// ========================================
	// Create Gin router
	// ========================================
	r := gin.Default()

	// Configure CORS for frontend
	r.Use(cors.New(cors.Config{
		AllowOriginFunc: func(origin string) bool {
			return true
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"timestamp": time.Now().Unix(),
		})
	})

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		slog.Info("Shutting down...")
		h.Close()
		os.Exit(0)
	}()

	// API routes
	api := r.Group("/api/v1")
	{
		// System info
		api.GET("/system/info", h.GetSystemInfo)
		api.GET("/models", h.GetModels)
		api.POST("/models/switch", h.SwitchModel)

		// Database routes
		api.GET("/databases", h.ListDatabases)
		api.GET("/databases/:id/schema", h.GetDatabaseSchema)
		api.GET("/databases/:id/tables", h.GetDatabaseTables)
		api.GET("/databases/:id/rich-context", h.GetRichContext)

		// Database connection management
		api.GET("/connections", h.ListConnections)
		api.POST("/connections", h.AddConnection)
		api.POST("/connections/test", h.TestConnection)
		api.DELETE("/connections/:id", h.RemoveConnection)
		api.POST("/connections/:id/sync-schema", h.SyncConnectionSchema)

		// Text2SQL routes
		api.POST("/text2sql", h.Text2SQL)
		api.POST("/text2sql/stream", h.Text2SQLStream)
		api.POST("/text2sql/suggest-fields", h.SuggestFields)
		api.POST("/text2sql/warmup", h.Warmup)

		// SQL execution
		api.POST("/databases/:id/execute", h.ExecuteSQL)

		// Onboarding routes
		api.GET("/onboarding/stream", h.OnboardingStream)

		// Lake-Base Storage routes
		api.GET("/lakebase/status", h.GetLakebaseStatus)
		api.POST("/lakebase/connect", h.ConnectLakebase)
		api.GET("/lakebase/datasources", h.ListLakebaseDatasources)
		api.GET("/lakebase/datasources/:id", h.GetLakebaseDatasource)
		api.GET("/lakebase/datasources/:id/context/:table", h.GetLakebaseTableContext)
		api.GET("/lakebase/datasources/:id/changelog", h.GetLakebaseChangeLogs)
		api.POST("/lakebase/datasources/:id/sync-schema", h.SyncSchema)
		api.POST("/lakebase/datasources/:id/embeddings", h.GenerateEmbeddings)
		api.POST("/lakebase/datasources/:id/generate-context", h.GenerateRichContextStream)
		api.GET("/lakebase/datasources/:id/generate-context/preview", h.PreviewForestChunks)
		api.DELETE("/lakebase/datasources/:id", h.DeleteDatasource)
		api.DELETE("/lakebase/datasources/:id/prune", h.PruneContext)
		api.POST("/lakebase/datasources/:id/context", h.AddContext)
		api.DELETE("/lakebase/datasources/:id/context", h.DeleteContext)

		// Evolution Demo routes
		api.GET("/evolution/status", h.GetEvolutionStatus)
		api.GET("/evolution/stages/:stage_id", h.GetEvolutionStagePreview)
		api.POST("/evolution/execute-stage/stream", h.ExecuteEvolutionStageStream)
		api.POST("/evolution/reset/stream", h.ResetEvolutionStream)

		// Agent routes (used by Evolution Panel)
		api.GET("/agent/logs/:id", h.GetAgentChangeLogs)

		// Grounding is integrated into text2sql pipeline — no standalone routes needed
	}

	// Serve static frontend files
	r.Static("/assets", "../web-new/dist/assets")
	r.StaticFile("/", "../web-new/dist/index.html")
	r.NoRoute(func(c *gin.Context) {
		if len(c.Request.URL.Path) >= 4 && c.Request.URL.Path[:4] == "/api" {
			c.JSON(http.StatusNotFound, gin.H{"error": "API endpoint not found"})
			return
		}
		c.File("../web-new/dist/index.html")
	})

	// Start server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	slog.Info("LUCID Server starting",
		"addr", addr,
		"api", fmt.Sprintf("http://localhost:%d/api/v1", cfg.Server.Port),
		"model", cfg.LLM.DefaultModel,
	)

	if err := r.Run(addr); err != nil {
		slog.Error("Failed to start server", "error", err)
		os.Exit(1)
	}
}

// runMigrateOnly connects to lakebase and runs auto-migration, then exits.
// Uses stdlib log since logger is not initialized in this path.
func runMigrateOnly(configPath string) {
	log.Println("Running database auto-migration...")

	lakebaseConfigPath := "configs/lakebase.yaml"
	if _, err := os.Stat(lakebaseConfigPath); err != nil {
		log.Fatalf("Lake-Base config not found at %s: %v", lakebaseConfigPath, err)
	}

	svc, err := services.NewLakebaseService(lakebaseConfigPath)
	if err != nil {
		log.Fatalf("Failed to initialize lakebase service: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := svc.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect to lakebase: %v", err)
	}
	defer svc.Close()

	log.Println("Auto-migration completed successfully")
}
