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

// InferenceService handles Text2SQL inference.
// It acts as a facade that delegates to the actual inference engine.
type InferenceService struct {
	config       *config.Config
	dbService    *DatabaseService
	engine       InferenceEngineInterface
	richCtxProv  RichContextProvider
	fieldSuggest FieldSuggesterInterface
	mu           sync.RWMutex
}

// InferenceServiceOptions holds optional dependencies
type InferenceServiceOptions struct {
	RichContextProvider RichContextProvider
	FieldSuggester      FieldSuggesterInterface
}

// NewInferenceService creates a new inference service
func NewInferenceService(
	cfg *config.Config,
	dbService *DatabaseService,
	engine InferenceEngineInterface,
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
func (s *InferenceService) GetAvailableModels() []ModelInfo {
	if s.engine == nil {
		return []ModelInfo{}
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
	if getter, ok := s.engine.(LLMModelGetter); ok {
		return getter.GetLLMModelInterface()
	}
	return nil
}

// Text2SQLRequest represents a text2sql request
type Text2SQLRequest struct {
	Question         string           `json:"question"`
	DatabaseID       string           `json:"database_id"`
	Database         string           `json:"database"`
	UseRichContext   bool             `json:"use_rich_context"`
	UseReact         bool             `json:"use_react"`
	MaxIterations    int              `json:"max_iterations"`
	FieldDescription string           `json:"field_description"`
	GroundingResult  *GroundingResult `json:"grounding_result,omitempty"`
}

// Text2SQLResult represents the result of text2sql
type Text2SQLResult struct {
	SQL             string                 `json:"sql"`
	ExecutionResult *interfaces.QueryResult `json:"execution_result,omitempty"`
	Metadata        Text2SQLMetadata       `json:"metadata"`
}

// Text2SQLMetadata contains metadata about the inference
type Text2SQLMetadata struct {
	SelectedTables     []string    `json:"selected_tables"`
	Iterations         int         `json:"iterations"`
	ReactTrace         []ReActStep `json:"react_trace"`
	RichContextUpdated bool        `json:"rich_context_updated"`
	ExecutionTimeMs    int64       `json:"execution_time_ms"`
	LLMCalls           int         `json:"llm_calls"`
	Model              string      `json:"model"`
}

// Execute performs synchronous text2sql inference
func (s *InferenceService) Execute(ctx context.Context, req *Text2SQLRequest) (*Text2SQLResult, error) {
	if s.engine == nil {
		return nil, fmt.Errorf("inference engine not initialized")
	}

	startTime := time.Now()

	inferReq := &InferenceRequest{
		Question:         req.Question,
		DatabaseID:       req.DatabaseID,
		Database:         req.Database,
		UseRichContext:   req.UseRichContext,
		UseReact:         req.UseReact,
		MaxIterations:    req.MaxIterations,
		FieldDescription: req.FieldDescription,
		GroundingResult:  req.GroundingResult,
	}

	if inferReq.MaxIterations == 0 {
		inferReq.MaxIterations = s.config.React.MaxIterations
	}

	result, err := s.engine.Execute(ctx, inferReq)
	if err != nil {
		return nil, fmt.Errorf("inference failed: %w", err)
	}

	response := &Text2SQLResult{
		SQL:             result.SQL,
		ExecutionResult: result.ExecutionResult,
		Metadata: Text2SQLMetadata{
			SelectedTables:     result.Metadata.SelectedTables,
			Iterations:         result.Metadata.Iterations,
			ReactTrace:         result.Metadata.ReactTrace,
			RichContextUpdated: result.Metadata.RichContextUpdated,
			ExecutionTimeMs:    time.Since(startTime).Milliseconds(),
			LLMCalls:           result.Metadata.LLMCalls,
			Model:              result.Metadata.Model,
		},
	}

	return response, nil
}

// ExecuteStream performs streaming text2sql inference
func (s *InferenceService) ExecuteStream(ctx context.Context, req *Text2SQLRequest, events chan<- StreamEvent) error {
	if s.engine == nil {
		return fmt.Errorf("inference engine not initialized")
	}

	inferReq := &InferenceRequest{
		Question:         req.Question,
		DatabaseID:       req.DatabaseID,
		Database:         req.Database,
		UseRichContext:   req.UseRichContext,
		UseReact:         req.UseReact,
		MaxIterations:    req.MaxIterations,
		FieldDescription: req.FieldDescription,
		GroundingResult:  req.GroundingResult,
	}

	if inferReq.MaxIterations == 0 {
		inferReq.MaxIterations = s.config.React.MaxIterations
	}

	return s.engine.ExecuteStream(ctx, inferReq, events)
}

// GetRichContext returns rich context for a database
func (s *InferenceService) GetRichContext(dbID, database string) (*RichContextInfo, error) {
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
func (s *InferenceService) SuggestFields(ctx context.Context, question, dbID, database, language string) (*SuggestFieldsResult, error) {
	if s.fieldSuggest == nil {
		return nil, fmt.Errorf("field suggester not available")
	}

	req := &SuggestFieldsRequest{
		Question:   question,
		DatabaseID: dbID,
		Database:   database,
		Language:   language,
	}

	return s.fieldSuggest.SuggestFields(ctx, req)
}

// ============================================
// Spider Dataset Support
// ============================================

var allowedSpiderDatabases = map[string]bool{
	"world_1": true, "car_1": true, "cre_Doc_Template_Mgt": true,
	"dog_kennels": true, "flight_2": true, "student_transcripts_tracking": true,
	"wta_1": true, "tvshow": true, "network_1": true, "concert_singer": true,
	"pets_1": true, "poker_player": true, "orchestra": true,
	"employee_hire_evaluation": true, "course_teach": true, "singer": true,
	"museum_visit": true, "battle_death": true, "voter_1": true,
	"real_estate_properties": true,
}

type SpiderDatabase struct {
	Name           string `json:"name"`
	Path           string `json:"path"`
	HasRichContext bool   `json:"has_rich_context"`
}

type SpiderQuestion struct {
	Question string `json:"question"`
	GoldSQL  string `json:"gold_sql"`
}

func (s *InferenceService) ListSpiderDatabases(dbID string) ([]SpiderDatabase, error) {
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
		entries, err := os.ReadDir(dbConfig.Path)
		if err != nil {
			return nil, err
		}
		for _, entry := range entries {
			if entry.IsDir() {
				dbName := entry.Name()
				if !allowedSpiderDatabases[dbName] {
					continue
				}
				dbPath := filepath.Join(dbConfig.Path, dbName, dbName+".sqlite")
				if _, err := os.Stat(dbPath); err == nil {
					databases = append(databases, SpiderDatabase{
						Name:           dbName,
						Path:           dbPath,
						HasRichContext: s.checkRichContextExists(dbName),
					})
				}
			}
		}
	case "mysql", "postgres":
		for dbName := range allowedSpiderDatabases {
			databases = append(databases, SpiderDatabase{
				Name:           dbName,
				HasRichContext: s.checkRichContextExists(dbName),
			})
		}
	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbConfig.Type)
	}

	return databases, nil
}

func (s *InferenceService) checkRichContextExists(database string) bool {
	if s.richCtxProv != nil {
		return s.richCtxProv.HasRichContext(database)
	}

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

func (s *InferenceService) LoadSpiderQuestions(database string) ([]SpiderQuestion, error) {
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
		MySQL      struct{ User, Password string } `json:"mysql"`
		PostgreSQL struct{ User, Password string } `json:"postgresql"`
	} `json:"default_credentials"`
}

func (s *InferenceService) GetDemoDatabases() *DemoDatabaseConfig {
	paths := []string{
		"/app/demo_databases.json",
		"system/docker/demo_databases.json",
		"docker/demo_databases.json",
		"demo_databases.json",
	}
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var cfg DemoDatabaseConfig
		if err := json.Unmarshal(data, &cfg); err != nil {
			continue
		}
		return &cfg
	}
	return nil
}

// ============================================
// Translation Support
// ============================================

type Translator interface {
	TranslateTexts(ctx context.Context, texts []string, targetLang string) (map[string]string, error)
}

var translatorInstance Translator
var translatorMu sync.RWMutex

func SetTranslator(t Translator) {
	translatorMu.Lock()
	translatorInstance = t
	translatorMu.Unlock()
}

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
// Spider Adapter Creation
// ============================================

func (s *InferenceService) CreateSpiderAdapter(dbID, database string) (interfaces.DBAdapter, error) {
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
		dbPath := filepath.Join(dbConfig.Path, database, database+".sqlite")
		adapterCfg = &interfaces.DBConfig{Type: "sqlite", FilePath: dbPath}
	case "mysql":
		adapterCfg = &interfaces.DBConfig{
			Type: "mysql", Host: dbConfig.Host, Port: dbConfig.Port,
			Database: database, User: dbConfig.User, Password: dbConfig.Password,
		}
	case "postgres":
		adapterCfg = &interfaces.DBConfig{
			Type: "postgresql", Host: dbConfig.Host, Port: dbConfig.Port,
			Database: database, User: dbConfig.User, Password: dbConfig.Password,
		}
	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbConfig.Type)
	}

	adp, err := s.dbService.CreateCustomAdapter(&AdapterConfig{
		Type: adapterCfg.Type, Host: adapterCfg.Host, Port: adapterCfg.Port,
		User: adapterCfg.User, Password: adapterCfg.Password,
		Database: adapterCfg.Database, Path: adapterCfg.FilePath,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create adapter: %w", err)
	}

	ctx := context.Background()
	if err := adp.Connect(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect to database %s: %w", database, err)
	}
	return adp, nil
}
