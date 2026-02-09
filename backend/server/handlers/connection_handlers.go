package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
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

	for _, db := range h.config.Databases {
		if db.ID == conn.ID {
			c.JSON(http.StatusConflict, gin.H{"error": "Connection ID already exists"})
			return
		}
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
	h.config.Databases = append(h.config.Databases, newDB)

	if _, err := h.dbService.GetAdapter(conn.ID); err != nil {
		h.config.Databases = h.config.Databases[:len(h.config.Databases)-1]
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Failed to connect to database: %v", err),
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

	found := false
	newDatabases := make([]config.DatabaseConfig, 0, len(h.config.Databases))
	for _, db := range h.config.Databases {
		if db.ID == connID {
			found = true
			h.dbService.CloseAdapter(connID)
		} else {
			newDatabases = append(newDatabases, db)
		}
	}

	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "Connection not found"})
		return
	}

	h.config.Databases = newDatabases
	c.JSON(http.StatusOK, gin.H{"message": "Connection removed successfully", "id": connID})
}

// SyncConnectionSchema creates/updates an rc_datasources record and syncs schema.
func (h *Handler) SyncConnectionSchema(c *gin.Context) {
	connID := c.Param("id")

	if h.lakebaseService == nil || !h.lakebaseService.IsConnected() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Lake-base service not available"})
		return
	}

	var dbCfg *config.DatabaseConfig
	for _, db := range h.config.Databases {
		if db.ID == connID {
			dbCfg = &db
			break
		}
	}
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
	connections := make([]map[string]interface{}, 0, len(h.config.Databases))
	for _, db := range h.config.Databases {
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

// ReleaseAllDemoConnections releases all demo database connections.
func (h *Handler) ReleaseAllDemoConnections(c *gin.Context) {
	demoConfigPath := os.Getenv("DEMO_DATABASES_PATH")
	if demoConfigPath == "" {
		demoConfigPath = "demo_databases.json"
	}

	var demoIDs []string
	data, err := os.ReadFile(demoConfigPath)
	if err == nil {
		var demoConfig struct {
			Databases []struct {
				ID          string                       `json:"id"`
				Connections map[string]map[string]string `json:"connections"`
			} `json:"databases"`
		}
		if json.Unmarshal(data, &demoConfig) == nil {
			for _, db := range demoConfig.Databases {
				for connType := range db.Connections {
					demoIDs = append(demoIDs, fmt.Sprintf("%s_%s", db.ID, connType))
				}
			}
		}
	}

	if len(demoIDs) == 0 {
		for _, db := range h.config.Databases {
			if strings.HasSuffix(db.ID, "_mysql") ||
				strings.HasSuffix(db.ID, "_sqlite") ||
				strings.HasSuffix(db.ID, "_postgres") ||
				strings.HasSuffix(db.ID, "_postgresql") {
				demoIDs = append(demoIDs, db.ID)
			}
		}
	}

	releasedCount := 0
	releasedIDs := []string{}
	newDatabases := make([]config.DatabaseConfig, 0, len(h.config.Databases))

	for _, db := range h.config.Databases {
		isDemo := false
		for _, demoID := range demoIDs {
			if db.ID == demoID {
				isDemo = true
				break
			}
		}
		if isDemo {
			h.dbService.CloseAdapter(db.ID)
			releasedCount++
			releasedIDs = append(releasedIDs, db.ID)
		} else {
			newDatabases = append(newDatabases, db)
		}
	}

	h.config.Databases = newDatabases
	c.JSON(http.StatusOK, gin.H{
		"message":  fmt.Sprintf("Released %d demo database connections", releasedCount),
		"released": releasedCount,
		"ids":      releasedIDs,
	})
}

// ListAvailableConnections is deprecated.
func (h *Handler) ListAvailableConnections(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"available": []interface{}{}, "count": 0})
}

// LoadDemoDatabases is deprecated.
func (h *Handler) LoadDemoDatabases(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "deprecated", "added": 0})
}
