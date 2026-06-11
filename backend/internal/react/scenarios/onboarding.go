package scenarios

import (
	"fmt"
	"sort"
	"strings"

	"atlas/internal/adapter"
	"atlas/internal/lakebase"
	"atlas/internal/react"
	reacttools "atlas/internal/react/tools"

	lctools "github.com/tmc/langchaingo/tools"
)

// OnboardingConfig holds parameters for the onboarding scenario.
type OnboardingConfig struct {
	DatasourceID  int64
	DBType        string
	Tables        []*lakebase.TableInfo
	Columns       []*lakebase.ColumnInfo
	Relations     []*lakebase.Relation
	MaxIterations int
	MinIterations int
	StepCallback  react.StepCallback
}

// TableCluster represents a connected component (subtree) in the FK graph.
// Each cluster is a group of tables connected by foreign-key relations.
type TableCluster struct {
	ID        int // Cluster index (0-based)
	Tables    []*lakebase.TableInfo
	Columns   []*lakebase.ColumnInfo
	Relations []*lakebase.Relation
}

// ForestDecomposeResult holds the output of forest decomposition.
type ForestDecomposeResult struct {
	Clusters     []*TableCluster
	TotalTables  int
	LargestSize  int
	MedianSize   int
	IsolatedCount int // Tables with no FK relations
}

// ForestDecompose builds an undirected FK graph from relations and decomposes
// it into connected components (a forest of table clusters).
//
// Algorithm:
//  1. Build adjacency list from FK relations (both directions).
//  2. Run BFS/DFS to find connected components.
//  3. Sort components by size (largest first).
//  4. Assign tables, columns, and intra-cluster relations to each cluster.
func ForestDecompose(
	tables []*lakebase.TableInfo,
	columns []*lakebase.ColumnInfo,
	relations []*lakebase.Relation,
) *ForestDecomposeResult {
	// Build table name → index mapping
	tableIdx := make(map[string]int, len(tables))
	for i, t := range tables {
		tableIdx[t.TableName] = i
	}

	// Build undirected adjacency list from FK relations
	adj := make(map[string]map[string]bool, len(tables))
	for _, t := range tables {
		adj[t.TableName] = make(map[string]bool)
	}
	for _, r := range relations {
		if _, ok := adj[r.FromTable]; ok {
			if _, ok2 := adj[r.ToTable]; ok2 {
				adj[r.FromTable][r.ToTable] = true
				adj[r.ToTable][r.FromTable] = true
			}
		}
	}

	// BFS to find connected components
	visited := make(map[string]bool, len(tables))
	var components [][]string

	for _, t := range tables {
		if visited[t.TableName] {
			continue
		}
		// BFS from this table
		component := []string{}
		queue := []string{t.TableName}
		visited[t.TableName] = true

		for len(queue) > 0 {
			cur := queue[0]
			queue = queue[1:]
			component = append(component, cur)

			for neighbor := range adj[cur] {
				if !visited[neighbor] {
					visited[neighbor] = true
					queue = append(queue, neighbor)
				}
			}
		}
		components = append(components, component)
	}

	// Sort components by size descending
	sort.Slice(components, func(i, j int) bool {
		return len(components[i]) > len(components[j])
	})

	// Build column lookup by table name
	colByTable := make(map[string][]*lakebase.ColumnInfo, len(tables))
	for _, c := range columns {
		colByTable[c.TableName] = append(colByTable[c.TableName], c)
	}

	// Build clusters
	clusters := make([]*TableCluster, 0, len(components))
	isolatedCount := 0

	for i, comp := range components {
		compSet := make(map[string]bool, len(comp))
		for _, tn := range comp {
			compSet[tn] = true
		}

		// Collect tables
		clusterTables := make([]*lakebase.TableInfo, 0, len(comp))
		for _, tn := range comp {
			if idx, ok := tableIdx[tn]; ok {
				clusterTables = append(clusterTables, tables[idx])
			}
		}

		// Collect columns for these tables
		clusterColumns := make([]*lakebase.ColumnInfo, 0)
		for _, tn := range comp {
			clusterColumns = append(clusterColumns, colByTable[tn]...)
		}

		// Collect intra-cluster relations
		clusterRelations := make([]*lakebase.Relation, 0)
		for _, r := range relations {
			if compSet[r.FromTable] && compSet[r.ToTable] {
				clusterRelations = append(clusterRelations, r)
			}
		}

		if len(comp) == 1 && len(clusterRelations) == 0 {
			isolatedCount++
		}

		clusters = append(clusters, &TableCluster{
			ID:        i,
			Tables:    clusterTables,
			Columns:   clusterColumns,
			Relations: clusterRelations,
		})
	}

	// Calculate median size
	sizes := make([]int, len(clusters))
	largest := 0
	for i, c := range clusters {
		sizes[i] = len(c.Tables)
		if sizes[i] > largest {
			largest = sizes[i]
		}
	}
	sort.Ints(sizes)
	median := 0
	if len(sizes) > 0 {
		median = sizes[len(sizes)/2]
	}

	return &ForestDecomposeResult{
		Clusters:      clusters,
		TotalTables:   len(tables),
		LargestSize:   largest,
		MedianSize:    median,
		IsolatedCount: isolatedCount,
	}
}

// MergeIsolatedTables groups isolated tables (single-table clusters with no FK)
// into batches of the given size for more efficient onboarding.
func MergeIsolatedTables(clusters []*TableCluster, batchSize int) []*TableCluster {
	if batchSize <= 0 {
		batchSize = 15
	}

	var connected []*TableCluster
	var isolated []*TableCluster

	for _, c := range clusters {
		if len(c.Tables) == 1 && len(c.Relations) == 0 {
			isolated = append(isolated, c)
		} else {
			connected = append(connected, c)
		}
	}

	// Merge isolated tables into batches
	for i := 0; i < len(isolated); i += batchSize {
		end := i + batchSize
		if end > len(isolated) {
			end = len(isolated)
		}
		batch := &TableCluster{
			ID: len(connected),
		}
		for _, iso := range isolated[i:end] {
			batch.Tables = append(batch.Tables, iso.Tables...)
			batch.Columns = append(batch.Columns, iso.Columns...)
		}
		connected = append(connected, batch)
	}

	// Re-number cluster IDs
	for i, c := range connected {
		c.ID = i
	}

	return connected
}

// ComputeChunkBudget calculates min/max iteration budgets for a table cluster.
// Budget scales with table count: ~7 iterations per table for large clusters
// (explore schema + sample data + set table desc + set column descs).
func ComputeChunkBudget(tableCount int) (minIter, maxIter int) {
	target := tableCount*3 + 10
	maxIter = max(15, int(float64(target)*1.5))

	// Dynamic per-chunk cap: scales with table count
	// ≤10 tables → cap 60, ≤25 tables → cap 150, ≤50 tables → cap 300, >50 → cap 500
	cap := 60
	switch {
	case tableCount > 50:
		cap = 500
	case tableCount > 25:
		cap = 300
	case tableCount > 10:
		cap = 150
	}
	if maxIter > cap {
		maxIter = cap
	}

	minIter = max(3, int(float64(target)*0.6))
	// Ensure minIter <= maxIter
	if minIter > maxIter {
		minIter = maxIter
	}
	return minIter, maxIter
}

// BuildOnboardingEngine creates a ReAct EngineConfig for database onboarding.
//
// The onboarding agent explores the database schema, discovers data patterns,
// and writes Rich Context (table descriptions, column descriptions, sample values, etc.)
// using the same tool set as RC generation but with a prompt tuned for first-time exploration.
func BuildOnboardingEngine(
	businessDB adapter.DBAdapter,
	rcWriter reacttools.RCWriter,
	cfg OnboardingConfig,
) *react.EngineConfig {
	if cfg.MaxIterations <= 0 {
		cfg.MaxIterations = 20
	}
	if cfg.MinIterations <= 0 {
		cfg.MinIterations = 5
	}

	sqlTool := reacttools.NewExecuteSQL(businessDB)
	rcTool := reacttools.NewSetRichContext(rcWriter, cfg.DatasourceID)

	toolsList := []lctools.Tool{sqlTool, rcTool}

	prompt := buildOnboardingPrompt(cfg)

	return &react.EngineConfig{
		MaxIterations: cfg.MaxIterations,
		MinIterations: cfg.MinIterations,
		SystemPrompt:  prompt,
		Tools:         toolsList,
		StepCallback:  cfg.StepCallback,
		LogMode:       "simple",
		Verbose:       true,
	}
}

func buildOnboardingPrompt(cfg OnboardingConfig) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf(`You are onboarding a new %s database into the ATLAS Text-to-SQL system.
Your mission is to thoroughly explore the database and generate Rich Context metadata that helps future SQL generation.

## Available Tools
- execute_sql: Run SELECT/SHOW/DESCRIBE queries to explore the database
- set_rich_context: Save discovered context (table descriptions, column descriptions, sample values, synonyms, business terms)

## Onboarding Workflow
1. **Understand the schema**: For each table, examine its columns, types, and constraints
2. **Sample the data**: Run SELECT * FROM table LIMIT 5 to understand actual data
3. **Discover distributions**: For text/enum columns, run GROUP BY to find value distributions
4. **Check data quality**: Look for NULLs, whitespace issues, unexpected patterns
5. **Identify relationships**: Understand foreign keys and cross-table patterns
6. **Write context**: Use set_rich_context to save ALL discoveries

## Rich Context Types
Use set_rich_context with JSON input:

- Table description: {"type":"table_description","table":"...","value":"2-3 sentence description of business purpose"}
- Column description: {"type":"column_description","table":"...","column":"...","value":"semantic meaning"}
- Sample values: {"type":"column_sample_values","table":"...","column":"...","value":"val1, val2, val3"}
- Synonyms: {"type":"column_synonyms","table":"...","column":"...","value":"synonym1, synonym2"}
- Business term: {"type":"business_term","value":"term","definition":"...","category":"..."}

## Important Rules
- ALWAYS explore with execute_sql BEFORE writing context — never guess
- Write context incrementally as you discover it
- Be thorough: cover ALL tables, not just a few
- For enum-like columns (< 20 distinct values), always record sample values
- For text columns, check for leading/trailing whitespace issues
- Primary keys only need brief descriptions ("Primary key, auto-increment")
- Focus effort on semantically interesting columns (status fields, categories, codes)
`, cfg.DBType))

	// Add schema overview
	sb.WriteString("\n## Database Schema\n")
	sb.WriteString(fmt.Sprintf("Total tables: %d\n\n", len(cfg.Tables)))

	colByTable := make(map[string][]*lakebase.ColumnInfo)
	for _, c := range cfg.Columns {
		colByTable[c.TableName] = append(colByTable[c.TableName], c)
	}

	for _, t := range cfg.Tables {
		sb.WriteString(fmt.Sprintf("### %s", t.TableName))
		if t.RowCount > 0 {
			sb.WriteString(fmt.Sprintf(" (%d rows)", t.RowCount))
		}
		sb.WriteString("\n")

		cols := colByTable[t.TableName]
		for _, c := range cols {
			flags := ""
			if c.IsPrimaryKey {
				flags += " [PK]"
			}
			if c.IsForeignKey {
				flags += " [FK]"
			}
			if c.IsNullable {
				flags += " [nullable]"
			}
			sb.WriteString(fmt.Sprintf("  - %s %s%s\n", c.ColumnName, c.DataType.String, flags))
		}
		sb.WriteString("\n")
	}

	// Relations
	if len(cfg.Relations) > 0 {
		sb.WriteString("### Foreign Key Relations\n")
		for _, r := range cfg.Relations {
			sb.WriteString(fmt.Sprintf("- %s.%s → %s.%s\n", r.FromTable, r.FromColumn, r.ToTable, r.ToColumn))
		}
		sb.WriteString("\n")
	}

	sb.WriteString(fmt.Sprintf(`## Iteration Budget
- Minimum: %d iterations (ensure thorough exploration)
- Maximum: %d iterations
- Budget is generous — explore ALL tables carefully.

Begin onboarding now. Start with execute_sql to understand the data.
`, cfg.MinIterations, cfg.MaxIterations))

	return sb.String()
}
