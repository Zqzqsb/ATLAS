package context

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// CatalogSource represents the source of a Rich Context entry
type CatalogSource string

const (
	SourceCatalog       CatalogSource = "catalog"       // From user-provided catalog
	SourceLLM           CatalogSource = "llm"           // From LLM inference
	SourceUser          CatalogSource = "user"          // Manually added by user
	SourceAutoCorrected CatalogSource = "auto_corrected" // Auto-corrected during runtime
	SourceAnalysis      CatalogSource = "analysis"      // From onboarding analysis
)

// EnhancedBusinessNote extends BusinessNote with more metadata for self-maintenance
type EnhancedBusinessNote struct {
	Content    string        `json:"content"`               // Business content
	ExpiresAt  string        `json:"expires_at,omitempty"`  // Expiration time (ISO 8601)
	Confidence float64       `json:"confidence,omitempty"`  // Confidence score (0-1)
	Source     CatalogSource `json:"source,omitempty"`      // Source of the information
	CreatedAt  time.Time     `json:"created_at,omitempty"`  // Creation time
	UpdatedAt  time.Time     `json:"updated_at,omitempty"`  // Last update time
	UpdatedBy  string        `json:"updated_by,omitempty"`  // Who/what updated it
	Reason     string        `json:"reason,omitempty"`      // Reason for update
	Version    int           `json:"version,omitempty"`     // Version number
}

// CatalogSchema represents the root structure of a catalog file
type CatalogSchema struct {
	Version      string         `json:"version" yaml:"version"`
	Database     string         `json:"database" yaml:"database"`
	Description  string         `json:"description,omitempty" yaml:"description,omitempty"`
	Tables       []CatalogTable `json:"tables" yaml:"tables"`
	GlobalNotes  []string       `json:"global_notes,omitempty" yaml:"global_notes,omitempty"`
	ImportedAt   time.Time      `json:"imported_at,omitempty" yaml:"-"`
	ImportedFrom string         `json:"imported_from,omitempty" yaml:"-"`
}

// CatalogTable represents a table definition in the catalog
type CatalogTable struct {
	Name        string          `json:"name" yaml:"name"`
	Description string          `json:"description,omitempty" yaml:"description,omitempty"`
	Owner       string          `json:"owner,omitempty" yaml:"owner,omitempty"`
	Tags        []string        `json:"tags,omitempty" yaml:"tags,omitempty"`
	Columns     []CatalogColumn `json:"columns,omitempty" yaml:"columns,omitempty"`
	// Business rules that apply to this table
	BusinessRules []string `json:"business_rules,omitempty" yaml:"business_rules,omitempty"`
	// Data quality notes
	QualityNotes []string `json:"quality_notes,omitempty" yaml:"quality_notes,omitempty"`
}

// CatalogColumn represents a column definition in the catalog
type CatalogColumn struct {
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	// Business values mapping (for enum columns)
	BusinessValues []BusinessValueMapping `json:"business_values,omitempty" yaml:"business_values,omitempty"`
	// Business rules for this column
	BusinessRules []string `json:"business_rules,omitempty" yaml:"business_rules,omitempty"`
	// Data quality notes
	DataQualityNotes string `json:"data_quality_notes,omitempty" yaml:"data_quality_notes,omitempty"`
	// Whether this column contains PII
	PII bool `json:"pii,omitempty" yaml:"pii,omitempty"`
	// Data format description
	Format string `json:"format,omitempty" yaml:"format,omitempty"`
	// Example values
	Examples []string `json:"examples,omitempty" yaml:"examples,omitempty"`
}

// BusinessValueMapping maps database values to business meanings
type BusinessValueMapping struct {
	Value   string `json:"value" yaml:"value"`
	Meaning string `json:"meaning" yaml:"meaning"`
}

// CatalogImporter handles importing catalog data into Rich Context
type CatalogImporter struct {
	mergeStrategy MergeStrategy
}

// MergeStrategy defines how to merge catalog data with existing Rich Context
type MergeStrategy string

const (
	MergeCatalogFirst MergeStrategy = "catalog_first" // Catalog takes priority
	MergeLLMFirst     MergeStrategy = "llm_first"     // LLM inference takes priority
	MergeMostRecent   MergeStrategy = "most_recent"   // Most recent update wins
)

// NewCatalogImporter creates a new CatalogImporter with default strategy
func NewCatalogImporter() *CatalogImporter {
	return &CatalogImporter{
		mergeStrategy: MergeCatalogFirst, // Catalog is authoritative by default
	}
}

// SetMergeStrategy sets the merge strategy
func (ci *CatalogImporter) SetMergeStrategy(strategy MergeStrategy) {
	ci.mergeStrategy = strategy
}

// ImportFromFile imports catalog from a JSON or YAML file
func (ci *CatalogImporter) ImportFromFile(filepath string) (*CatalogSchema, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read catalog file: %w", err)
	}

	var catalog CatalogSchema

	// Detect format based on extension
	if strings.HasSuffix(strings.ToLower(filepath), ".yaml") || strings.HasSuffix(strings.ToLower(filepath), ".yml") {
		err = yaml.Unmarshal(data, &catalog)
	} else {
		err = json.Unmarshal(data, &catalog)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to parse catalog file: %w", err)
	}

	catalog.ImportedAt = time.Now()
	catalog.ImportedFrom = filepath

	return &catalog, nil
}

// ImportFromJSON imports catalog from JSON string
func (ci *CatalogImporter) ImportFromJSON(jsonData string) (*CatalogSchema, error) {
	var catalog CatalogSchema
	if err := json.Unmarshal([]byte(jsonData), &catalog); err != nil {
		return nil, fmt.Errorf("failed to parse catalog JSON: %w", err)
	}
	catalog.ImportedAt = time.Now()
	catalog.ImportedFrom = "json_upload"
	return &catalog, nil
}

// ImportFromYAML imports catalog from YAML string
func (ci *CatalogImporter) ImportFromYAML(yamlData string) (*CatalogSchema, error) {
	var catalog CatalogSchema
	if err := yaml.Unmarshal([]byte(yamlData), &catalog); err != nil {
		return nil, fmt.Errorf("failed to parse catalog YAML: %w", err)
	}
	catalog.ImportedAt = time.Now()
	catalog.ImportedFrom = "yaml_upload"
	return &catalog, nil
}

// MergeIntoContext merges catalog data into an existing SharedContext
func (ci *CatalogImporter) MergeIntoContext(ctx *SharedContext, catalog *CatalogSchema) *MergeResult {
	result := &MergeResult{
		TablesProcessed: 0,
		ColumnsUpdated:  0,
		NewEntries:      0,
		UpdatedEntries:  0,
		SkippedEntries:  0,
		Details:         []MergeDetail{},
	}

	ctx.mu.Lock()
	defer ctx.mu.Unlock()

	for _, catalogTable := range catalog.Tables {
		// Find matching table in context
		table, exists := ctx.Tables[catalogTable.Name]
		if !exists {
			// Table doesn't exist in context, skip but record
			result.Details = append(result.Details, MergeDetail{
				TableName: catalogTable.Name,
				Action:    "skipped",
				Reason:    "Table not found in database schema",
			})
			result.SkippedEntries++
			continue
		}

		result.TablesProcessed++

		// Merge table-level description
		if catalogTable.Description != "" {
			if table.Description == "" || ci.shouldOverwrite(SourceCatalog, SourceLLM) {
				table.Description = catalogTable.Description
				result.Details = append(result.Details, MergeDetail{
					TableName: catalogTable.Name,
					Field:     "description",
					Action:    "updated",
					Source:    string(SourceCatalog),
				})
				result.UpdatedEntries++
			}
		}

		// Initialize RichContext if nil
		if table.RichContext == nil {
			table.RichContext = make(map[string]RichContextValue)
		}

		// Merge business rules
		if len(catalogTable.BusinessRules) > 0 {
			ci.mergeRichContextEntry(table, "business_rules",
				strings.Join(catalogTable.BusinessRules, "; "),
				SourceCatalog, result)
		}

		// Merge quality notes
		if len(catalogTable.QualityNotes) > 0 {
			ci.mergeRichContextEntry(table, "quality_notes",
				strings.Join(catalogTable.QualityNotes, "; "),
				SourceCatalog, result)
		}

		// Merge tags
		if len(catalogTable.Tags) > 0 {
			ci.mergeRichContextEntry(table, "tags",
				strings.Join(catalogTable.Tags, ", "),
				SourceCatalog, result)
		}

		// Merge owner
		if catalogTable.Owner != "" {
			ci.mergeRichContextEntry(table, "owner", catalogTable.Owner, SourceCatalog, result)
		}

		// Merge column-level information
		for _, catalogCol := range catalogTable.Columns {
			result.ColumnsUpdated++

			// Column description
			if catalogCol.Description != "" {
				key := fmt.Sprintf("%s_description", catalogCol.Name)
				ci.mergeRichContextEntry(table, key, catalogCol.Description, SourceCatalog, result)
			}

			// Business values (enum meanings)
			if len(catalogCol.BusinessValues) > 0 {
				var mappings []string
				for _, bv := range catalogCol.BusinessValues {
					mappings = append(mappings, fmt.Sprintf("%s=%s", bv.Value, bv.Meaning))
				}
				key := fmt.Sprintf("%s_enum_meaning", catalogCol.Name)
				ci.mergeRichContextEntry(table, key, strings.Join(mappings, ", "), SourceCatalog, result)
			}

			// Business rules for column
			if len(catalogCol.BusinessRules) > 0 {
				key := fmt.Sprintf("%s_business_rules", catalogCol.Name)
				ci.mergeRichContextEntry(table, key, strings.Join(catalogCol.BusinessRules, "; "), SourceCatalog, result)
			}

			// Data quality notes
			if catalogCol.DataQualityNotes != "" {
				key := fmt.Sprintf("%s_quality_notes", catalogCol.Name)
				ci.mergeRichContextEntry(table, key, catalogCol.DataQualityNotes, SourceCatalog, result)
			}

			// PII marker
			if catalogCol.PII {
				key := fmt.Sprintf("%s_pii", catalogCol.Name)
				ci.mergeRichContextEntry(table, key, "⚠️ Contains PII - handle with care", SourceCatalog, result)
			}

			// Format
			if catalogCol.Format != "" {
				key := fmt.Sprintf("%s_format", catalogCol.Name)
				ci.mergeRichContextEntry(table, key, catalogCol.Format, SourceCatalog, result)
			}

			// Examples
			if len(catalogCol.Examples) > 0 {
				key := fmt.Sprintf("%s_examples", catalogCol.Name)
				ci.mergeRichContextEntry(table, key, strings.Join(catalogCol.Examples, ", "), SourceCatalog, result)
			}
		}
	}

	return result
}

// mergeRichContextEntry merges a single entry into table's RichContext
func (ci *CatalogImporter) mergeRichContextEntry(table *TableMetadata, key, value string, source CatalogSource, result *MergeResult) {
	existing, exists := table.RichContext[key]

	if !exists {
		// New entry
		table.RichContext[key] = RichContextValue{
			BusinessNote: BusinessNote{
				Content:   fmt.Sprintf("%s [source: %s]", value, source),
				ExpiresAt: "", // Catalog entries don't expire by default
			},
		}
		result.NewEntries++
		result.Details = append(result.Details, MergeDetail{
			TableName: table.Name,
			Field:     key,
			Action:    "created",
			Source:    string(source),
		})
	} else if ci.shouldOverwrite(source, ci.detectSource(existing.Content)) {
		// Update existing entry
		table.RichContext[key] = RichContextValue{
			BusinessNote: BusinessNote{
				Content:   fmt.Sprintf("%s [source: %s]", value, source),
				ExpiresAt: existing.ExpiresAt, // Preserve expiration
			},
		}
		result.UpdatedEntries++
		result.Details = append(result.Details, MergeDetail{
			TableName: table.Name,
			Field:     key,
			Action:    "updated",
			Source:    string(source),
		})
	} else {
		result.SkippedEntries++
	}
}

// shouldOverwrite determines if new source should overwrite existing source
func (ci *CatalogImporter) shouldOverwrite(newSource, existingSource CatalogSource) bool {
	switch ci.mergeStrategy {
	case MergeCatalogFirst:
		// Catalog always wins, user > catalog > llm
		priority := map[CatalogSource]int{
			SourceUser:          4,
			SourceCatalog:       3,
			SourceAutoCorrected: 2,
			SourceLLM:           1,
			SourceAnalysis:      1,
		}
		return priority[newSource] >= priority[existingSource]
	case MergeLLMFirst:
		// LLM wins (useful for experimentation)
		return newSource == SourceLLM
	case MergeMostRecent:
		// Always overwrite (handled elsewhere based on timestamp)
		return true
	default:
		return newSource == SourceCatalog
	}
}

// detectSource tries to detect the source from content marker
func (ci *CatalogImporter) detectSource(content string) CatalogSource {
	if strings.Contains(content, "[source: catalog]") {
		return SourceCatalog
	}
	if strings.Contains(content, "[source: user]") {
		return SourceUser
	}
	if strings.Contains(content, "[source: auto_corrected]") {
		return SourceAutoCorrected
	}
	return SourceLLM // Default to LLM
}

// MergeResult holds the result of a catalog merge operation
type MergeResult struct {
	TablesProcessed int           `json:"tables_processed"`
	ColumnsUpdated  int           `json:"columns_updated"`
	NewEntries      int           `json:"new_entries"`
	UpdatedEntries  int           `json:"updated_entries"`
	SkippedEntries  int           `json:"skipped_entries"`
	Details         []MergeDetail `json:"details"`
}

// MergeDetail holds details about a single merge operation
type MergeDetail struct {
	TableName string `json:"table_name"`
	Field     string `json:"field,omitempty"`
	Action    string `json:"action"`   // "created", "updated", "skipped"
	Source    string `json:"source,omitempty"`
	Reason    string `json:"reason,omitempty"`
}

// ValidateCatalog validates a catalog schema
func ValidateCatalog(catalog *CatalogSchema) []string {
	var errors []string

	if catalog.Database == "" {
		errors = append(errors, "database name is required")
	}

	if len(catalog.Tables) == 0 {
		errors = append(errors, "at least one table is required")
	}

	for i, table := range catalog.Tables {
		if table.Name == "" {
			errors = append(errors, fmt.Sprintf("table[%d]: name is required", i))
		}

		for j, col := range table.Columns {
			if col.Name == "" {
				errors = append(errors, fmt.Sprintf("table[%d].columns[%d]: name is required", i, j))
			}

			for k, bv := range col.BusinessValues {
				if bv.Value == "" {
					errors = append(errors, fmt.Sprintf("table[%d].columns[%d].business_values[%d]: value is required", i, j, k))
				}
				if bv.Meaning == "" {
					errors = append(errors, fmt.Sprintf("table[%d].columns[%d].business_values[%d]: meaning is required", i, j, k))
				}
			}
		}
	}

	return errors
}

// GenerateCatalogTemplate generates an example catalog template
func GenerateCatalogTemplate() *CatalogSchema {
	return &CatalogSchema{
		Version:     "1.0",
		Database:    "your_database_name",
		Description: "Brief description of the database purpose",
		Tables: []CatalogTable{
			{
				Name:        "orders",
				Description: "Customer order records",
				Owner:       "sales_team",
				Tags:        []string{"core", "transactional"},
				BusinessRules: []string{
					"Orders can only be cancelled within 24 hours",
					"Refund orders have negative amounts",
				},
				QualityNotes: []string{
					"Some historical data has NULL customer_id",
				},
				Columns: []CatalogColumn{
					{
						Name:        "status",
						Description: "Order status code",
						BusinessValues: []BusinessValueMapping{
							{Value: "0", Meaning: "Pending"},
							{Value: "1", Meaning: "Paid"},
							{Value: "2", Meaning: "Shipped"},
							{Value: "3", Meaning: "Completed"},
							{Value: "-1", Meaning: "Cancelled"},
						},
					},
					{
						Name:        "amount",
						Description: "Order amount in cents",
						BusinessRules: []string{
							"Divide by 100 to get actual currency value",
							"Negative values indicate refunds",
						},
					},
					{
						Name:             "customer_phone",
						Description:      "Customer phone number",
						PII:              true,
						Format:           "11-digit number, some with +86 prefix",
						DataQualityNotes: "May contain formatting inconsistencies",
					},
				},
			},
		},
		GlobalNotes: []string{
			"All timestamps are in UTC",
			"Soft deletes use is_deleted flag",
		},
	}
}

// ExportCatalogToYAML exports a catalog to YAML format
func ExportCatalogToYAML(catalog *CatalogSchema) (string, error) {
	data, err := yaml.Marshal(catalog)
	if err != nil {
		return "", fmt.Errorf("failed to marshal catalog to YAML: %w", err)
	}
	return string(data), nil
}

// ExportCatalogToJSON exports a catalog to JSON format
func ExportCatalogToJSON(catalog *CatalogSchema) (string, error) {
	data, err := json.MarshalIndent(catalog, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal catalog to JSON: %w", err)
	}
	return string(data), nil
}
