package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"lucid/internal/config"
	"lucid/internal/lakebase"
	"lucid/server/services"
)

// ConnectionConfig represents database connection configuration.
type ConnectionConfig struct {
	ID       string `json:"id" binding:"required"`
	Name     string `json:"name" binding:"required"`
	Type     string `json:"type" binding:"required"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
	Path     string `json:"path"`
}

// ConnectionStatus represents connection test result.
type ConnectionStatus struct {
	ID        string `json:"id"`
	Connected bool   `json:"connected"`
	Message   string `json:"message"`
	Version   string `json:"version,omitempty"`
	Latency   int64  `json:"latency_ms"`
}

// AddConnection adds a new database connection.
func (h *Handler) AddConnection(c *gin.Context) {
	var conn ConnectionConfig
	if err := c.ShouldBindJSON(&conn); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if conn.Type != "mysql" && conn.Type != "mariadb" && conn.Type != "postgresql" && conn.Type != "sqlite" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid database type. Must be mysql, mariadb, postgresql, or sqlite",
		})
		return
	}

	adapterType := conn.Type
	if adapterType == "mariadb" {
		adapterType = "mysql"
	}

	newDB := config.DatabaseConfig{
		ID:       conn.ID,
		Name:     conn.Name,
		Type:     adapterType,
		Host:     conn.Host,
		Port:     conn.Port,
		User:     conn.User,
		Password: conn.Password,
		Database: conn.Database,
		Path:     conn.Path,
	}

	if err := h.dbService.AddDatabase(newDB); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Failed to add connection: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Connection added successfully",
		"connection": conn,
	})
}

// RemoveConnection removes a database connection.
func (h *Handler) RemoveConnection(c *gin.Context) {
	connID := c.Param("id")

	if !h.dbService.RemoveDatabase(connID) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Connection not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Connection removed successfully", "id": connID})
}

// SyncConnectionSchema creates/updates an rc_datasources record and syncs schema.
func (h *Handler) SyncConnectionSchema(c *gin.Context) {
	connID := c.Param("id")

	if h.lakebaseService == nil || !h.lakebaseService.IsConnected() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Lake-base service not available"})
		return
	}

	dbCfg := h.dbService.FindDatabase(connID)
	if dbCfg == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Connection not found"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Minute)
	defer cancel()

	ds, err := h.lakebaseService.GetOrCreateDatasource(ctx, &lakebase.Datasource{
		Name:         dbCfg.ID,
		DBType:       dbCfg.Type,
		Host:         sql.NullString{String: dbCfg.Host, Valid: dbCfg.Host != ""},
		Port:         sql.NullInt32{Int32: int32(dbCfg.Port), Valid: dbCfg.Port > 0},
		Username:     sql.NullString{String: dbCfg.User, Valid: dbCfg.User != ""},
		DatabaseName: sql.NullString{String: dbCfg.Database, Valid: dbCfg.Database != ""},
		Status:       "active",
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to register datasource: %v", err)})
		return
	}

	adapter, err := h.dbService.GetAdapter(connID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Cannot connect to database: %v", err)})
		return
	}

	result, err := h.lakebaseService.SyncSchema(ctx, ds.ID, adapter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Schema sync failed: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"datasource_id": ds.ID,
		"tables":        result.TablesCount,
		"columns":       result.ColumnsCount,
		"relations":     result.RelationsCount,
	})
}

// TestConnection tests a database connection.
func (h *Handler) TestConnection(c *gin.Context) {
	var conn ConnectionConfig
	if err := c.ShouldBindJSON(&conn); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	startTime := time.Now()
	status := ConnectionStatus{ID: conn.ID}

	adapter, err := h.dbService.CreateCustomAdapter(&services.AdapterConfig{
		Type:     conn.Type,
		Host:     conn.Host,
		Port:     conn.Port,
		User:     conn.User,
		Password: conn.Password,
		Database: conn.Database,
		Path:     conn.Path,
	})
	if err != nil {
		status.Connected = false
		status.Message = fmt.Sprintf("Failed to create adapter: %v", err)
		status.Latency = time.Since(startTime).Milliseconds()
		c.JSON(http.StatusOK, status)
		return
	}
	defer adapter.Close()

	if err := adapter.Connect(ctx); err != nil {
		status.Connected = false
		status.Message = fmt.Sprintf("Failed to connect: %v", err)
		status.Latency = time.Since(startTime).Milliseconds()
		c.JSON(http.StatusOK, status)
		return
	}

	version, err := adapter.GetDatabaseVersion(ctx)
	if err != nil {
		status.Connected = true
		status.Message = "Connected successfully (version query failed)"
		status.Latency = time.Since(startTime).Milliseconds()
		c.JSON(http.StatusOK, status)
		return
	}

	status.Connected = true
	status.Message = "Connected successfully"
	status.Version = version
	status.Latency = time.Since(startTime).Milliseconds()
	c.JSON(http.StatusOK, status)
}

// ListConnections returns all configured connections with status.
func (h *Handler) ListConnections(c *gin.Context) {
	databases := h.dbService.ListDatabases()
	connections := make([]map[string]interface{}, 0, len(databases))
	for _, db := range databases {
		conn := map[string]interface{}{
			"id":       db.ID,
			"name":     db.Name,
			"type":     db.Type,
			"host":     db.Host,
			"port":     db.Port,
			"database": db.Database,
			"user":     db.User,
		}
		if db.Type == "sqlite" {
			conn["path"] = db.Path
		}
		connections = append(connections, conn)
	}
	c.JSON(http.StatusOK, gin.H{"connections": connections, "count": len(connections)})
}
