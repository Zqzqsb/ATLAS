package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"lucid/config"
	"lucid/interfaces"
)

// InferenceService handles Text2SQL inference
// It acts as a facade that delegates to the actual inference engine
type InferenceService struct {
	config       *config.Config
	dbService    *DatabaseService
	engine       interfaces.InferenceEngine
	richCtxProv  interfaces.RichContextProvider
	fieldSuggest interfaces.FieldSuggester
	mu           sync.RWMutex
}

// InferenceServiceOptions holds optional dependencies
type InferenceServiceOptions struct {
	RichContextProvider interfaces.RichContextProvider
	FieldSuggester      interfaces.FieldSuggester
}

// NewInferenceService creates a new inference service
func NewInferenceService(
	cfg *config.Config,
	dbService *DatabaseService,
	engine interfaces.InferenceEngine,
	opts *InferenceServiceOptions,
) *InferenceService {
	svc := &InferenceService{
		config:    cfg,
		dbService: dbService,
		engine:    engine,
	}

	if opts != nil {
		svc.richCtxProv = opts.RichContextProvider
		svc.fieldSuggest = opts.FieldSuggester
	}

	return svc
}

// GetAvailableModels returns list of available models
func (s *InferenceService) GetAvailableModels() []interfaces.ModelInfo {
	if s.engine == nil {
		return []interfaces.ModelInfo{}
	}
	return s.engine.GetAvailableModels()
}

// SwitchModel switches the LLM model
func (s *InferenceService) SwitchModel(modelID string) error {
	if s.engine == nil {
		return fmt.Errorf("inference engine not initialized")
	}
	return s.engine.SwitchModel(modelID)
}

// GetCurrentModel returns the current model ID
func (s *InferenceService) GetCurrentModel() string {
	if s.engine == nil {
		return ""
	}
	return s.engine.GetCurrentModel()
}

// LLMModelGetter is an interface for getting the LLM model
type LLMModelGetter interface {
	GetLLMModelInterface() interface{}
}

// GetLLMModel returns the LLM model if available
func (s *InferenceService) GetLLMModel() interface{} {
	if s.engine == nil {
		return nil
	}
	// Try to get LLM model from engine using type assertion
	if getter, ok := s.engine.(LLMModelGetter); ok {
		return getter.GetLLMModelInterface()
	}
	return nil
}

// Text2SQLRequest represents a text2sql request
type Text2SQLRequest struct {
	Question         string                      `json:"question"`
	DatabaseID       string                      `json:"database_id"`
	Database         string                      `json:"database"` // Specific database name (for Spider)
	UseRichContext   bool                        `json:"use_rich_context"`
	UseReact         bool                        `json:"use_react"`
	MaxIterations    int                         `json:"max_iterations"`
	FieldDescription string                      `json:"field_description"`                // Optional field clarification description
	GroundingResult  *interfaces.GroundingResult `json:"grounding_result,omitempty"` // Pre-computed grounding result
}

// Text2SQLResult represents the result of text2sql
type Text2SQLResult struct {
	SQL             string                   `json:"sql"`
	ExecutionResult *interfaces.QueryResult  `json:"execution_result,omitempty"`
	Metadata        Text2SQLMetadata         `json:"metadata"`
}

// Text2SQLMetadata contains metadata about the inference
type Text2SQLMetadata struct {
	SelectedTables     []string    `json:"selected_tables"`
	Iterations         int         `json:"iterations"`
	ReactTrace         []ReactStep `json:"react_trace"`
	RichContextUpdated bool        `json:"rich_context_updated"`
	ExecutionTimeMs    int64       `json:"execution_time_ms"`
	LLMCalls           int         `json:"llm_calls"`
	Model              string      `json:"model"` // Model used for this inference
}

// ReactStep represents a ReAct step for visualization
type ReactStep struct {
	Step        int         `json:"step"`
	Thought     string      `json:"thought"`
	Action      string      `json:"action"`
	ActionInput interface{} `json:"action_input,omitempty"`
	Observation string      `json:"observation,omitempty"`
	Timestamp   int64       `json:"timestamp"`
	Phase       string      `json:"phase"` // "schema_linking" or "sql_generation"
}

// Execute performs synchronous text2sql inference
func (s *InferenceService) Execute(ctx context.Context, req *Text2SQLRequest) (*Text2SQLResult, error) {
	if s.engine == nil {
		return nil, fmt.Errorf("inference engine not initialized")
	}

	startTime := time.Now()

	// Build inference request
	inferReq := &interfaces.InferenceRequest{
		Question:         req.Question,
		DatabaseID:       req.DatabaseID,
		Database:         req.Database,
		UseRichContext:   req.UseRichContext,
		UseReact:         req.UseReact,
		MaxIterations:    req.MaxIterations,
		FieldDescription: req.FieldDescription,
		GroundingResult:  req.GroundingResult, // Pass pre-computed grounding result
	}

	if inferReq.MaxIterations == 0 {
		inferReq.MaxIterations = s.config.React.MaxIterations
	}

	// Execute inference
	result, err := s.engine.Execute(ctx, inferReq)
	if err != nil {
		return nil, fmt.Errorf("inference failed: %w", err)
	}

	// Build response
	response := &Text2SQLResult{
		SQL:             result.SQL,
		ExecutionResult: result.ExecutionResult,
		Metadata: Text2SQLMetadata{
			SelectedTables:     result.Metadata.SelectedTables,
			Iterations:         result.Metadata.Iterations,
			ReactTrace:         convertInterfaceReactSteps(result.Metadata.ReactTrace),
			RichContextUpdated: result.Metadata.RichContextUpdated,
			ExecutionTimeMs:    time.Since(startTime).Milliseconds(),
			LLMCalls:           result.Metadata.LLMCalls,
			Model:              result.Metadata.Model,
		},
	}

	return response, nil
}

// ExecuteStream performs streaming text2sql inference
func (s *InferenceService) ExecuteStream(ctx context.Context, req *Text2SQLRequest, events chan<- interfaces.StreamEvent) error {
	if s.engine == nil {
		return fmt.Errorf("inference engine not initialized")
	}

	// Build inference request
	inferReq := &interfaces.InferenceRequest{
		Question:         req.Question,
		DatabaseID:       req.DatabaseID,
		Database:         req.Database,
		UseRichContext:   req.UseRichContext,
		UseReact:         req.UseReact,
		MaxIterations:    req.MaxIterations,
		FieldDescription: req.FieldDescription,
		GroundingResult:  req.GroundingResult, // Pass pre-computed grounding result
	}

	if inferReq.MaxIterations == 0 {
		inferReq.MaxIterations = s.config.React.MaxIterations
	}

	// Execute streaming inference - delegate to engine
	return s.engine.ExecuteStream(ctx, inferReq, events)
}

// convertInterfaceReactSteps converts interface ReAct steps to service format
func convertInterfaceReactSteps(steps []interfaces.ReActStep) []ReactStep {
	result := make([]ReactStep, len(steps))
	for i, step := range steps {
		result[i] = ReactStep{
			Step:        step.Step,
			Thought:     step.Thought,
			Action:      step.Action,
			ActionInput: step.ActionInput,
			Observation: step.Observation,
			Timestamp:   step.Timestamp,
			Phase:       step.Phase,
		}
	}
	return result
}

// GetRichContext returns rich context for a database
func (s *InferenceService) GetRichContext(dbID, database string) (*interfaces.RichContextInfo, error) {
	if s.richCtxProv == nil {
		return nil, fmt.Errorf("rich context provider not available")
	}
	return s.richCtxProv.GetRichContext(dbID, database)
}

// HasRichContext checks if rich context exists for a database
func (s *InferenceService) HasRichContext(database string) bool {
	if s.richCtxProv == nil {
		return false
	}
	return s.richCtxProv.HasRichContext(database)
}

// SuggestFields suggests output fields based on the question
func (s *InferenceService) SuggestFields(ctx context.Context, question, dbID, database, language string) (*interfaces.SuggestFieldsResult, error) {
	if s.fieldSuggest == nil {
		return nil, fmt.Errorf("field suggester not available")
	}

	req := &interfaces.SuggestFieldsRequest{
		Question:   question,
		DatabaseID: dbID,
		Database:   database,
		Language:   language,
	}

	return s.fieldSuggest.SuggestFields(ctx, req)
}

// ============================================
// Spider Dataset Support (independent of internal packages)
// ============================================

// allowedSpiderDatabases is a whitelist of Spider databases that have sample questions
var allowedSpiderDatabases = map[string]bool{
	"world_1":                      true,
	"car_1":                        true,
	"cre_Doc_Template_Mgt":         true,
	"dog_kennels":                  true,
	"flight_2":                     true,
	"student_transcripts_tracking": true,
	"wta_1":                        true,
	"tvshow":                       true,
	"network_1":                    true,
	"concert_singer":               true,
	"pets_1":                       true,
	"poker_player":                 true,
	"orchestra":                    true,
	"employee_hire_evaluation":     true,
	"course_teach":                 true,
	"singer":                       true,
	"museum_visit":                 true,
	"battle_death":                 true,
	"voter_1":                      true,
	"real_estate_properties":       true,
}

// SpiderDatabase represents a Spider database
type SpiderDatabase struct {
	Name           string `json:"name"`
	Path           string `json:"path"`
	HasRichContext bool   `json:"has_rich_context"`
}

// SpiderQuestion represents a Spider question
type SpiderQuestion struct {
	Question string `json:"question"`
	GoldSQL  string `json:"gold_sql"`
}

// ListSpiderDatabases lists available Spider databases
func (s *InferenceService) ListSpiderDatabases(dbID string) ([]SpiderDatabase, error) {
	// Find database config
	var dbConfig *config.DatabaseConfig
	for _, db := range s.config.Databases {
		if db.ID == dbID {
			dbConfig = &db
			break
		}
	}

	if dbConfig == nil {
		return nil, fmt.Errorf("database config not found: %s", dbID)
	}

	var databases []SpiderDatabase

	switch dbConfig.Type {
	case "sqlite":
		// Read directory for SQLite
		entries, err := os.ReadDir(dbConfig.Path)
		if err != nil {
			return nil, err
		}

		for _, entry := range entries {
			if entry.IsDir() {
				dbName := entry.Name()

				// Only include databases that are in the whitelist
				if !allowedSpiderDatabases[dbName] {
					continue
				}

				dbPath := filepath.Join(dbConfig.Path, dbName, dbName+".sqlite")

				if _, err := os.Stat(dbPath); err == nil {
					hasRichContext := s.checkRichContextExists(dbName)
					databases = append(databases, SpiderDatabase{
						Name:           dbName,
						Path:           dbPath,
						HasRichContext: hasRichContext,
					})
				}
			}
		}

	case "mysql", "postgres":
		// For MySQL/PostgreSQL, return the whitelist databases directly
		for dbName := range allowedSpiderDatabases {
			hasRichContext := s.checkRichContextExists(dbName)
			databases = append(databases, SpiderDatabase{
				Name:           dbName,
				Path:           "",
				HasRichContext: hasRichContext,
			})
		}

	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbConfig.Type)
	}

	return databases, nil
}

// checkRichContextExists checks if rich context exists for a database
func (s *InferenceService) checkRichContextExists(database string) bool {
	if s.richCtxProv != nil {
		return s.richCtxProv.HasRichContext(database)
	}

	// Fallback: check file paths
	possiblePaths := []string{
		filepath.Join("data", "spider", "rich_context", database+".json"),
		filepath.Join("contexts", "sqlite", "spider_old", database+".json"),
		filepath.Join("contexts", "sqlite", database+".json"),
		filepath.Join("contexts", "mysql", database+".json"),
		filepath.Join("contexts", "postgres", database+".json"),
	}

	for _, contextPath := range possiblePaths {
		if _, err := os.Stat(contextPath); err == nil {
			return true
		}
	}
	return false
}

// LoadSpiderQuestions loads Spider questions for a database
func (s *InferenceService) LoadSpiderQuestions(database string) ([]SpiderQuestion, error) {
	// Try to load from dev.json
	devPath := filepath.Join("data", "spider_data", "dev.json")

	data, err := os.ReadFile(devPath)
	if err != nil {
		return nil, err
	}

	var allQuestions []struct {
		DbId     string `json:"db_id"`
		Question string `json:"question"`
		Query    string `json:"query"`
	}

	if err := json.Unmarshal(data, &allQuestions); err != nil {
		return nil, err
	}

	// Filter by database
	var questions []SpiderQuestion
	for _, q := range allQuestions {
		if q.DbId == database {
			questions = append(questions, SpiderQuestion{
				Question: q.Question,
				GoldSQL:  q.Query,
			})
		}
	}

	return questions, nil
}

// ============================================
// Demo Database Support
// ============================================

// DemoDatabaseConfig represents the demo_databases.json structure
type DemoDatabaseConfig struct {
	Description string `json:"description"`
	GeneratedAt string `json:"generated_at"`
	Databases   []struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Domain      string `json:"domain"`
		Complexity  string `json:"complexity"`
		DataScale   string `json:"data_scale"`
		Connections map[string]struct {
			Type     string `json:"type"`
			Host     string `json:"host"`
			Port     int    `json:"port"`
			User     string `json:"user"`
			Password string `json:"password"`
			Database string `json:"database"`
			Path     string `json:"path"`
		} `json:"connections"`
		SampleQuestions []string `json:"sample_questions"`
	} `json:"databases"`
	DefaultCredentials struct {
		MySQL struct {
			User     string `json:"user"`
			Password string `json:"password"`
		} `json:"mysql"`
		PostgreSQL struct {
			User     string `json:"user"`
			Password string `json:"password"`
		} `json:"postgresql"`
	} `json:"default_credentials"`
}

// GetDemoDatabases reads and returns the demo_databases.json configuration
func (s *InferenceService) GetDemoDatabases() *DemoDatabaseConfig {
	// Try multiple paths (container paths first, then local dev paths)
	paths := []string{
		"/app/demo_databases.json",           // Container path
		"system/docker/demo_databases.json",  // Local dev path
		"docker/demo_databases.json",
		"demo_databases.json",
	}

	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		var config DemoDatabaseConfig
		if err := json.Unmarshal(data, &config); err != nil {
			continue
		}

		return &config
	}

	return nil
}

// ============================================
// Translation Support
// ============================================

// Translator interface for text translation
type Translator interface {
	TranslateTexts(ctx context.Context, texts []string, targetLang string) (map[string]string, error)
}

// translator holds the translator instance (can be set by engine)
var translatorInstance Translator
var translatorMu sync.RWMutex

// SetTranslator sets the translator instance
func SetTranslator(t Translator) {
	translatorMu.Lock()
	translatorInstance = t
	translatorMu.Unlock()
}

// TranslateTexts translates texts using the configured translator
func (s *InferenceService) TranslateTexts(ctx context.Context, texts []string, targetLang string) (map[string]string, error) {
	translatorMu.RLock()
	t := translatorInstance
	translatorMu.RUnlock()

	if t == nil {
		return nil, fmt.Errorf("translator not available")
	}

	return t.TranslateTexts(ctx, texts, targetLang)
}

// ============================================
// Spider Adapter Creation (independent)
// ============================================

// CreateSpiderAdapter creates an adapter for a specific Spider database
func (s *InferenceService) CreateSpiderAdapter(dbID, database string) (interfaces.DBAdapter, error) {
	// Find database config
	var dbConfig *config.DatabaseConfig
	for _, db := range s.config.Databases {
		if db.ID == dbID {
			dbConfig = &db
			break
		}
	}

	if dbConfig == nil {
		return nil, fmt.Errorf("database config not found: %s", dbID)
	}

	var adapterCfg *interfaces.DBConfig

	switch dbConfig.Type {
	case "sqlite":
		// Spider SQLite: path/database/database.sqlite
		dbPath := filepath.Join(dbConfig.Path, database, database+".sqlite")
		adapterCfg = &interfaces.DBConfig{
			Type:     "sqlite",
			FilePath: dbPath,
		}
	case "mysql":
		adapterCfg = &interfaces.DBConfig{
			Type:     "mysql",
			Host:     dbConfig.Host,
			Port:     dbConfig.Port,
			Database: database,
			User:     dbConfig.User,
			Password: dbConfig.Password,
		}
	case "postgres":
		adapterCfg = &interfaces.DBConfig{
			Type:     "postgresql",
			Host:     dbConfig.Host,
			Port:     dbConfig.Port,
			Database: database,
			User:     dbConfig.User,
			Password: dbConfig.Password,
		}
	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbConfig.Type)
	}

	// Use database service's adapter factory
	adp, err := s.dbService.CreateCustomAdapter(&AdapterConfig{
		Type:     adapterCfg.Type,
		Host:     adapterCfg.Host,
		Port:     adapterCfg.Port,
		User:     adapterCfg.User,
		Password: adapterCfg.Password,
		Database: adapterCfg.Database,
		Path:     adapterCfg.FilePath,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create adapter: %w", err)
	}

	// Connect to database
	ctx := context.Background()
	if err := adp.Connect(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect to database %s: %w", database, err)
	}

	return adp, nil
}

// ModelInfo re-export for handlers
type ModelInfo = interfaces.ModelInfo
