package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// DatabaseInfo represents database information for API response.
type DatabaseInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// ListDatabases returns all configured databases.
func (h *Handler) ListDatabases(c *gin.Context) {
	databases := make([]DatabaseInfo, 0, len(h.config.Databases))
	for _, db := range h.config.Databases {
		databases = append(databases, DatabaseInfo{
			ID:   db.ID,
			Name: db.Name,
			Type: db.Type,
		})
	}
	c.JSON(http.StatusOK, gin.H{"databases": databases})
}

// GetDatabaseSchema returns schema information for a database.
func (h *Handler) GetDatabaseSchema(c *gin.Context) {
	dbID := c.Param("id")

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	schema, err := h.dbService.GetSchema(ctx, dbID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, schema)
}

// GetDatabaseTables returns list of tables for a database.
func (h *Handler) GetDatabaseTables(c *gin.Context) {
	dbID := c.Param("id")

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	schema, err := h.dbService.GetSchema(ctx, dbID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	tables := make([]string, len(schema.Tables))
	for i, t := range schema.Tables {
		tables[i] = t.Name
	}
	c.JSON(http.StatusOK, gin.H{"database_id": dbID, "tables": tables})
}

// GetRichContext returns rich context for a database.
// Redirects to lakebase-based context; legacy file-based RC is removed.
func (h *Handler) GetRichContext(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Use /api/v1/lakebase/datasources/:id for rich context",
	})
}

// ExecuteSQL executes SQL on a database.
func (h *Handler) ExecuteSQL(c *gin.Context) {
	dbID := c.Param("id")

	var req struct {
		SQL string `json:"sql" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	result, err := h.dbService.ExecuteSQL(ctx, dbID, req.SQL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"database_id": dbID,
		"sql":         req.SQL,
		"result":      result,
	})
}
