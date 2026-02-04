package lakebase

import (
	"encoding/json"
	"testing"
)

func TestVectorToString(t *testing.T) {
	tests := []struct {
		name    string
		input   []float32
		wantLen int // Just check approximate result, not exact string due to float precision
	}{
		{
			name:    "empty vector",
			input:   []float32{},
			wantLen: 2, // "[]"
		},
		{
			name:    "single element",
			input:   []float32{0.5},
			wantLen: 13, // "[0.50000000]" approximately
		},
		{
			name:    "multiple elements",
			input:   []float32{0.1, 0.2, 0.3},
			wantLen: 35, // approximately
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := vectorToString(tt.input)
			// Check that result starts with [ and ends with ]
			if len(result) == 0 || result[0] != '[' || result[len(result)-1] != ']' {
				t.Errorf("vectorToString(%v) = %q, should start with [ and end with ]", tt.input, result)
			}
			// Check approximate length (allow some variance for floating point)
			if len(result) < tt.wantLen-5 || len(result) > tt.wantLen+5 {
				t.Errorf("vectorToString(%v) len = %d, want approximately %d", tt.input, len(result), tt.wantLen)
			}
		})
	}
}

func TestParseVectorFromText(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantLen   int
		wantError bool
	}{
		{
			name:      "empty vector",
			input:     "[]",
			wantLen:   0,
			wantError: false,
		},
		{
			name:      "single element",
			input:     "[0.5]",
			wantLen:   1,
			wantError: false,
		},
		{
			name:      "multiple elements",
			input:     "[0.1, 0.2, 0.3]",
			wantLen:   3,
			wantError: false,
		},
		{
			name:      "invalid format no brackets",
			input:     "0.1, 0.2",
			wantLen:   0,
			wantError: true,
		},
		{
			name:      "invalid number",
			input:     "[0.1, abc]",
			wantLen:   0,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseVectorFromText(tt.input)
			if (err != nil) != tt.wantError {
				t.Errorf("parseVectorFromText(%q) error = %v, wantError %v", tt.input, err, tt.wantError)
				return
			}
			if !tt.wantError && len(result) != tt.wantLen {
				t.Errorf("parseVectorFromText(%q) got len %d, want %d", tt.input, len(result), tt.wantLen)
			}
		})
	}
}

func TestNewEnumMeaningContent(t *testing.T) {
	values := map[string]string{
		"0": "inactive",
		"1": "active",
		"2": "suspended",
	}

	content, err := NewEnumMeaningContent(values)
	if err != nil {
		t.Fatalf("NewEnumMeaningContent() error = %v", err)
	}

	// Parse back and verify
	parsed, err := ParseEnumMeaningContent(content)
	if err != nil {
		t.Fatalf("ParseEnumMeaningContent() error = %v", err)
	}

	if len(parsed.Values) != len(values) {
		t.Errorf("got %d values, want %d", len(parsed.Values), len(values))
	}

	for k, v := range values {
		if parsed.Values[k] != v {
			t.Errorf("Values[%q] = %q, want %q", k, parsed.Values[k], v)
		}
	}
}

func TestNewBusinessRuleContent(t *testing.T) {
	rules := []string{"Customer must have valid email", "Order total must be positive"}
	constraints := []string{"NOT NULL", "CHECK (total > 0)"}

	content, err := NewBusinessRuleContent(rules, constraints)
	if err != nil {
		t.Fatalf("NewBusinessRuleContent() error = %v", err)
	}

	parsed, err := ParseBusinessRuleContent(content)
	if err != nil {
		t.Fatalf("ParseBusinessRuleContent() error = %v", err)
	}

	if len(parsed.Rules) != len(rules) {
		t.Errorf("got %d rules, want %d", len(parsed.Rules), len(rules))
	}

	if len(parsed.Constraints) != len(constraints) {
		t.Errorf("got %d constraints, want %d", len(parsed.Constraints), len(constraints))
	}
}

func TestNewJoinHintContent(t *testing.T) {
	relatedTables := []string{"orders", "customers"}
	joinKeys := []string{"customer_id"}
	description := "One-to-many relationship between customers and orders"

	content, err := NewJoinHintContent(relatedTables, joinKeys, description)
	if err != nil {
		t.Fatalf("NewJoinHintContent() error = %v", err)
	}

	parsed, err := ParseJoinHintContent(content)
	if err != nil {
		t.Fatalf("ParseJoinHintContent() error = %v", err)
	}

	if len(parsed.RelatedTables) != len(relatedTables) {
		t.Errorf("got %d related tables, want %d", len(parsed.RelatedTables), len(relatedTables))
	}

	if parsed.Description != description {
		t.Errorf("Description = %q, want %q", parsed.Description, description)
	}
}

func TestNewSemanticContent(t *testing.T) {
	description := "Customer identifier"
	synonyms := []string{"client_id", "user_id"}
	businessTerms := []string{"customer", "client"}

	content, err := NewSemanticContent(description, synonyms, businessTerms)
	if err != nil {
		t.Fatalf("NewSemanticContent() error = %v", err)
	}

	parsed, err := ParseSemanticContent(content)
	if err != nil {
		t.Fatalf("ParseSemanticContent() error = %v", err)
	}

	if parsed.Description != description {
		t.Errorf("Description = %q, want %q", parsed.Description, description)
	}

	if len(parsed.Synonyms) != len(synonyms) {
		t.Errorf("got %d synonyms, want %d", len(parsed.Synonyms), len(synonyms))
	}
}

func TestContextTypes(t *testing.T) {
	tests := []struct {
		contextType ContextType
		expected    string
	}{
		{ContextTypeEnumMeaning, "enum_meaning"},
		{ContextTypeBusinessRule, "business_rule"},
		{ContextTypeJoinHint, "join_hint"},
		{ContextTypeDataQuality, "data_quality"},
		{ContextTypeSemantic, "semantic"},
	}

	for _, tt := range tests {
		t.Run(string(tt.contextType), func(t *testing.T) {
			if string(tt.contextType) != tt.expected {
				t.Errorf("ContextType = %q, want %q", tt.contextType, tt.expected)
			}
		})
	}
}

func TestContextSources(t *testing.T) {
	tests := []struct {
		source   ContextSource
		expected string
	}{
		{SourceLLM, "llm"},
		{SourceCatalog, "catalog"},
		{SourceUser, "user"},
		{SourceAutoCorrected, "auto_corrected"},
	}

	for _, tt := range tests {
		t.Run(string(tt.source), func(t *testing.T) {
			if string(tt.source) != tt.expected {
				t.Errorf("ContextSource = %q, want %q", tt.source, tt.expected)
			}
		})
	}
}

func TestEntityTypes(t *testing.T) {
	tests := []struct {
		entityType EntityType
		expected   string
	}{
		{EntityTypeTable, "table"},
		{EntityTypeColumn, "column"},
		{EntityTypeContext, "context"},
	}

	for _, tt := range tests {
		t.Run(string(tt.entityType), func(t *testing.T) {
			if string(tt.entityType) != tt.expected {
				t.Errorf("EntityType = %q, want %q", tt.entityType, tt.expected)
			}
		})
	}
}

func TestChangeTypes(t *testing.T) {
	tests := []struct {
		changeType ChangeType
		expected   string
	}{
		{ChangeTypeSchemaChange, "schema_change"},
		{ChangeTypeContextUpdate, "context_update"},
		{ChangeTypeContextExpire, "context_expire"},
	}

	for _, tt := range tests {
		t.Run(string(tt.changeType), func(t *testing.T) {
			if string(tt.changeType) != tt.expected {
				t.Errorf("ChangeType = %q, want %q", tt.changeType, tt.expected)
			}
		})
	}
}

func TestDatasourceSerialization(t *testing.T) {
	ds := &Datasource{
		ID:           1,
		Name:         "test_db",
		DBType:       "mariadb",
		Host:         "localhost",
		Port:         3310,
		Username:     "root",
		Password:     "secret", // Should be hidden
		DatabaseName: "testdb",
		Status:       DatasourceStatusEnabled,
	}

	data, err := json.Marshal(ds)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	// Password should not appear in JSON
	jsonStr := string(data)
	if contains(jsonStr, "secret") {
		t.Error("Password should not be in JSON output")
	}

	// Verify other fields are present
	if !contains(jsonStr, "test_db") {
		t.Error("Name should be in JSON output")
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Host != "127.0.0.1" {
		t.Errorf("Default Host = %q, want %q", cfg.Host, "127.0.0.1")
	}
	if cfg.Port != 3310 {
		t.Errorf("Default Port = %d, want %d", cfg.Port, 3310)
	}
	if cfg.MaxOpenConns != 20 {
		t.Errorf("Default MaxOpenConns = %d, want %d", cfg.MaxOpenConns, 20)
	}
}

func TestDefaultLakebaseConfig(t *testing.T) {
	cfg := DefaultLakebaseConfig()

	if cfg.Lakebase.Host != "127.0.0.1" {
		t.Errorf("Default Lakebase.Host = %q, want %q", cfg.Lakebase.Host, "127.0.0.1")
	}
	if cfg.Embedding.Dimension != 1536 {
		t.Errorf("Default Embedding.Dimension = %d, want %d", cfg.Embedding.Dimension, 1536)
	}
	if cfg.VectorSearch.TopK != 10 {
		t.Errorf("Default VectorSearch.TopK = %d, want %d", cfg.VectorSearch.TopK, 10)
	}
	if !cfg.Agent.EnableDDLDetection {
		t.Error("Default Agent.EnableDDLDetection should be true")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
