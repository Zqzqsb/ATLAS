package scenarios

import (
	"fmt"
	"strings"

	"lucid/internal/adapter"
	"lucid/internal/react"
	reacttools "lucid/internal/react/tools"

	lctools "github.com/tmc/langchaingo/tools"
)

// InferenceConfig holds parameters for the inference (SQL generation) scenario.
type InferenceConfig struct {
	// Database
	DBType string

	// Schema & Context
	ContextPrompt  string // Pre-built rich schema prompt
	UseRichContext  bool

	// Iteration budget
	MaxIterations int

	// Clarify mode
	ClarifyMode             string   // "off" | "force"
	ResultFields            []string
	ResultFieldsDescription string

	// Callback
	StepCallback react.StepCallback
}

// BuildInferenceEngine creates a ReAct EngineConfig for the SQL generation pipeline.
//
// Tools: execute_sql (with DryRun + observation injection), verify_sql (with observation injection)
// Iteration strategy: actualMax = claimedMax + 3 (small buffer for verify retries)
func BuildInferenceEngine(
	dbAdapter adapter.DBAdapter,
	cfg InferenceConfig,
) (*react.EngineConfig, *reacttools.InferenceSQLTool) {
	if cfg.MaxIterations <= 0 {
		cfg.MaxIterations = 5
	}

	sqlTool := reacttools.NewInferenceSQLTool(dbAdapter, true)
	verifySQLTool := reacttools.NewVerifySQLTool(dbAdapter, cfg.DBType)

	toolsList := []lctools.Tool{sqlTool, verifySQLTool}

	prompt := buildInferencePrompt(cfg)

	return &react.EngineConfig{
		MaxIterations:     cfg.MaxIterations,
		ActualMaxOverride: cfg.MaxIterations + 3,
		SystemPrompt:      prompt,
		Tools:             toolsList,
		StepCallback:      cfg.StepCallback,
		LogMode:           "simple",
	}, sqlTool
}

// SchemaLinkingConfig holds parameters for the ReAct schema linking scenario.
type SchemaLinkingConfig struct {
	DBAdapter    adapter.DBAdapter
	SchemaDesc   string // Pre-formatted schema description
	Query        string
	StepCallback react.StepCallback
}

// BuildSchemaLinkingEngine creates a ReAct EngineConfig for schema linking.
//
// Tool: execute_sql (no DryRun, no observation injection needed)
// Iteration strategy: claimed=5, actual=15 (generous buffer for exploration)
func BuildSchemaLinkingEngine(cfg SchemaLinkingConfig) *react.EngineConfig {
	sqlTool := reacttools.NewExecuteSQL(cfg.DBAdapter)

	prompt := buildSchemaLinkingPrompt(cfg)

	return &react.EngineConfig{
		MaxIterations:     5,
		ActualMaxOverride: 15,
		SystemPrompt:      prompt,
		Tools:             []lctools.Tool{sqlTool},
		StepCallback:      cfg.StepCallback,
		LogMode:           "simple",
	}
}

// --- Prompt builders ---

func buildInferencePrompt(cfg InferenceConfig) string {
	var sb strings.Builder

	sb.WriteString("You are a SQL expert. Generate SQL to answer the question.\n\n")

	// Database type info
	if cfg.DBType != "" {
		sb.WriteString(fmt.Sprintf("**Database Type: %s**\n", cfg.DBType))
		sb.WriteString(fmt.Sprintf("CRITICAL: Write SQL that strictly follows %s syntax rules.\n", cfg.DBType))
		sb.WriteString("Common syntax notes:\n")
		sb.WriteString("- Use backticks for identifiers, single quotes for strings\n")
		sb.WriteString("- LIMIT syntax: LIMIT offset, count\n")
		sb.WriteString("- Use CONCAT() for string concatenation\n")
		sb.WriteString("\n")
	}

	// Rich Context
	if cfg.ContextPrompt != "" {
		sb.WriteString("Database Schema:\n")
		sb.WriteString(cfg.ContextPrompt)
		sb.WriteString("\n\n")
	}

	// SQL Best Practices (only in Rich Context mode)
	if cfg.UseRichContext {
		sb.WriteString(`IMPORTANT: Rich Context may be outdated or incorrect. When Rich Context conflicts with actual database data, trust the database.

SQL Best Practices:
1. TEXT fields storing numbers: Use CAST(field AS INTEGER/REAL) for comparisons and sorting
2. NULL handling:
   - NULL means "unknown/uncertain", not zero.
   - When doing aggregations on numeric data stored in TEXT fields (like 'MPG' or 'Horsepower'), be aware of non-numeric string values like 'null'.
   - Filter both SQL NULLs and string NULLs: WHERE field IS NOT NULL AND field != 'null'
3. String matching:
   - Use exact values from Rich Context when available (e.g., if Rich Context lists "USA, UK, France", use these exact strings)
   - If no exact values in Rich Context and NOT in ReAct mode: use case-insensitive matching (LOWER(field) = LOWER('value'))
   - If no exact values in Rich Context and IN ReAct mode: explore with execute_sql to find exact values first
4. Duplicates: When the question asks for a list of items (e.g., names, cities), duplicates are often undesirable. If your query joins tables in a way that might create duplicates (e.g., one student has multiple pets), consider using DISTINCT to ensure unique results.
5. Zero values:
   - Zero (0) means "business non-existence" (e.g., population=0 means no people)
   - Zero is different from NULL (NULL = unknown, 0 = known to be zero)
   - Check Rich Context for specific meaning of zero in each field
6. Extreme values (MIN/MAX/TOP/LIMIT):
   - When finding extreme values (youngest, oldest, highest, lowest, etc.):
     * ALWAYS return ALL rows with the extreme value (handle ties properly)
     * Use subquery pattern: WHERE column = (SELECT MIN/MAX(column) FROM table)
     * Example: SELECT * FROM table WHERE value = (SELECT MAX(value) FROM table)
   - AVOID: ORDER BY ... LIMIT 1 (only returns one arbitrary row when there are ties)
   - Exception: If the question explicitly asks for "one" or "any one", then LIMIT 1 is acceptable
7. Value Mapping: When the question contains specific text values (e.g., "amc hornet sportabout (sw)"), you MUST verify which column contains this value before using it in a WHERE clause. DO NOT GUESS between similar columns (e.g., 'Make' vs 'Model'). Use 'execute_sql' with a 'WHERE' clause to check for the value's existence.
8. Data format conflicts:
   - If Rich Context says "2-digit year (70=1970)" but query returns 0 results, try 4-digit year (1970)
   - Always verify actual data format with execute_sql when encountering unexpected empty results
9. Data Formatting and Whitespace:
   - Be cautious of hidden characters or formatting that can cause 'WHERE' clause mismatches, especially in 'TEXT' fields.
   - **Leading/Trailing Spaces:** Values might have extra spaces (e.g., '' USA '' instead of ''USA''). Use 'TRIM()' (e.g., 'WHERE TRIM(Country) = ''USA''') to handle this.
   - **Special Characters:** Data might be enclosed in quotes or other characters (e.g., '''"France"''').
   - If a query with a 'WHERE' clause on a 'TEXT' field unexpectedly returns no results, suspect a formatting issue. Use 'execute_sql' with 'LIKE ''%value%''' to investigate the actual data format.

`)
	}

	// Tools
	sb.WriteString(`Available Tools:
- execute_sql: Execute SQL and see results
- verify_sql: Validate SQL via EXPLAIN before giving Final Answer
  → Returns: ✅ VERIFY_PASSED or ❌ VERIFY_FAILED, plus EXPLAIN execution plan and performance warnings

Workflow:
1. Analyze question and schema
2. If string values missing from Rich Context → use execute_sql to find them
3. Write SQL following best practices
4. ALWAYS call verify_sql to validate your SQL and inspect the execution plan
   - If ❌ FAILED: fix the SQL error and call verify_sql AGAIN. Repeat until it passes.
   - NEVER give Final Answer with SQL that has not passed verify_sql.
   - If ✅ PASSED with no warnings: proceed to Final Answer
   - If ✅ PASSED with performance warnings: evaluate the warnings:
     * Full table scan on small tables (≤1000 rows) is acceptable
     * Full table scan on large tables (>1000 rows): try to optimize (add WHERE, use indexed columns)
     * If optimization is not feasible, proceed to Final Answer with the current SQL
5. Provide Final Answer — the SQL in your Final Answer MUST be the one that passed verify_sql

`)

	// Output format
	sb.WriteString(`Output Format (choose ONE):
A) Use tool:
   Thought: [reasoning]
   Action: [tool_name]
   Action Input: [input]

B) Give answer:
   Thought: [reasoning]
   Final Answer: [SQL only, no markdown]

⚠️ NEVER write "Action: None"! If no tool needed, use option B.

`)

	// Critical rules
	sb.WriteString(fmt.Sprintf(`Critical Rules:
1. Field Order: SELECT fields MUST match expected order exactly (no table prefixes)
2. Iterations: %d max. Track: "Iteration X/%d"
3. Efficiency: Only use execute_sql when truly uncertain about data values.
4. ALWAYS verify: Call verify_sql before Final Answer. Review the EXPLAIN plan — optimize if large table scans are avoidable, otherwise accept.
5. Final Answer: SQL only, no explanations

`, cfg.MaxIterations, cfg.MaxIterations))

	// Force mode field requirements
	if cfg.ClarifyMode == "force" && len(cfg.ResultFields) > 0 {
		fieldsStr := strings.Join(cfg.ResultFields, ", ")
		sb.WriteString(fmt.Sprintf("⚠️ REQUIRED OUTPUT FIELDS:\nYour SQL query MUST return EXACTLY these fields in this EXACT ORDER: %s\n", fieldsStr))
		if cfg.ResultFieldsDescription != "" {
			sb.WriteString(fmt.Sprintf("Field descriptions: %s\n", cfg.ResultFieldsDescription))
		}
		sb.WriteString("\nCRITICAL: Use these field names WITHOUT table prefixes (e.g., 'Name' not 'singer.Name').\n")
		sb.WriteString("Any deviation from this field list will be considered INCORRECT.\n\n")

		// Repeat at end for long-range attention
		sb.WriteString(fmt.Sprintf(`
⚠️ REMINDER - REQUIRED OUTPUT FIELDS ⚠️
Before Final Answer, verify your SQL returns these EXACT fields in EXACT order:
Required: %s
`, fieldsStr))
		if cfg.ResultFieldsDescription != "" {
			sb.WriteString(fmt.Sprintf("(%s)\n", cfg.ResultFieldsDescription))
		}
		sb.WriteString("If field is a name/description, JOIN the referenced table. Do NOT return IDs when names are required.\n")
	}

	return sb.String()
}

func buildSchemaLinkingPrompt(cfg SchemaLinkingConfig) string {
	return fmt.Sprintf(`You are a database expert. Identify which tables are relevant to answer the question.

⚠️  ITERATION LIMIT: You have maximum 5 iterations to complete this task. Be efficient!

Available Tables:
%s

Question: %s

Foreign key relationships are shown above. Use them to:
1. Identify direct relationships between tables
2. Find intermediate junction tables for many-to-many relationships
3. Trace the join path from source to target tables

You can use execute_sql to:
- Verify data existence: SELECT COUNT(*) FROM table
- Check join validity: SELECT COUNT(*) FROM t1 JOIN t2 ON ...
- Explore sample data: SELECT * FROM table LIMIT 3
- Check column values: SELECT DISTINCT column FROM table LIMIT 5

Workflow:
1. Identify tables with columns that seem relevant to the question.
2. Use the foreign key relationships to find all necessary tables for joins.
3. If you are unsure about a table's relevance, use 'execute_sql' to sample its data.
4. Provide the final list of tables.

Output Format:
A) Use tool to explore:
   Thought: [reasoning]
   Action: execute_sql
   Action Input: [SQL query]

B) Give final answer:
   Thought: [reasoning]
   Final Answer: table1, table2, table3

IMPORTANT:
- Output comma-separated table names only in Final Answer
- Include ALL tables needed for joins (don't miss intermediate tables)
- For NOT queries, include base table
- For foreign key columns, include referenced tables
- If all tables needed, output: all
- If no tables needed, output: none

Output:`, cfg.SchemaDesc, cfg.Query)
}
