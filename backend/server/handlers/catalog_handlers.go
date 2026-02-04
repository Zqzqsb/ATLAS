package handlers

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"lucid/internal/context"
)

// CatalogUploadRequest represents a catalog upload request body
type CatalogUploadRequest struct {
	ConnectionID string `json:"connection_id" binding:"required"`
	Format       string `json:"format"` // "yaml" or "json", auto-detected if empty
	Content      string `json:"content" binding:"required"`
}

// CatalogImportResult represents the result of a catalog import
type CatalogImportResult struct {
	Success         bool                   `json:"success"`
	Message         string                 `json:"message"`
	DatabaseMatched string                 `json:"database_matched,omitempty"`
	MergeResult     *context.MergeResult   `json:"merge_result,omitempty"`
	ValidationErrs  []string               `json:"validation_errors,omitempty"`
}

// UploadCatalog handles catalog file upload and import
// POST /api/catalog/upload
func (h *Handler) UploadCatalog(c *gin.Context) {
	var req CatalogUploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, CatalogImportResult{
			Success: false,
			Message: fmt.Sprintf("Invalid request: %v", err),
		})
		return
	}

	// Get the connection's Rich Context
	richContext, err := h.dbService.GetRichContext(req.ConnectionID)
	if err != nil {
		c.JSON(http.StatusNotFound, CatalogImportResult{
			Success: false,
			Message: fmt.Sprintf("Connection not found or no Rich Context available: %v", err),
		})
		return
	}

	// Create catalog importer
	importer := context.NewCatalogImporter()

	// Parse catalog based on format
	var catalog *context.CatalogSchema
	format := strings.ToLower(req.Format)
	
	// Auto-detect format if not specified
	if format == "" {
		content := strings.TrimSpace(req.Content)
		if strings.HasPrefix(content, "{") {
			format = "json"
		} else {
			format = "yaml"
		}
	}

	switch format {
	case "yaml", "yml":
		catalog, err = importer.ImportFromYAML(req.Content)
	case "json":
		catalog, err = importer.ImportFromJSON(req.Content)
	default:
		c.JSON(http.StatusBadRequest, CatalogImportResult{
			Success: false,
			Message: fmt.Sprintf("Unsupported format: %s. Use 'yaml' or 'json'", req.Format),
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusBadRequest, CatalogImportResult{
			Success: false,
			Message: fmt.Sprintf("Failed to parse catalog: %v", err),
		})
		return
	}

	// Validate catalog
	validationErrors := context.ValidateCatalog(catalog)
	if len(validationErrors) > 0 {
		c.JSON(http.StatusBadRequest, CatalogImportResult{
			Success:        false,
			Message:        "Catalog validation failed",
			ValidationErrs: validationErrors,
		})
		return
	}

	// Merge catalog into Rich Context
	mergeResult := importer.MergeIntoContext(richContext, catalog)

	// Save updated Rich Context
	if err := h.dbService.SaveRichContext(req.ConnectionID, richContext); err != nil {
		c.JSON(http.StatusInternalServerError, CatalogImportResult{
			Success: false,
			Message: fmt.Sprintf("Failed to save updated Rich Context: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, CatalogImportResult{
		Success:         true,
		Message:         fmt.Sprintf("Successfully imported catalog for database '%s'", catalog.Database),
		DatabaseMatched: catalog.Database,
		MergeResult:     mergeResult,
	})
}

// UploadCatalogFile handles multipart file upload for catalog
// POST /api/catalog/upload-file
func (h *Handler) UploadCatalogFile(c *gin.Context) {
	connectionID := c.PostForm("connection_id")
	if connectionID == "" {
		c.JSON(http.StatusBadRequest, CatalogImportResult{
			Success: false,
			Message: "connection_id is required",
		})
		return
	}

	// Get uploaded file
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, CatalogImportResult{
			Success: false,
			Message: fmt.Sprintf("Failed to get uploaded file: %v", err),
		})
		return
	}
	defer file.Close()

	// Read file content
	content, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusBadRequest, CatalogImportResult{
			Success: false,
			Message: fmt.Sprintf("Failed to read file: %v", err),
		})
		return
	}

	// Detect format from filename
	ext := strings.ToLower(filepath.Ext(header.Filename))
	var format string
	switch ext {
	case ".yaml", ".yml":
		format = "yaml"
	case ".json":
		format = "json"
	default:
		c.JSON(http.StatusBadRequest, CatalogImportResult{
			Success: false,
			Message: fmt.Sprintf("Unsupported file extension: %s. Use .yaml, .yml, or .json", ext),
		})
		return
	}

	// Forward to main upload handler logic
	req := CatalogUploadRequest{
		ConnectionID: connectionID,
		Format:       format,
		Content:      string(content),
	}

	h.processCatalogUpload(c, req)
}

// processCatalogUpload is the shared logic for catalog upload
func (h *Handler) processCatalogUpload(c *gin.Context, req CatalogUploadRequest) {
	// Get the connection's Rich Context
	richContext, err := h.dbService.GetRichContext(req.ConnectionID)
	if err != nil {
		c.JSON(http.StatusNotFound, CatalogImportResult{
			Success: false,
			Message: fmt.Sprintf("Connection not found or no Rich Context available: %v", err),
		})
		return
	}

	// Create catalog importer
	importer := context.NewCatalogImporter()

	// Parse catalog based on format
	var catalog *context.CatalogSchema
	switch req.Format {
	case "yaml", "yml":
		catalog, err = importer.ImportFromYAML(req.Content)
	case "json":
		catalog, err = importer.ImportFromJSON(req.Content)
	default:
		c.JSON(http.StatusBadRequest, CatalogImportResult{
			Success: false,
			Message: fmt.Sprintf("Unsupported format: %s", req.Format),
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusBadRequest, CatalogImportResult{
			Success: false,
			Message: fmt.Sprintf("Failed to parse catalog: %v", err),
		})
		return
	}

	// Validate catalog
	validationErrors := context.ValidateCatalog(catalog)
	if len(validationErrors) > 0 {
		c.JSON(http.StatusBadRequest, CatalogImportResult{
			Success:        false,
			Message:        "Catalog validation failed",
			ValidationErrs: validationErrors,
		})
		return
	}

	// Merge catalog into Rich Context
	mergeResult := importer.MergeIntoContext(richContext, catalog)

	// Save updated Rich Context
	if err := h.dbService.SaveRichContext(req.ConnectionID, richContext); err != nil {
		c.JSON(http.StatusInternalServerError, CatalogImportResult{
			Success: false,
			Message: fmt.Sprintf("Failed to save updated Rich Context: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, CatalogImportResult{
		Success:         true,
		Message:         fmt.Sprintf("Successfully imported catalog for database '%s'", catalog.Database),
		DatabaseMatched: catalog.Database,
		MergeResult:     mergeResult,
	})
}

// GetCatalogTemplate returns an example catalog template
// GET /api/catalog/template
func (h *Handler) GetCatalogTemplate(c *gin.Context) {
	format := c.DefaultQuery("format", "yaml")
	
	template := context.GenerateCatalogTemplate()
	
	var output string
	var err error
	var contentType string
	
	switch strings.ToLower(format) {
	case "yaml", "yml":
		output, err = context.ExportCatalogToYAML(template)
		contentType = "text/yaml"
	case "json":
		output, err = context.ExportCatalogToJSON(template)
		contentType = "application/json"
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Unsupported format: %s", format),
		})
		return
	}
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to generate template: %v", err),
		})
		return
	}
	
	c.Header("Content-Type", contentType)
	c.String(http.StatusOK, output)
}

// GetCatalogSchema returns the JSON schema for catalog validation
// GET /api/catalog/schema
func (h *Handler) GetCatalogSchema(c *gin.Context) {
	schema := map[string]interface{}{
		"$schema":     "http://json-schema.org/draft-07/schema#",
		"title":       "ReActSQL Catalog Schema",
		"description": "Schema for ReActSQL database catalog files",
		"type":        "object",
		"required":    []string{"version", "database", "tables"},
		"properties": map[string]interface{}{
			"version": map[string]interface{}{
				"type":        "string",
				"description": "Catalog schema version",
				"default":     "1.0",
			},
			"database": map[string]interface{}{
				"type":        "string",
				"description": "Database name that this catalog describes",
			},
			"description": map[string]interface{}{
				"type":        "string",
				"description": "Brief description of the database",
			},
			"tables": map[string]interface{}{
				"type":        "array",
				"description": "List of table definitions",
				"items": map[string]interface{}{
					"type":     "object",
					"required": []string{"name"},
					"properties": map[string]interface{}{
						"name": map[string]interface{}{
							"type":        "string",
							"description": "Table name",
						},
						"description": map[string]interface{}{
							"type":        "string",
							"description": "Table description",
						},
						"owner": map[string]interface{}{
							"type":        "string",
							"description": "Team or person responsible for this table",
						},
						"tags": map[string]interface{}{
							"type":        "array",
							"items":       map[string]interface{}{"type": "string"},
							"description": "Tags for categorization",
						},
						"business_rules": map[string]interface{}{
							"type":        "array",
							"items":       map[string]interface{}{"type": "string"},
							"description": "Business rules that apply to this table",
						},
						"quality_notes": map[string]interface{}{
							"type":        "array",
							"items":       map[string]interface{}{"type": "string"},
							"description": "Data quality notes",
						},
						"columns": map[string]interface{}{
							"type":        "array",
							"description": "Column definitions",
							"items": map[string]interface{}{
								"type":     "object",
								"required": []string{"name"},
								"properties": map[string]interface{}{
									"name": map[string]interface{}{
										"type": "string",
									},
									"description": map[string]interface{}{
										"type": "string",
									},
									"business_values": map[string]interface{}{
										"type":        "array",
										"description": "Enum value mappings",
										"items": map[string]interface{}{
											"type":     "object",
											"required": []string{"value", "meaning"},
											"properties": map[string]interface{}{
												"value": map[string]interface{}{
													"type": "string",
												},
												"meaning": map[string]interface{}{
													"type": "string",
												},
											},
										},
									},
									"business_rules": map[string]interface{}{
										"type":  "array",
										"items": map[string]interface{}{"type": "string"},
									},
									"data_quality_notes": map[string]interface{}{
										"type": "string",
									},
									"pii": map[string]interface{}{
										"type":        "boolean",
										"description": "Whether this column contains PII",
									},
									"format": map[string]interface{}{
										"type":        "string",
										"description": "Data format description",
									},
									"examples": map[string]interface{}{
										"type":        "array",
										"items":       map[string]interface{}{"type": "string"},
										"description": "Example values",
									},
								},
							},
						},
					},
				},
			},
			"global_notes": map[string]interface{}{
				"type":        "array",
				"items":       map[string]interface{}{"type": "string"},
				"description": "Global notes that apply to all tables",
			},
		},
	}

	c.JSON(http.StatusOK, schema)
}

// ExportCatalogFromContext exports current Rich Context as a catalog file
// GET /api/catalog/export/:connection_id
func (h *Handler) ExportCatalogFromContext(c *gin.Context) {
	connectionID := c.Param("connection_id")
	format := c.DefaultQuery("format", "yaml")

	richContext, err := h.dbService.GetRichContext(connectionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": fmt.Sprintf("Rich Context not found: %v", err),
		})
		return
	}

	// Convert Rich Context to Catalog format
	catalog := convertRichContextToCatalog(richContext)

	var output string
	var contentType string
	var filename string

	switch strings.ToLower(format) {
	case "yaml", "yml":
		output, err = context.ExportCatalogToYAML(catalog)
		contentType = "text/yaml"
		filename = fmt.Sprintf("%s_catalog.yaml", richContext.DatabaseName)
	case "json":
		output, err = context.ExportCatalogToJSON(catalog)
		contentType = "application/json"
		filename = fmt.Sprintf("%s_catalog.json", richContext.DatabaseName)
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Unsupported format: %s", format),
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to export catalog: %v", err),
		})
		return
	}

	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.String(http.StatusOK, output)
}

// convertRichContextToCatalog converts SharedContext to CatalogSchema
func convertRichContextToCatalog(ctx *context.SharedContext) *context.CatalogSchema {
	catalog := &context.CatalogSchema{
		Version:     "1.0",
		Database:    ctx.DatabaseName,
		Description: fmt.Sprintf("Exported catalog from %s database", ctx.DatabaseName),
		Tables:      []context.CatalogTable{},
	}

	for tableName, table := range ctx.Tables {
		catalogTable := context.CatalogTable{
			Name:        tableName,
			Description: table.Description,
			Columns:     []context.CatalogColumn{},
		}

		// Convert Rich Context to column definitions
		for key, value := range table.RichContext {
			// Parse column-level entries
			if strings.Contains(key, "_enum_meaning") {
				colName := strings.TrimSuffix(key, "_enum_meaning")
				col := findOrCreateColumn(&catalogTable, colName)
				col.BusinessValues = parseEnumMeanings(value.Content)
			} else if strings.Contains(key, "_description") {
				colName := strings.TrimSuffix(key, "_description")
				col := findOrCreateColumn(&catalogTable, colName)
				col.Description = value.Content
			} else if strings.Contains(key, "_business_rules") {
				colName := strings.TrimSuffix(key, "_business_rules")
				col := findOrCreateColumn(&catalogTable, colName)
				col.BusinessRules = strings.Split(value.Content, "; ")
			} else if strings.Contains(key, "_quality_notes") {
				colName := strings.TrimSuffix(key, "_quality_notes")
				col := findOrCreateColumn(&catalogTable, colName)
				col.DataQualityNotes = value.Content
			} else if key == "business_rules" {
				catalogTable.BusinessRules = strings.Split(value.Content, "; ")
			} else if key == "tags" {
				catalogTable.Tags = strings.Split(value.Content, ", ")
			} else if key == "owner" {
				catalogTable.Owner = value.Content
			}
		}

		catalog.Tables = append(catalog.Tables, catalogTable)
	}

	return catalog
}

// findOrCreateColumn finds or creates a column in the catalog table
func findOrCreateColumn(table *context.CatalogTable, name string) *context.CatalogColumn {
	for i := range table.Columns {
		if table.Columns[i].Name == name {
			return &table.Columns[i]
		}
	}
	table.Columns = append(table.Columns, context.CatalogColumn{Name: name})
	return &table.Columns[len(table.Columns)-1]
}

// parseEnumMeanings parses "value1=meaning1, value2=meaning2" format
func parseEnumMeanings(content string) []context.BusinessValueMapping {
	// Remove source marker if present
	content = strings.TrimSuffix(content, "]")
	if idx := strings.LastIndex(content, " [source:"); idx > 0 {
		content = content[:idx]
	}

	var mappings []context.BusinessValueMapping
	parts := strings.Split(content, ", ")
	for _, part := range parts {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) == 2 {
			mappings = append(mappings, context.BusinessValueMapping{
				Value:   strings.TrimSpace(kv[0]),
				Meaning: strings.TrimSpace(kv[1]),
			})
		}
	}
	return mappings
}
