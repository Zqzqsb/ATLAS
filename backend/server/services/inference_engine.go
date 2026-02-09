package services

import (
	"context"
	"fmt"

	"github.com/tmc/langchaingo/llms"

	"lucid/internal/adapter"
	"lucid/internal/inference"
	"lucid/internal/lakebase"
)

// LakebaseContextLoader defines methods needed from lakebase service for loading rich context.
type LakebaseContextLoader interface {
	GetDatasourceByName(ctx context.Context, name string) (*lakebase.Datasource, error)
	GetTablesByDatasource(ctx context.Context, dsID int64) ([]*lakebase.TableInfo, error)
	GetColumnsByDatasource(ctx context.Context, dsID int64) ([]*lakebase.ColumnInfo, error)
}

// InferenceEngine wraps internal/inference.Pipeline.
type InferenceEngine struct {
	llm          llms.Model
	currentModel string
	modelConfigs map[string]ModelInfo
	lakebaseSvc  LakebaseContextLoader
	dbService    *DatabaseService
}

// NewInferenceEngine creates a new inference engine.
func NewInferenceEngine(llm llms.Model, dbService *DatabaseService) *InferenceEngine {
	return &InferenceEngine{
		llm:          llm,
		currentModel: "default",
		modelConfigs: make(map[string]ModelInfo),
		dbService:    dbService,
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
		UseRichContext:  req.UseRichContext,
		UseReact:        req.UseReact,
		ReactLinking:  req.UseReact,
		UseDryRun:     true,
		MaxIterations: req.MaxIterations,
		ClarifyMode:   "off",
		DBName:        req.DatabaseID,
		DBType:        dbType,
	}
	if config.MaxIterations == 0 {
		config.MaxIterations = 5
	}
	return config
}

func (e *InferenceEngine) loadSchemaContext(ctx context.Context, dbName string) *inference.SchemaContext {
	if e.lakebaseSvc == nil {
		return nil
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

