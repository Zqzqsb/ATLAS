package context

import (
	"fmt"
	"time"
)

// RichContextMaintainer handles self-maintenance of Rich Context
// Including expiration detection, validation, and auto-update tracking
type RichContextMaintainer struct {
	ctx *SharedContext

	// Configuration
	defaultExpiration time.Duration // Default expiration for new entries
	warningThreshold  time.Duration // Time before expiration to show warning
}

// MaintainerConfig holds configuration for RichContextMaintainer
type MaintainerConfig struct {
	DefaultExpiration time.Duration // Default: 30 days
	WarningThreshold  time.Duration // Default: 7 days
}

// NewRichContextMaintainer creates a new Rich Context maintainer
func NewRichContextMaintainer(ctx *SharedContext, config MaintainerConfig) *RichContextMaintainer {
	if config.DefaultExpiration == 0 {
		config.DefaultExpiration = 30 * 24 * time.Hour // 30 days
	}
	if config.WarningThreshold == 0 {
		config.WarningThreshold = 7 * 24 * time.Hour // 7 days
	}

	return &RichContextMaintainer{
		ctx:               ctx,
		defaultExpiration: config.DefaultExpiration,
		warningThreshold:  config.WarningThreshold,
	}
}

// ExpirationStatus represents the expiration status of a Rich Context entry
type ExpirationStatus int

const (
	StatusValid      ExpirationStatus = iota // Entry is valid
	StatusExpiringSoon                        // Entry will expire soon (within warning threshold)
	StatusExpired                             // Entry has expired
	StatusNoExpiry                            // Entry has no expiration set
)

// EntryStatus holds the status of a single Rich Context entry
type EntryStatus struct {
	TableName  string           `json:"table_name"`
	Key        string           `json:"key"`
	Content    string           `json:"content"`
	ExpiresAt  string           `json:"expires_at,omitempty"`
	Status     ExpirationStatus `json:"status"`
	StatusText string           `json:"status_text"`
	DaysLeft   int              `json:"days_left,omitempty"`
	Source     string           `json:"source,omitempty"`
}

// MaintenanceReport contains the full maintenance status report
type MaintenanceReport struct {
	DatabaseName   string        `json:"database_name"`
	CheckedAt      time.Time     `json:"checked_at"`
	TotalEntries   int           `json:"total_entries"`
	ValidEntries   int           `json:"valid_entries"`
	ExpiringEntries int          `json:"expiring_entries"`
	ExpiredEntries int           `json:"expired_entries"`
	NoExpiryEntries int          `json:"no_expiry_entries"`
	Details        []EntryStatus `json:"details,omitempty"`
}

// CheckExpiration performs a full expiration check on all Rich Context entries
func (m *RichContextMaintainer) CheckExpiration() *MaintenanceReport {
	m.ctx.mu.RLock()
	defer m.ctx.mu.RUnlock()

	report := &MaintenanceReport{
		DatabaseName: m.ctx.DatabaseName,
		CheckedAt:    time.Now(),
		Details:      []EntryStatus{},
	}

	now := time.Now()

	for tableName, table := range m.ctx.Tables {
		for key, value := range table.RichContext {
			report.TotalEntries++

			status := EntryStatus{
				TableName: tableName,
				Key:       key,
				Content:   value.Content,
				ExpiresAt: value.ExpiresAt,
			}

			// Detect source from content marker
			status.Source = string(detectSourceFromContent(value.Content))

			if value.ExpiresAt == "" {
				status.Status = StatusNoExpiry
				status.StatusText = "No expiration set"
				report.NoExpiryEntries++
			} else {
				expiresAt, err := time.Parse(time.RFC3339, value.ExpiresAt)
				if err != nil {
					// Try simpler format
					expiresAt, err = time.Parse("2006-01-02", value.ExpiresAt)
				}

				if err != nil {
					status.Status = StatusNoExpiry
					status.StatusText = "Invalid expiration format"
					report.NoExpiryEntries++
				} else if expiresAt.Before(now) {
					status.Status = StatusExpired
					status.StatusText = "Expired"
					status.DaysLeft = -int(now.Sub(expiresAt).Hours() / 24)
					report.ExpiredEntries++
				} else if expiresAt.Before(now.Add(m.warningThreshold)) {
					status.Status = StatusExpiringSoon
					status.StatusText = "Expiring soon"
					status.DaysLeft = int(expiresAt.Sub(now).Hours() / 24)
					report.ExpiringEntries++
				} else {
					status.Status = StatusValid
					status.StatusText = "Valid"
					status.DaysLeft = int(expiresAt.Sub(now).Hours() / 24)
					report.ValidEntries++
				}
			}

			report.Details = append(report.Details, status)
		}
	}

	return report
}

// GetExpiredEntries returns only the expired entries
func (m *RichContextMaintainer) GetExpiredEntries() []EntryStatus {
	report := m.CheckExpiration()
	var expired []EntryStatus
	for _, entry := range report.Details {
		if entry.Status == StatusExpired {
			expired = append(expired, entry)
		}
	}
	return expired
}

// GetExpiringEntries returns entries that will expire soon
func (m *RichContextMaintainer) GetExpiringEntries() []EntryStatus {
	report := m.CheckExpiration()
	var expiring []EntryStatus
	for _, entry := range report.Details {
		if entry.Status == StatusExpiringSoon {
			expiring = append(expiring, entry)
		}
	}
	return expiring
}

// UpdateEntry updates a Rich Context entry with auto-maintenance metadata
func (m *RichContextMaintainer) UpdateEntry(tableName, key, content string, source CatalogSource, reason string) error {
	m.ctx.mu.Lock()
	defer m.ctx.mu.Unlock()

	table, exists := m.ctx.Tables[tableName]
	if !exists {
		return fmt.Errorf("table %s not found", tableName)
	}

	if table.RichContext == nil {
		table.RichContext = make(map[string]RichContextValue)
	}

	// Calculate expiration based on source
	var expiresAt string
	switch source {
	case SourceCatalog:
		// Catalog entries don't expire (authoritative)
		expiresAt = ""
	case SourceLLM, SourceAnalysis:
		// LLM-inferred entries expire after default period
		expiresAt = time.Now().Add(m.defaultExpiration).Format("2006-01-02")
	case SourceAutoCorrected:
		// Auto-corrected entries have shorter expiration for review
		expiresAt = time.Now().Add(m.defaultExpiration / 2).Format("2006-01-02")
	case SourceUser:
		// User entries don't expire
		expiresAt = ""
	}

	// Build content with source marker
	markedContent := fmt.Sprintf("%s [source: %s]", content, source)
	if reason != "" {
		markedContent = fmt.Sprintf("%s [reason: %s]", markedContent, reason)
	}

	table.RichContext[key] = RichContextValue{
		BusinessNote: BusinessNote{
			Content:   markedContent,
			ExpiresAt: expiresAt,
		},
	}

	return nil
}

// RecordUpdate records an update to the update history
type UpdateHistoryEntry struct {
	Timestamp time.Time     `json:"timestamp"`
	TableName string        `json:"table_name"`
	Key       string        `json:"key"`
	OldValue  string        `json:"old_value,omitempty"`
	NewValue  string        `json:"new_value"`
	Source    CatalogSource `json:"source"`
	Reason    string        `json:"reason,omitempty"`
}

// UpdateHistory tracks changes to Rich Context
type UpdateHistory struct {
	Entries []UpdateHistoryEntry `json:"entries"`
	MaxSize int                  `json:"-"` // Maximum number of entries to keep
}

// NewUpdateHistory creates a new update history tracker
func NewUpdateHistory(maxSize int) *UpdateHistory {
	if maxSize <= 0 {
		maxSize = 100
	}
	return &UpdateHistory{
		Entries: []UpdateHistoryEntry{},
		MaxSize: maxSize,
	}
}

// Add adds a new entry to the history
func (h *UpdateHistory) Add(entry UpdateHistoryEntry) {
	entry.Timestamp = time.Now()
	h.Entries = append([]UpdateHistoryEntry{entry}, h.Entries...)

	// Trim if exceeds max size
	if len(h.Entries) > h.MaxSize {
		h.Entries = h.Entries[:h.MaxSize]
	}
}

// GetRecent returns the most recent n entries
func (h *UpdateHistory) GetRecent(n int) []UpdateHistoryEntry {
	if n > len(h.Entries) {
		n = len(h.Entries)
	}
	return h.Entries[:n]
}

// detectSourceFromContent tries to detect the source from content marker
func detectSourceFromContent(content string) CatalogSource {
	switch {
	case contains(content, "[source: catalog]"):
		return SourceCatalog
	case contains(content, "[source: user]"):
		return SourceUser
	case contains(content, "[source: auto_corrected]"):
		return SourceAutoCorrected
	case contains(content, "[source: analysis]"):
		return SourceAnalysis
	default:
		return SourceLLM
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// FormatMaintenanceReportForPrompt formats the maintenance report for LLM prompt
func FormatMaintenanceReportForPrompt(report *MaintenanceReport) string {
	if report == nil {
		return ""
	}

	result := fmt.Sprintf("## Rich Context Maintenance Status\n\n")
	result += fmt.Sprintf("Database: %s\n", report.DatabaseName)
	result += fmt.Sprintf("Last checked: %s\n\n", report.CheckedAt.Format("2006-01-02 15:04:05"))

	if report.ExpiredEntries > 0 {
		result += fmt.Sprintf("⚠️ **%d EXPIRED entries** require attention:\n", report.ExpiredEntries)
		for _, entry := range report.Details {
			if entry.Status == StatusExpired {
				result += fmt.Sprintf("- %s.%s: %s (expired %d days ago)\n",
					entry.TableName, entry.Key, truncate(entry.Content, 50), -entry.DaysLeft)
			}
		}
		result += "\n"
	}

	if report.ExpiringEntries > 0 {
		result += fmt.Sprintf("⏰ **%d entries** expiring soon:\n", report.ExpiringEntries)
		for _, entry := range report.Details {
			if entry.Status == StatusExpiringSoon {
				result += fmt.Sprintf("- %s.%s: expires in %d days\n",
					entry.TableName, entry.Key, entry.DaysLeft)
			}
		}
		result += "\n"
	}

	return result
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
