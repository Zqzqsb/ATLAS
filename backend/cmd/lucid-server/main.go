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
	"lucid/server/handlers"
	"lucid/server/services"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "configs/system.yaml", "Path to configuration file")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

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
		log.Printf("⚠️  LLM config not found: %v", err)
		log.Println("   Inference features will be limited...")
	} else {
		// Use system.yaml default_model as key, fallback to config's own default
		modelKey := cfg.LLM.DefaultModel
		if modelKey == "" {
			modelKey = llmCfg.DefaultModel
		} else {
			llmCfg.DefaultModel = modelKey
		}
		llmModel, err = llmCfg.CreateLLMByKey(modelKey)
		if err != nil {
			log.Printf("⚠️  LLM initialization failed (model=%s): %v", modelKey, err)
		} else {
			log.Printf("✅ LLM initialized: %s", modelKey)
		}
	}

	inferenceEngine := services.NewInferenceEngine(llmModel, dbService)

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
	if _, err := os.Stat(lakebaseConfigPath); err == nil {
		lakebaseService, err = services.NewLakebaseService(lakebaseConfigPath)
		if err != nil {
			log.Printf("⚠️  Lake-Base service initialization failed: %v", err)
			log.Println("   Continuing without Lake-Base features...")
		} else {
			// Connect to Lake-Base storage
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			if err := lakebaseService.Connect(ctx); err != nil {
				log.Printf("⚠️  Lake-Base connection failed: %v", err)
				lakebaseService = nil
			} else {
				log.Println("✅ Lake-Base storage connected successfully")
				// Set lakebase service to inference engine for rich context loading
				inferenceEngine.SetLakebaseService(lakebaseService)
			}
			cancel()
		}
	} else {
		log.Println("ℹ️  Lake-Base config not found, skipping...")
	}

	// ========================================
	// Startup Schema Sync (populate rc_tables/rc_columns from information_schema)
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
	var groundingHandlers *handlers.GroundingHandlers
	if lakebaseService != nil {
		// Get vector repository from lakebase service
		vectorRepo := lakebaseService.GetVectorRepository()
		embedder := lakebaseService.GetEmbeddingProvider()

		if vectorRepo != nil && embedder != nil {
			// Create grounding service (works with or without LLM for coarse-only mode)
			groundingService = grounding.NewService(&grounding.ServiceConfig{
				DatasourceID: 1, // Default datasource, can be changed per request
				VectorRepo:   vectorRepo,
				Embedder:     embedder,
				LLMModel:     llmModel, // Can be nil for coarse-only mode
				Config:       grounding.DefaultGroundingConfig(),
			})
			groundingHandlers = handlers.NewGroundingHandlers(groundingService)
			if llmModel != nil {
				log.Println("✅ Semantic Grounding service initialized (full mode)")
			} else {
				log.Println("✅ Semantic Grounding service initialized (coarse-only mode)")
			}
		} else {
			log.Println("⚠️  Semantic Grounding skipped: missing vector repo or embedder")
		}
	}

	// ========================================
	// Initialize Agent & Evolution Services
	// ========================================
	if lakebaseService != nil {
		pool := lakebaseService.GetPool()
		repo := lakebaseService.GetRepository()
		if pool != nil && repo != nil {
			handlers.InitAgentService(pool, nil)
			agentSvc := handlers.GetAgentService()
			if agentSvc != nil {
				handlers.InitEvolutionService(pool, repo, agentSvc)
				log.Println("✅ Agent & Evolution services initialized")
			}
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
		log.Fatalf("Failed to initialize handlers: %v", err)
	}


	// ========================================
	// Create Gin router
	// ========================================
	r := gin.Default()

	// Configure CORS for frontend - allow all origins for demo purposes
	r.Use(cors.New(cors.Config{
		AllowOriginFunc: func(origin string) bool {
			// Allow all origins for demo deployment
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
		log.Println("Shutting down...")
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
		api.POST("/connections/release-all", h.ReleaseAllDemoConnections)

		// Text2SQL routes
		api.POST("/text2sql", h.Text2SQL)
		api.POST("/text2sql/stream", h.Text2SQLStream)
		api.POST("/text2sql/suggest-fields", h.SuggestFields)

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
		api.POST("/lakebase/datasources/:id/generate-context", h.GenerateRichContext)
		api.POST("/lakebase/datasources/:id/generate-context/stream", h.GenerateRichContextStream)
		api.DELETE("/lakebase/datasources/:id", h.DeleteDatasource)
		api.DELETE("/lakebase/datasources/:id/prune", h.PruneContext)

		// Evolution Demo routes
		api.GET("/evolution/status", h.GetEvolutionStatus)
		api.GET("/evolution/stages/:stage_id", h.GetEvolutionStagePreview)
		api.POST("/evolution/execute-stage", h.ExecuteEvolutionStage)
		api.POST("/evolution/execute-stage/stream", h.ExecuteEvolutionStageStream)
		api.POST("/evolution/reset", h.ResetEvolution)
		api.POST("/evolution/reset/stream", h.ResetEvolutionStream)

		// Semantic Grounding routes
		if groundingHandlers != nil {
			api.POST("/grounding/ground", groundingHandlers.Ground)
			api.GET("/grounding/stream", groundingHandlers.GroundStream)
			api.GET("/grounding/config", groundingHandlers.GetConfig)
			api.PUT("/grounding/config", groundingHandlers.UpdateConfig)
			api.POST("/grounding/format", groundingHandlers.FormatPrompt)
		}
	}

	// Serve static frontend files from web-new/dist (production)
	r.Static("/assets", "../web-new/dist/assets")
	r.StaticFile("/", "../web-new/dist/index.html")
	r.NoRoute(func(c *gin.Context) {
		// Don't serve index.html for API routes - return 404 instead
		if len(c.Request.URL.Path) >= 4 && c.Request.URL.Path[:4] == "/api" {
			c.JSON(http.StatusNotFound, gin.H{"error": "API endpoint not found"})
			return
		}
		c.File("../web-new/dist/index.html")
	})

	// Start server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Printf("🚀 LUCID Server starting on %s", addr)
	log.Printf("📊 API endpoint: http://localhost:%d/api/v1", cfg.Server.Port)
	log.Printf("🔧 LLM Model: %s", cfg.LLM.DefaultModel)
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
