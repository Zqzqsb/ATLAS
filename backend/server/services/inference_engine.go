package services

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/tmc/langchaingo/llms"

	"lucid/internal/adapter"
	"lucid/internal/inference"
	"lucid/internal/lakebase"
	"lucid/internal/llm"
)

// LakebaseContextLoader defines methods needed from lakebase service for loading rich context.
type LakebaseContextLoader interface {
	GetDatasourceByName(ctx context.Context, name string) (*lakebase.Datasource, error)
	GetTablesByDatasource(ctx context.Context, dsID int64) ([]*lakebase.TableInfo, error)
	GetColumnsByDatasource(ctx context.Context, dsID int64) ([]*lakebase.ColumnInfo, error)
}

// schemaCacheEntry holds a cached SchemaContext with expiration time.
type schemaCacheEntry struct {
	schema    *inference.SchemaContext
	expiresAt time.Time
}

// InferenceEngine wraps internal/inference.Pipeline.
type InferenceEngine struct {
	llm          llms.Model
	currentModel string
	modelConfigs map[string]ModelInfo
	lakebaseSvc  LakebaseContextLoader
	dbService    *DatabaseService

	// Schema cache: keyed by database name, avoids repeated DB queries
	schemaCache sync.Map // map[string]*schemaCacheEntry
	cacheTTL    time.Duration
}

// NewInferenceEngine creates a new inference engine.
func NewInferenceEngine(llm llms.Model, dbService *DatabaseService) *InferenceEngine {
	return &InferenceEngine{
		llm:          llm,
		currentModel: "default",
		modelConfigs: make(map[string]ModelInfo),
		dbService:    dbService,
		cacheTTL:     5 * time.Minute, // Default TTL: 5 minutes
	}
}

// LoadModelConfigs populates modelConfigs from llm.Config so the /models API
// returns the real model list from llm_config.json.
func (e *InferenceEngine) LoadModelConfigs(cfg *llm.Config, defaultKey string) {
	for key, mc := range cfg.Models {
		// Derive a human-friendly display name from the config key
		name := formatModelName(key)
		provider := guessProvider(mc.BaseURL)
		e.modelConfigs[key] = ModelInfo{
			ID:        key,
			Name:      name,
			Provider:  provider,
			IsDefault: key == defaultKey,
		}
	}
	e.currentModel = defaultKey
}

// formatModelName turns a config key like "deepseek_v3_2" into "DeepSeek V3 2".
func formatModelName(key string) string {
	words := strings.Split(key, "_")
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	return strings.Join(words, " ")
}

// guessProvider infers the provider name from the base URL.
func guessProvider(baseURL string) string {
	switch {
	case strings.Contains(baseURL, "volces.com"):
		return "volcengine"
	case strings.Contains(baseURL, "dashscope.aliyuncs.com"):
		return "aliyun"
	case strings.Contains(baseURL, "openai.com"):
		return "openai"
	default:
		return "openai-compatible"
	}
}

// SetLakebaseService sets the lakebase service for loading rich context.
func (e *InferenceEngine) SetLakebaseService(svc LakebaseContextLoader) {
	e.lakebaseSvc = svc
}

// SetLLM updates the LLM model.
func (e *InferenceEngine) SetLLM(llm llms.Model) {
	e.llm = llm
}

// InvalidateSchemaCache invalidates the schema cache for a specific database,
// or all databases if dbName is empty.
func (e *InferenceEngine) InvalidateSchemaCache(dbName string) {
	if dbName == "" {
		e.schemaCache = sync.Map{}
		return
	}
	e.schemaCache.Delete(dbName)
}

// SetCacheTTL updates the cache TTL duration.
func (e *InferenceEngine) SetCacheTTL(ttl time.Duration) {
	e.cacheTTL = ttl
}

// WarmupSchema pre-loads schema context into cache for a given database.
// This is called by the warmup endpoint to avoid cold-start latency on the first query.
func (e *InferenceEngine) WarmupSchema(ctx context.Context, dbName string) {
	_ = e.loadSchemaContext(ctx, dbName)
}

// Execute runs inference using internal Pipeline.
func (e *InferenceEngine) Execute(ctx context.Context, req *InferenceRequest) (*InferenceResult, error) {
	if e.llm == nil {
		return nil, fmt.Errorf("LLM not initialized")
	}

	dbAdapter, err := e.createAdapter(req.DatabaseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get adapter: %w", err)
	}
	defer dbAdapter.Close()

	if err := dbAdapter.Connect(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	config := e.buildConfig(req, dbAdapter.GetDatabaseType())
	pipeline := inference.NewPipeline(e.llm, dbAdapter, config)
	defer pipeline.Reset()

	if req.UseRichContext {
		if sc := e.loadSchemaContext(ctx, req.DatabaseID); sc != nil {
			pipeline.SetSchemaContext(sc)
		}
	}

	// If external linking result is available, inject it to skip internal Schema Linking
	if len(req.LinkedTables) > 0 {
		pipeline.SetPreLinkedContext(&inference.PreLinkedContext{
			SelectedTables: req.LinkedTables,
			ContextPrompt:  req.LinkedContextPrompt,
		})
	}

	result, err := pipeline.Execute(ctx, req.Question)
	if err != nil {
		return nil, err
	}
	return e.convertResult(result), nil
}

// ExecuteStream runs streaming inference.
func (e *InferenceEngine) ExecuteStream(ctx context.Context, req *InferenceRequest, events chan<- StreamEvent) error {
	if e.llm == nil {
		return fmt.Errorf("LLM not initialized")
	}

	dbAdapter, err := e.createAdapter(req.DatabaseID)
	if err != nil {
		return fmt.Errorf("failed to get adapter: %w", err)
	}
	defer dbAdapter.Close()

	if err := dbAdapter.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	config := e.buildConfig(req, dbAdapter.GetDatabaseType())
	pipeline := inference.NewPipeline(e.llm, dbAdapter, config)
	defer pipeline.Reset()

	if req.UseRichContext {
		if sc := e.loadSchemaContext(ctx, req.DatabaseID); sc != nil {
			pipeline.SetSchemaContext(sc)
		}
	}

	// If external linking result is available, inject it to skip internal Schema Linking
	if len(req.LinkedTables) > 0 {
		pipeline.SetPreLinkedContext(&inference.PreLinkedContext{
			SelectedTables: req.LinkedTables,
			ContextPrompt:  req.LinkedContextPrompt,
		})
	}

	pipeline.SetStepCallback(func(step inference.ReActStep, eventType string) {
		events <- StreamEvent{
			Type: EventType(eventType),
			Data: ReActStep{
				Step:        step.Step,
				Thought:     step.Thought,
				Action:      step.Action,
				ActionInput: step.ActionInput,
				Observation: step.Observation,
				Phase:       step.Phase,
			},
		}
	})

	result, err := pipeline.Execute(ctx, req.Question)
	if err != nil {
		events <- StreamEvent{
			Type: EventError,
			Data: ErrorEventData{Error: err.Error()},
		}
		return err
	}

	events <- StreamEvent{
		Type: EventComplete,
		Data: e.convertResult(result),
	}
	return nil
}

func (e *InferenceEngine) GetAvailableModels() []ModelInfo {
	models := make([]ModelInfo, 0, len(e.modelConfigs))
	for _, m := range e.modelConfigs {
		models = append(models, m)
	}
	if len(models) == 0 {
		return []ModelInfo{
			{ID: "default", Name: "Default Model", Provider: "unknown", IsDefault: true},
		}
	}
	return models
}

func (e *InferenceEngine) SwitchModel(modelID string) error {
	e.currentModel = modelID
	return nil
}

func (e *InferenceEngine) GetCurrentModel() string {
	return e.currentModel
}

func (e *InferenceEngine) GetLLMModel() interface{} {
	return e.llm
}

// GetLLMModelInterface implements LLMModelGetter.
func (e *InferenceEngine) GetLLMModelInterface() interface{} {
	return e.llm
}

// --- Internal helpers ---

func (e *InferenceEngine) createAdapter(databaseID string) (adapter.DBAdapter, error) {
	if e.dbService == nil {
		return nil, fmt.Errorf("database service not available")
	}
	return e.dbService.NewIsolatedAdapter(databaseID)
}

func (e *InferenceEngine) buildConfig(req *InferenceRequest, dbType string) *inference.Config {
	config := &inference.Config{
		UseRichContext: req.UseRichContext,
		UseReact:       req.UseReact,
		ReactLinking:   req.UseReact,
		UseDryRun:      true,
		MaxIterations:  req.MaxIterations,
		ClarifyMode:    "off",
		DBName:         req.DatabaseID,
		DBType:         dbType,
	}
	if config.MaxIterations == 0 {
		config.MaxIterations = 5
	}

	// Wire up Field Alignment: parse field description into ResultFields for force mode
	if req.FieldDescription != "" {
		config.ClarifyMode = "force"
		config.ResultFieldsDescription = req.FieldDescription
		// Parse "FieldName (description), FieldName2 (description2)" into field names
		var fields []string
		for _, part := range strings.Split(req.FieldDescription, ",") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			// Extract field name before the parenthesis
			if idx := strings.Index(part, "("); idx > 0 {
				fields = append(fields, strings.TrimSpace(part[:idx]))
			} else {
				fields = append(fields, part)
			}
		}
		config.ResultFields = fields
	}

	return config
}

func (e *InferenceEngine) loadSchemaContext(ctx context.Context, dbName string) *inference.SchemaContext {
	if e.lakebaseSvc == nil {
		return nil
	}

	// Check cache first
	if cached, ok := e.schemaCache.Load(dbName); ok {
		entry := cached.(*schemaCacheEntry)
		if time.Now().Before(entry.expiresAt) {
			return entry.schema
		}
		// Cache expired, remove it
		e.schemaCache.Delete(dbName)
	}

	ds, err := e.lakebaseSvc.GetDatasourceByName(ctx, dbName)
	if err != nil {
		return nil
	}
	tables, err := e.lakebaseSvc.GetTablesByDatasource(ctx, ds.ID)
	if err != nil {
		return nil
	}
	columns, err := e.lakebaseSvc.GetColumnsByDatasource(ctx, ds.ID)
	if err != nil {
		return nil
	}

	sc := &inference.SchemaContext{
		DatabaseName: dbName,
		DatabaseType: ds.DBType,
		Tables:       make(map[string]*inference.SchemaTable, len(tables)),
	}

	columnsByTable := make(map[string][]*lakebase.ColumnInfo)
	for _, col := range columns {
		columnsByTable[col.TableName] = append(columnsByTable[col.TableName], col)
	}

	for _, t := range tables {
		st := &inference.SchemaTable{
			Name:     t.TableName,
			RowCount: t.RowCount,
		}
		if t.Description.Valid {
			st.Description = t.Description.String
		}
		if cols, ok := columnsByTable[t.TableName]; ok {
			for _, col := range cols {
				colSchema := inference.SchemaColumn{
					Name:         col.ColumnName,
					IsPrimaryKey: col.IsPrimaryKey,
					IsNullable:   col.IsNullable,
				}
				if col.DataType.Valid {
					colSchema.Type = col.DataType.String
				}
				if col.Description.Valid {
					colSchema.Description = col.Description.String
				}
				if col.SampleValues.Valid {
					colSchema.SampleValues = col.SampleValues.String
				}
				if col.Synonyms.Valid {
					colSchema.Synonyms = col.Synonyms.String
				}
				st.Columns = append(st.Columns, colSchema)
			}
		}
		sc.Tables[t.TableName] = st
	}

	// Store in cache
	e.schemaCache.Store(dbName, &schemaCacheEntry{
		schema:    sc,
		expiresAt: time.Now().Add(e.cacheTTL),
	})

	return sc
}

func (e *InferenceEngine) convertResult(r *inference.Result) *InferenceResult {
	steps := make([]ReActStep, len(r.ReActSteps))
	for i, s := range r.ReActSteps {
		steps[i] = ReActStep{
			Step:        s.Step,
			Thought:     s.Thought,
			Action:      s.Action,
			ActionInput: s.ActionInput,
			Observation: s.Observation,
			Phase:       s.Phase,
		}
	}

	result := &InferenceResult{
		SQL: r.GeneratedSQL,
		Metadata: InferenceMetadata{
			SelectedTables: r.SelectedTables,
			Iterations:     r.LLMCalls,
			ExecutionTime:  r.TotalTime,
			ReactTrace:     steps,
			Model:          e.currentModel,
		},
	}

	if r.ExecutionResult != nil {
		if qr, ok := r.ExecutionResult.(*adapter.QueryResult); ok {
			result.ExecutionResult = qr
		}
	}
	return result
}
