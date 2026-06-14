import type { DaemonArch } from './ktx'

export const daemonArch: DaemonArch = {
  id: 'daemon',
  input: {
    label: 'ktx-daemon · Python sidecar',
    note: 'TS CLI 通过 managed-python-daemon / managed-python-http 启动并保活；CPU 密集任务走 ProcessPoolExecutor',
  },
  endpoints: [
    {
      group: 'Semantic Layer',
      icon: 'i-lucide-cpu',
      accent: 'amber',
      routes: [
        { path: 'POST /semantic-layer/query', desc: 'SourceDefinition → SemanticEngine.query → {sql, dialect, columns, plan}' },
        { path: 'POST /semantic-layer/validate', desc: '五项一致性 + duplicate-measure 检查' },
        { path: 'POST /semantic-layer/generate-sources', desc: 'schema scan → SourceDefinition 列表（自动猜列类型/ID/join）' },
      ],
    },
    {
      group: 'Database & SQL',
      icon: 'i-lucide-database',
      accent: 'blue',
      routes: [
        { path: 'POST /database/introspect', desc: 'Postgres READ ONLY：TABLES/COLUMNS/FK SQL' },
        { path: 'POST /sql/analyze-batch', desc: 'sqlglot fingerprint + tables_touched + literal_slots' },
        { path: 'POST /sql/validate-read-only', desc: '黑名单 Insert/Update/Delete/Alter/Create…' },
        { path: 'POST /sql/parse-table-identifier', desc: 'Looker 表标识符解析（${SCHEMA}.${TABLE}）' },
      ],
    },
    {
      group: 'Embeddings & LookML',
      icon: 'i-lucide-spline',
      accent: 'violet',
      routes: [
        { path: 'POST /embeddings/compute', desc: '单条 embedding（all-MiniLM-L6-v2 · 384d）' },
        { path: 'POST /embeddings/compute-bulk', desc: '批量 embedding' },
        { path: 'POST /lookml/parse', desc: 'LookML view/explore/derived_table → ktx 结构' },
      ],
    },
    {
      group: 'Code Execution',
      icon: 'i-lucide-code-2',
      accent: 'rose',
      routes: [
        { path: 'POST /code/execute', desc: '进程内 exec + pandas/numpy/requests（需 --enable-code-execution）' },
      ],
    },
  ],
  pools: [
    { name: 'ProcessPoolExecutor', desc: 'sqlglot analyze_batch / fingerprint 等 CPU 密集任务', icon: 'i-lucide-cpu' },
    { name: 'SentenceTransformers', desc: '懒加载 all-MiniLM-L6-v2，单/批 embedding 接口', icon: 'i-lucide-spline' },
    { name: 'uvicorn', desc: 'serve-http 长驻 · telemetry 钩子（crash → report_exception）', icon: 'i-lucide-server' },
  ],
  ports: [
    { tsPort: 'KtxSemanticLayerComputePort', httpRoute: '/semantic-layer/*', usedBy: 'sl_query · sl_validate · ingest WU' },
    { tsPort: 'SqlAnalysisPort', httpRoute: '/sql/*', usedBy: 'historic-sql ingest · sql_execution MCP' },
    { tsPort: 'KtxEmbeddingPort', httpRoute: '/embeddings/*', usedBy: 'search index · scan enrichment' },
    { tsPort: 'LiveDatabaseIntrospectionPort', httpRoute: '/database/introspect', usedBy: 'live-database adapter · scan' },
  ],
  startup: 'ktx mcp start → managed-python-daemon → uvicorn ktx_daemon.app:create_app\nstdin CLI: semantic-query / semantic-validate / embedding-compute / serve-http',
  insights: {
    input: 'ktx 是 pnpm + uv 双 workspace：TS 管编排/IO/Git，Python 管重计算（SL 规划、SQL 解析、嵌入、LookML）。daemon 是跨语言边界。',
    bridge: [
      { icon: 'i-lucide-bridge', title: 'TS 编排 · Python 计算', body: 'CLI/MCP 在 Node 侧调度一切用户可见流程；遇到 SL query / SQL analyze / embedding 时 HTTP 委派给 daemon——语言各取所长。' },
      { icon: 'i-lucide-shield', title: 'validate-read-only 硬闸', body: '所有经 daemon 执行的 SQL 先过黑名单校验；MCP sql_execution 和 historic-sql ingest 共享同一端口——写操作在 Python 层就被拒绝。' },
    ],
    isolate: [
      { icon: 'i-lucide-box', title: '按需启动、无托管服务', body: '没有云端 daemon；ktx mcp start 时本地拉起 Python 进程，Agent 断开即停——零运维、零外发。' },
    ],
  },
}
