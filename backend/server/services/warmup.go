package services

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/tmc/langchaingo/llms"

	"lucid/config"
	"lucid/interfaces"
)

// WarmupService handles pre-warming of various services for better first-request performance
type WarmupService struct {
	config         *config.Config
	llmModel       llms.Model
	lakebaseService *LakebaseService
	adapterFactory interfaces.AdapterFactory
	
	// Cache for warmed resources
	schemaCache    map[string]interface{}
	schemaCacheMu  sync.RWMutex
	
	// Warmup status
	warmed         bool
	warmupDuration time.Duration
}

// NewWarmupService creates a new warmup service
func NewWarmupService(cfg *config.Config, llm llms.Model, lakebase *LakebaseService, factory interfaces.AdapterFactory) *WarmupService {
	return &WarmupService{
		config:          cfg,
		llmModel:        llm,
		lakebaseService: lakebase,
		adapterFactory:  factory,
		schemaCache:     make(map[string]interface{}),
	}
}

// Warmup performs all pre-warming tasks
func (s *WarmupService) Warmup(ctx context.Context) error {
	if s.warmed {
		return nil
	}

	startTime := time.Now()
	log.Println("🔥 Starting system warmup...")

	var wg sync.WaitGroup
	errChan := make(chan error, 4)

	// Warmup LLM in parallel
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := s.warmupLLM(ctx); err != nil {
			errChan <- fmt.Errorf("LLM warmup: %w", err)
		}
	}()

	// Warmup database connections in parallel
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := s.warmupDatabaseConnections(ctx); err != nil {
			errChan <- fmt.Errorf("DB warmup: %w", err)
		}
	}()

	// Warmup embedding service in parallel
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := s.warmupEmbedding(ctx); err != nil {
			errChan <- fmt.Errorf("Embedding warmup: %w", err)
		}
	}()

	// Warmup schema cache in parallel
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := s.warmupSchemaCache(ctx); err != nil {
			errChan <- fmt.Errorf("Schema cache warmup: %w", err)
		}
	}()

	wg.Wait()
	close(errChan)

	// Collect errors
	var errors []error
	for err := range errChan {
		errors = append(errors, err)
		log.Printf("⚠️  Warmup error: %v", err)
	}

	s.warmupDuration = time.Since(startTime)
	s.warmed = true

	if len(errors) > 0 {
		log.Printf("⚠️  Warmup completed with %d errors in %v", len(errors), s.warmupDuration)
	} else {
		log.Printf("✅ System warmup completed in %v", s.warmupDuration)
	}

	return nil
}

// warmupLLM sends a simple request to warm up LLM connection
func (s *WarmupService) warmupLLM(ctx context.Context) error {
	if s.llmModel == nil {
		return nil
	}

	log.Println("   🔄 Warming up LLM connection...")
	warmupStart := time.Now()

	// Send a minimal prompt to establish connection and warm up model
	_, err := s.llmModel.Call(ctx, "Hello")
	if err != nil {
		return fmt.Errorf("LLM warmup call failed: %w", err)
	}

	log.Printf("   ✅ LLM warmed up in %v", time.Since(warmupStart))
	return nil
}

// warmupDatabaseConnections pre-establishes database connections
func (s *WarmupService) warmupDatabaseConnections(ctx context.Context) error {
	if s.adapterFactory == nil || s.config == nil {
		return nil
	}

	log.Println("   🔄 Warming up database connections...")
	warmupStart := time.Now()

	// Warmup lakebase connection if available
	if s.lakebaseService != nil {
		// Connection is already established during initialization
		log.Println("   ✅ Lake-Base connection already established")
	}

	// Warmup configured databases
	for _, dbCfg := range s.config.Databases {
		cfg := &interfaces.DBConfig{
			Type:     dbCfg.Type,
			Host:     dbCfg.Host,
			Port:     dbCfg.Port,
			Database: dbCfg.Database,
			User:     dbCfg.User,
			Password: dbCfg.Password,
			FilePath: dbCfg.Path,
		}

		adapter, err := s.adapterFactory(cfg)
		if err != nil {
			log.Printf("   ⚠️  Failed to create adapter for %s: %v", dbCfg.ID, err)
			continue
		}

		if err := adapter.Connect(ctx); err != nil {
			log.Printf("   ⚠️  Failed to connect to %s: %v", dbCfg.ID, err)
			adapter.Close()
			continue
		}

		// Execute a simple query to fully warm up the connection
		_, _ = adapter.ExecuteQuery(ctx, "SELECT 1")
		adapter.Close()
	}

	log.Printf("   ✅ Database connections warmed up in %v", time.Since(warmupStart))
	return nil
}

// warmupEmbedding pre-warms the embedding service
func (s *WarmupService) warmupEmbedding(ctx context.Context) error {
	if s.lakebaseService == nil {
		return nil
	}

	embedder := s.lakebaseService.GetEmbeddingProvider()
	if embedder == nil {
		return nil
	}

	log.Println("   🔄 Warming up embedding service...")
	warmupStart := time.Now()

	// Generate embedding for a simple test text
	_, err := embedder.Embed(ctx, "warmup test")
	if err != nil {
		return fmt.Errorf("embedding warmup failed: %w", err)
	}

	log.Printf("   ✅ Embedding service warmed up in %v", time.Since(warmupStart))
	return nil
}

// warmupSchemaCache pre-loads schema information into cache
func (s *WarmupService) warmupSchemaCache(ctx context.Context) error {
	if s.lakebaseService == nil {
		return nil
	}

	log.Println("   🔄 Warming up schema cache...")
	warmupStart := time.Now()

	// Get all datasources and pre-load their schemas
	datasources, err := s.lakebaseService.ListDatasources(ctx)
	if err != nil {
		return fmt.Errorf("failed to list datasources: %w", err)
	}

	for _, ds := range datasources {
		// Pre-load tables
		tables, err := s.lakebaseService.GetTablesByDatasource(ctx, ds.ID)
		if err != nil {
			continue
		}

		s.schemaCacheMu.Lock()
		s.schemaCache[fmt.Sprintf("tables:%d", ds.ID)] = tables
		s.schemaCacheMu.Unlock()

		// Pre-load columns
		columns, err := s.lakebaseService.GetColumnsByDatasource(ctx, ds.ID)
		if err != nil {
			continue
		}

		s.schemaCacheMu.Lock()
		s.schemaCache[fmt.Sprintf("columns:%d", ds.ID)] = columns
		s.schemaCacheMu.Unlock()
	}

	log.Printf("   ✅ Schema cache warmed up in %v (cached %d items)", time.Since(warmupStart), len(s.schemaCache))
	return nil
}

// GetWarmupStatus returns warmup status
func (s *WarmupService) GetWarmupStatus() map[string]interface{} {
	return map[string]interface{}{
		"warmed":          s.warmed,
		"warmup_duration": s.warmupDuration.String(),
		"schema_cache_size": len(s.schemaCache),
	}
}

// GetCachedSchema retrieves cached schema if available
func (s *WarmupService) GetCachedSchema(key string) (interface{}, bool) {
	s.schemaCacheMu.RLock()
	defer s.schemaCacheMu.RUnlock()
	val, ok := s.schemaCache[key]
	return val, ok
}
