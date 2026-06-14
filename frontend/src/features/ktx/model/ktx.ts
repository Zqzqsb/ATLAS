// ktx architecture model — single source of truth for the /ktx panorama deck.
// Models the post-2025 layout (pnpm + uv workspace) of Kaelio/ktx.
import { ACCENTS, type Accent, type AccentKey, type ArchLayer, type ArchNode } from '../../arch/model/architecture'
import type { Insight, NamedItem } from '../../arch/model/modules'
import { slArch } from './sl-arch'
import { memoryArch } from './memory-arch'
import { daemonArch } from './daemon-arch'
import { storageArch } from './storage-arch'

export { ACCENTS }
export type { Accent, AccentKey, ArchLayer, ArchNode, Insight, NamedItem }

/* ─── L0 panorama: ingest → store → search/serve stack ─── */
export const KTX_LAYERS: ArchLayer[] = [
  {
    id: 'agent',
    title: 'Agent Clients',
    subtitle: 'BYO 外部 Agent · 通过 MCP 或 CLI 接入（ktx 不内置 LLM 服务）',
    icon: 'i-lucide-bot',
    accent: 'slate',
    cols: 3,
    nodes: [
      {
        id: 'mcp-client',
        label: 'MCP Clients',
        sublabel: 'Claude Code · Codex · Cursor · OpenCode（function-calling）',
        icon: 'i-lucide-plug-zap',
        accent: 'slate',
        flow: 'mcp',
        span: 1,
        codeRefs: ['packages/cli/src/mcp-stdio-server.ts', 'packages/cli/src/managed-mcp-daemon.ts'],
      },
      {
        id: 'cli-cmds',
        label: 'ktx CLI',
        sublabel: 'setup · ingest · status · sl · wiki · mcp · sql · doctor',
        icon: 'i-lucide-terminal',
        accent: 'slate',
        span: 1,
        codeRefs: ['packages/cli/src/cli-program.ts', 'packages/cli/src/commands/'],
      },
      {
        id: 'skills-md',
        label: 'Skill: ktx',
        sublabel: 'skills/ktx/SKILL.md · 教 Agent 安装 + 使用 ktx',
        icon: 'i-lucide-scroll-text',
        accent: 'slate',
        span: 1,
        codeRefs: ['skills/ktx/'],
      },
    ],
  },
  {
    id: 'pipelines',
    title: 'Pipelines',
    subtitle: '三大核心流程 · ingest 构建上下文，serve 执行查询，memory 在对话尾部自学',
    icon: 'i-lucide-workflow',
    accent: 'violet',
    cols: 3,
    nodes: [
      {
        id: 'ingest',
        label: 'Ingest Pipeline',
        sublabel: 'fetch → chunk → WU agent → reconcile → finalize → commit',
        icon: 'i-lucide-import',
        accent: 'emerald',
        flow: 'ingest',
        span: 1,
        codeRefs: ['packages/cli/src/context/ingest/ingest-bundle.runner.ts'],
      },
      {
        id: 'search',
        label: 'Search & Serve',
        sublabel: 'Wiki / SL 混合检索（FTS5 + token + 嵌入 → RRF 融合）',
        icon: 'i-lucide-search',
        accent: 'blue',
        flow: 'search',
        span: 1,
        codeRefs: ['packages/cli/src/context/search/hybrid-search-core.ts'],
      },
      {
        id: 'memory-agent',
        label: 'Memory Agent',
        sublabel: '对话尾启动 · 隔离 worktree · LLM 写回 wiki/SL · squash 合并',
        icon: 'i-lucide-brain-circuit',
        accent: 'amber',
        flow: 'memory',
        span: 1,
        codeRefs: ['packages/cli/src/context/memory/memory-agent.service.ts'],
      },
    ],
  },
  {
    id: 'mcp',
    title: 'MCP / Tool Surface',
    subtitle: '把上下文以 MCP 工具暴露给 Agent · 8 个核心 tool · stdio / HTTP',
    icon: 'i-lucide-plug',
    accent: 'violet',
    cols: 2,
    nodes: [
      {
        id: 'mcp-server',
        label: 'MCP Server',
        sublabel: 'stdio + HTTP 双形态 · context_tools 注册中心',
        icon: 'i-lucide-server',
        accent: 'violet',
        flow: 'mcp',
        span: 1,
        codeRefs: ['packages/cli/src/context/mcp/server.ts', 'packages/cli/src/context/mcp/context-tools.ts'],
      },
      {
        id: 'sl-engine',
        label: 'Semantic Layer',
        sublabel: 'sl_query / sl_read_source · YAML → planner → 方言 SQL',
        icon: 'i-lucide-cpu',
        accent: 'amber',
        flow: 'sl',
        span: 1,
        codeRefs: ['python/ktx-sl/src/ktx_sl/planner.py', 'python/ktx-sl/src/ktx_sl/generator.py'],
      },
    ],
  },
  {
    id: 'storage',
    title: 'Storage · git-backed project',
    subtitle: 'Git 化语义层 + wiki + 索引 · ktx.yaml 声明式配置',
    icon: 'i-lucide-folder-git-2',
    accent: 'indigo',
    cols: 4,
    nodes: [
      {
        id: 'wiki-store',
        label: 'wiki/',
        sublabel: 'global / user · Markdown 业务知识 + 矛盾标注',
        icon: 'i-lucide-book-marked',
        accent: 'indigo',
        flow: 'storage',
        span: 1,
      },
      {
        id: 'sl-store',
        label: 'semantic-layer/',
        sublabel: 'YAML sources · _schema 分片 · 关系 / measure',
        icon: 'i-lucide-box',
        accent: 'indigo',
        flow: 'storage',
        span: 1,
      },
      {
        id: 'raw-sources',
        label: 'raw-sources/',
        sublabel: 'scan-report / per-table / 关系画像 / ingest 痕迹',
        icon: 'i-lucide-archive',
        accent: 'indigo',
        flow: 'storage',
        span: 1,
      },
      {
        id: 'sqlite-state',
        label: '.ktx/db.sqlite',
        sublabel: 'FTS5 · embeddings cache · ingest runs · memory runs',
        icon: 'i-lucide-database',
        accent: 'indigo',
        flow: 'storage',
        span: 1,
      },
    ],
  },
  {
    id: 'foundation',
    title: 'Foundation · Daemon + Connectors + LLM',
    subtitle: 'Python 计算 sidecar + 7 个原生数据源连接器 + 多家 LLM/Embedding',
    icon: 'i-lucide-layers',
    accent: 'slate',
    cols: 3,
    nodes: [
      {
        id: 'py-daemon',
        label: 'ktx-daemon',
        sublabel: 'FastAPI · SL 引擎 / introspect / SQL 解析 / 嵌入 / LookML',
        icon: 'i-lucide-cog',
        accent: 'slate',
        flow: 'daemon',
        span: 1,
        codeRefs: ['python/ktx-daemon/src/ktx_daemon/app.py'],
      },
      {
        id: 'connectors',
        label: 'Scan Connectors',
        sublabel: 'postgres · bigquery · snowflake · mysql · clickhouse · sqlserver · sqlite',
        icon: 'i-lucide-plug-2',
        accent: 'slate',
        span: 1,
        codeRefs: ['packages/cli/src/connectors/'],
      },
      {
        id: 'llm-prov',
        label: 'LLM / Embedding',
        sublabel: 'Anthropic API · Vertex AI · AI Gateway · Claude Code SDK · Codex SDK · MiniLM',
        icon: 'i-lucide-brain',
        accent: 'slate',
        span: 1,
        codeRefs: ['packages/cli/src/context/llm/'],
      },
    ],
  },
]
export interface KtxFlowDef {
  id: string
  label: string
  title: string
  subtitle: string
  icon: string
  accent: AccentKey
}
export const ktxFlows: KtxFlowDef[] = [
  {
    id: 'ingest',
    label: 'Ingest Pipeline',
    title: 'Ingest · 6-Stage Bundle Runner',
    subtitle: 'fetch → diff → chunk → WU agent loop → reconciliation → finalize → commit。每个 SourceAdapter（dbt / lookml / looker / metabase / notion / historic-sql / live-database）实现统一协议，runner 串行驱动。',
    icon: 'i-lucide-import',
    accent: 'emerald',
  },
  {
    id: 'search',
    label: 'Search & Discover',
    title: 'Search · Hybrid FTS5 + Embeddings + RRF',
    subtitle: 'wiki / SL 同台检索：lexical（FTS5）· token · semantic（cosine） 三路打分，由 RRF 融合；query → discover_data / dictionary_search / wiki_search 暴露给 Agent。',
    icon: 'i-lucide-search',
    accent: 'blue',
  },
  {
    id: 'mcp',
    label: 'MCP Surface',
    title: 'MCP · 把上下文层暴露成工具',
    subtitle: '一个 stdio / HTTP 服务，注册 connection_list / wiki_search / wiki_read / entity_details / sl_read_source / sl_query / sql_execution / discover_data / memory_ingest 等工具，由 ktx mcp start 拉起。',
    icon: 'i-lucide-plug',
    accent: 'violet',
  },
  {
    id: 'sl',
    label: 'Semantic Layer',
    title: 'Semantic Layer · 13-step 规划与 SQL 生成',
    subtitle: 'YAML sources（models / relationships / measures / segments）→ JoinGraph Dijkstra → planner（dimensions / measures / fan-out / chasm 检测 / locality）→ SqlGenerator → sqlglot transpile（postgres scaffold → 任意方言）。',
    icon: 'i-lucide-cpu',
    accent: 'amber',
  },
  {
    id: 'memory',
    label: 'Memory Agent',
    title: 'Memory Agent · 对话尾自学闭环',
    subtitle: '在用户对话结束后，按需启动一个独立 LLM agent：在隔离 worktree 中写 wiki / SL，跑 SL 验证（warehouse probe），再 squash 合并回主分支——并发安全、可审计。',
    icon: 'i-lucide-brain-circuit',
    accent: 'amber',
  },
  {
    id: 'daemon',
    label: 'Python Daemon',
    title: 'ktx-daemon · 跨语言计算 sidecar',
    subtitle: 'FastAPI 长驻进程：semantic-layer/{query, validate, generate-sources}、database/introspect、embeddings/compute*、sql/{parse-table-identifier, validate-read-only, analyze-batch}、lookml/parse、code/execute。CPU 任务跑 ProcessPool。',
    icon: 'i-lucide-cog',
    accent: 'rose',
  },
  {
    id: 'storage',
    label: 'Storage Layout',
    title: 'Storage · git-backed project + .ktx/ 状态',
    subtitle: 'ktx.yaml 声明式配置 + wiki/、semantic-layer/、raw-sources/ 全 Git 化；.ktx/db.sqlite 索引（FTS5 + embeddings cache + ingest/memory runs）+ .ktx/worktrees/ 隔离分支；凭据存仓外。',
    icon: 'i-lucide-folder-git-2',
    accent: 'indigo',
  },
]
export function getKtxFlow(id: string | null): KtxFlowDef | null {
  if (!id) return null
  return ktxFlows.find((f) => f.id === id) ?? null
}
/* ─── L1 module internal data ─── */

export interface IngestArch {
  id: string
  input: { label: string; note: string }
  adapters: { name: string; icon: string; accent: AccentKey; desc: string }[]
  stages: { name: string; role: string; icon: string; accent: AccentKey; bullets: string[]; tools?: NamedItem[] }[]
  workUnitTools: NamedItem[]
  artifacts: { path: string; desc: string; icon: string }[]
  ports: NamedItem[]
  insights: {
    input: string
    pipeline: Insight[]
    finalize: Insight[]
  }
}

export interface SearchArch {
  id: string
  input: { label: string; note: string; example: string }
  scorers: { name: string; icon: string; accent: AccentKey; desc: string; weight: string }[]
  fusion: { name: string; algo: string; note: string; example: string }
  surfaces: NamedItem[]
  index: { table: string; cols: string[]; note: string }[]
  embeddingProvider: { name: string; dim: string; note: string }
  insights: {
    input: string
    score: Insight[]
    fuse: Insight[]
  }
}

export interface McpArch {
  id: string
  input: { label: string; note: string; clients: string[] }
  transports: { name: string; cmd: string; note: string; icon: string }[]
  tools: { name: string; desc: string; accent: AccentKey; reads?: string; writes?: string }[]
  ports: NamedItem[]
  sampleCall: { tool: string; req: string; resp: string }
  insights: {
    input: string
    surface: Insight[]
    safety: Insight[]
  }
}

export interface SlArch {
  id: string
  input: { label: string; note: string; example: string }
  loader: { title: string; items: NamedItem[] }
  joinGraph: { title: string; algo: string; note: string }
  plannerSteps: { name: string; desc: string }[]
  generator: { paths: NamedItem[]; transpile: string; note: string }
  validate: NamedItem[]
  sqlBefore: string
  sqlAfter: string
  insights: {
    input: string
    plan: Insight[]
    fanout: Insight[]
  }
}

export interface MemoryArch {
  id: string
  input: { label: string; note: string; sources: string[] }
  signals: NamedItem[]
  worktreeFlow: { name: string; desc: string; icon: string }[]
  toolBelt: NamedItem[]
  validation: { title: string; items: NamedItem[]; note: string }
  resultRecord: { field: string; desc: string }[]
  insights: {
    input: string
    isolation: Insight[]
    learn: Insight[]
  }
}

export interface DaemonArch {
  id: string
  input: { label: string; note: string }
  endpoints: { group: string; icon: string; accent: AccentKey; routes: { path: string; desc: string }[] }[]
  pools: { name: string; desc: string; icon: string }[]
  ports: { tsPort: string; httpRoute: string; usedBy: string }[]
  startup: string
  insights: {
    input: string
    bridge: Insight[]
    isolate: Insight[]
  }
}

export interface StorageArch {
  id: string
  input: { label: string; note: string }
  paths: { path: string; role: string; commit: '✓' | '✗'; icon: string; accent: AccentKey; desc: string }[]
  ktxYaml: string
  sqliteTables: NamedItem[]
  resolution: { name: string; desc: string }[]
  insights: {
    input: string
    git: Insight[]
    state: Insight[]
  }
}

export interface KtxModuleData {
  id: string
  accent: AccentKey
  ingest?: IngestArch
  search?: SearchArch
  mcp?: McpArch
  sl?: SlArch
  memory?: MemoryArch
  daemon?: DaemonArch
  storage?: StorageArch
}

const ingestArch: IngestArch = {
  id: 'ingest',
  input: {
    label: 'ktx ingest <connection>',
    note: '从 ktx.yaml 解析 source（dbt / lookml / looker / metabase / notion / historic-sql / live-database）→ resolveConnectionSelection → runKtxPublicIngest',
  },
  adapters: [
    { name: 'live-database', icon: 'i-lucide-database-zap', accent: 'emerald', desc: 'connector.introspect()，按表切 WU，写 scan-report.json' },
    { name: 'dbt', icon: 'i-lucide-package', accent: 'amber', desc: 'manifest.json/catalog.json → 同步 source/test 到 SL' },
    { name: 'lookml', icon: 'i-lucide-folder-tree', accent: 'violet', desc: 'daemon /lookml/parse 把 view/explore/derived_table 转成可消费结构' },
    { name: 'looker', icon: 'i-lucide-bar-chart-3', accent: 'blue', desc: 'API 拉 explore/model → 决定 target connection（mapping）' },
    { name: 'metabase', icon: 'i-lucide-line-chart', accent: 'indigo', desc: 'fanout-planner 把一个 connection 拆成多 target 仓库' },
    { name: 'notion', icon: 'i-lucide-file-text', accent: 'slate', desc: 'page-triage 先 LLM 决定 skip/light/full，按 cursor 增量' },
    { name: 'historic-sql', icon: 'i-lucide-history', accent: 'rose', desc: 'pgss / Snowflake / BQ JOBS_BY_PROJECT → fingerprint → wiki' },
    { name: 'metricflow', icon: 'i-lucide-ruler', accent: 'amber', desc: 'dbt-MetricFlow 的 metric YAML' },
  ],
  stages: [
    {
      name: 'Stage 0 · Prepare + Diff',
      role: 'fetch + stage + 比较前后版本',
      icon: 'i-lucide-git-compare',
      accent: 'slate',
      bullets: [
        'adapter.fetch(pullConfig, stagedDir, ctx) — 远端 → 本地暂存',
        'stageRawFilesStage1 — 拷贝 + SHA-256 per file',
        'diffSetService.compute(prev, curr) → DiffSet { added, modified, deleted, unchanged }',
      ],
    },
    {
      name: 'Stage 2 · Chunk + Cluster',
      role: '切成 WorkUnit',
      icon: 'i-lucide-scissors',
      accent: 'blue',
      bullets: [
        'adapter.chunk(stagedDir, diffSet) → ChunkResult { workUnits, contextReport, parseArtifacts }',
        'adapter.clusterWorkUnits? — 嵌入聚类合并过小 WU',
        '构建 StageIndex（WU + resolutions + 状态）',
      ],
    },
    {
      name: 'Stage 3 · Work-Unit Agent Loop',
      role: 'LLM 多步循环 · 并行受限',
      icon: 'i-lucide-bot',
      accent: 'violet',
      bullets: [
        'pLimit 并发；每个 WU 跑独立 ToolSession',
        'agentRunner.runLoop({ toolSet, stepBudget }) — Reason → Act → Observe',
        '失败时 reset 到 WU pre-state（cycle-safe）',
        '产出 WorkUnitOutcome { status, preSha, postSha, actions[], touchedSlSources[] }',
      ],
    },
    {
      name: 'Stage 4 · Reconciliation',
      role: '跨 WU 冲突合并',
      icon: 'i-lucide-merge',
      accent: 'amber',
      bullets: [
        '只有任一 WU 写过才触发',
        '另一个 agent loop，使用 eviction-list / stage-diff / stage-list / conflict-resolution / emit-* 工具',
        '输出 ReconciliationOutcome { skipped, stopReason, metrics }',
      ],
    },
    {
      name: 'Stage 5-6 · Validate + Commit',
      role: '工件最终闸门 + git 提交',
      icon: 'i-lucide-shield-check',
      accent: 'emerald',
      bullets: [
        'validateFinalIngestArtifacts — wiki refs + SL source 一致性',
        'validateProvenanceRawPaths — raw_path → artifact 审计',
        'repairWikiSlRefs / final-gate-repair — 失败可自修复',
        'commit 到主分支；写 ingest run 到 .ktx/db.sqlite',
      ],
    },
  ],
  workUnitTools: [
    { name: 'context-evidence-search / read / neighbors', desc: 'WU 内可选证据检索（RAG over staged）' },
    { name: 'context-candidate-write / mark', desc: '提名候选事实（promote / merge / reject / conflict）' },
    { name: 'sql-edit-replacer', desc: 'SL 字段 / SQL 表达式安全替换' },
    { name: 'wiki / sl emit-tools', desc: '在 worktree 内写 wiki 页 / SL source' },
  ],
  artifacts: [
    { path: 'wiki/global/*.md · wiki/user/<id>/*.md', desc: 'Markdown 业务知识页', icon: 'i-lucide-book-marked' },
    { path: 'semantic-layer/<conn>/*.yaml + _schema/', desc: 'SL source + manifest 分片', icon: 'i-lucide-box' },
    { path: 'raw-sources/<conn>/<source>/<sync>/', desc: '原始抓取 + 哈希 + report', icon: 'i-lucide-archive' },
    { path: '.ktx/db.sqlite', desc: 'IngestRunStore + FTS5 索引同步', icon: 'i-lucide-database' },
  ],
  ports: [
    { name: 'SourceAdapter', desc: 'detect / fetch / chunk / clusterWorkUnits? / project? / finalize? / describeScope? / onPullSucceeded?' },
    { name: 'IngestProvenancePort', desc: 'raw_path → artifact 审计跟踪' },
    { name: 'LockingService', desc: 'config:repo 单写锁，配合 worktree 隔离' },
    { name: 'MemoryFlowEventSink', desc: '中间事件流推 UI / progress port' },
  ],
  insights: {
    input: '一条 `ktx ingest`，但背后是按 SourceAdapter 协议解耦的 8 种数据源——同一 6-stage 管线驱动所有适配器，不写专属流程。',
    pipeline: [
      { icon: 'i-lucide-network', title: '协议化 SourceAdapter', body: '每种数据源只实现 detect/fetch/chunk 等少量钩子；runner 提供 stage、并发、限流、worktree、provenance、SQLite 状态共用基础设施——加新源不动管线。' },
      { icon: 'i-lucide-bot', title: 'WU 是 LLM agent loop', body: 'Stage 3 不是确定式 ETL，而是给每个 WorkUnit 一个工具腰带 + step budget 的 ReAct 循环：agent 自行决定何时搜证据、何时写 wiki / SL，失败回滚到 preSha 干净重来。' },
      { icon: 'i-lucide-git-fork', title: 'worktree + squash 隔离', body: 'WU / Reconciliation 都跑在独立 git worktree 上，最后 squash 合并；并发 ingest 与对话期 memory agent 互不污染主分支。' },
    ],
    finalize: [
      { icon: 'i-lucide-shield-check', title: '终态闸门', body: 'validateFinalIngestArtifacts + validateProvenanceRawPaths + wiki↔SL ref 修复，把"上下文工件可信"变成 commit 级别的硬约束——不通过不进主分支。' },
    ],
  },
}

const searchArch: SearchArch = {
  id: 'search',
  input: {
    label: 'Agent / CLI 自然语言查询',
    note: '统一搜 wiki + SL · 服务于 wiki_search / discover_data / dictionary_search 工具',
    example: 'ktx wiki "refund policy" · ktx sl "revenue"',
  },
  scorers: [
    { name: 'Lexical (FTS5)', icon: 'i-lucide-type', accent: 'amber', desc: 'SQLite FTS5 BM25 · 倒排索引精确词命中', weight: 'rank=1/k+i' },
    { name: 'Token Overlap', icon: 'i-lucide-tally-5', accent: 'blue', desc: '归一化 token 集合命中数 · 处理短查询稳健性', weight: 'rank=1/k+i' },
    { name: 'Semantic (Cosine)', icon: 'i-lucide-spline', accent: 'violet', desc: 'embedding cosine 距离 · 语义近似', weight: 'rank=1/k+i' },
  ],
  fusion: {
    name: 'RRF Fuser',
    algo: 'score = Σ_scorer 1 / (k + rank)',
    note: 'Reciprocal Rank Fusion · 不依赖各路 score 数值尺度，名次决定贡献',
    example: 'doc#42  lex#3 + tok#2 + sem#5  →  RRF = 1/63 + 1/62 + 1/65 ≈ 0.0476',
  },
  surfaces: [
    { name: 'wiki_search', desc: 'searchLocalKnowledgePages → readLocalKnowledgePage' },
    { name: 'discover_data', desc: 'scan-report 上的实体探索（schema / table / column）' },
    { name: 'dictionary_search', desc: 'SL 模型 / measure / dimension 字典' },
    { name: 'entity_details', desc: '指定表 → 列 + FK + 抽样值（按 display ref 解析）' },
  ],
  index: [
    { table: 'wiki_chunks_fts', cols: ['page_id', 'body', 'tags'], note: 'FTS5 表，wiki 页切片后建索引' },
    { table: 'wiki_chunks_emb', cols: ['chunk_id', 'embedding'], note: '同表 sidecar，存嵌入向量' },
    { table: 'sl_entities', cols: ['kind', 'name', 'desc', 'embedding'], note: 'SL 模型/列/measure 字典' },
  ],
  embeddingProvider: {
    name: 'KtxEmbeddingPort',
    dim: '384d (本地 MiniLM) 或 OpenAI text-embedding-3-small',
    note: '默认走 daemon /embeddings/compute（sentence-transformers，离线、无外部依赖）',
  },
  insights: {
    input: '一句话同时打到 wiki（叙事知识）与 SL（结构化语义）——上下文层把两类资产放在一张检索网下，避免 Agent 自己拼接来源。',
    score: [
      { icon: 'i-lucide-tally-5', title: '三路打分各管一段', body: 'FTS5 抓精确词、token 抓短句、cosine 抓语义；不同查询形态各自贡献，互为兜底——单跑一路在 wiki/SL 混合语料上都不稳。' },
      { icon: 'i-lucide-cpu', title: 'SQLite 内嵌、零运维', body: '索引就是 .ktx/db.sqlite 的几张 FTS5 + sidecar 表；ktx 没有外部向量库依赖，索引随 wiki/SL 一起更新（增量重建）。' },
    ],
    fuse: [
      { icon: 'i-lucide-blocks', title: 'RRF 而非加权和', body: 'RRF 只用名次不用分数，避免不同 scorer 数值尺度差异搞乱融合；新增一路 scorer 也不用重调权重——可加性强。' },
    ],
  },
}

const mcpArch: McpArch = {
  id: 'mcp',
  input: {
    label: '外部 Agent · function-calling',
    note: 'ktx 不内置 LLM；任何支持 MCP 的 Agent 客户端都能用',
    clients: ['Claude Code (Claude Pro/Max sub)', 'Codex (本地 auth)', 'Cursor', 'OpenCode', 'Generic MCP host'],
  },
  transports: [
    { name: 'stdio', cmd: 'ktx mcp start --project-dir <p>', note: 'Agent 客户端通过 stdio 子进程拉起；ktx 进程内调度', icon: 'i-lucide-terminal' },
    { name: 'HTTP', cmd: 'ktx mcp http --port 7321', note: 'managed-mcp-daemon 守护，客户端通过 HTTP MCP 连接', icon: 'i-lucide-server' },
  ],
  tools: [
    { name: 'connection_list', desc: '列已配置的 connection / source', accent: 'slate', reads: 'ktx.yaml' },
    { name: 'wiki_search + wiki_read', desc: '混合检索 + 读取 wiki 页', accent: 'blue', reads: 'wiki/' },
    { name: 'entity_details', desc: '表 → 列 / FK / 抽样', accent: 'emerald', reads: 'raw-sources/<conn>/scan-report.json' },
    { name: 'discover_data', desc: '探索 schema / table / column', accent: 'emerald', reads: 'raw-sources/' },
    { name: 'dictionary_search', desc: 'SL 字典（model / measure / dimension）', accent: 'amber', reads: 'sl_entities (sqlite)' },
    { name: 'sl_read_source', desc: '读 SL source YAML', accent: 'amber', reads: 'semantic-layer/' },
    { name: 'sl_query', desc: 'measures + dimensions → 方言 SQL', accent: 'amber', reads: 'daemon /semantic-layer/query' },
    { name: 'sql_execution', desc: '只读 SQL 校验后执行', accent: 'rose', reads: 'daemon /sql/validate-read-only', writes: '禁写 — 仅 SELECT' },
    { name: 'memory_ingest', desc: '把对话尾事实写回 wiki/SL', accent: 'violet', writes: 'wiki/ + semantic-layer/（worktree）' },
  ],
  ports: [
    { name: 'McpServer', desc: 'context-tools.ts 注册中心 · types.ts 工具 schema' },
    { name: 'local-project-ports.ts', desc: '把本地项目接成统一端口集合（wiki / SL / scan / search / sl_query / memory）' },
    { name: 'KtxLlmRuntimePort', desc: 'memory_ingest 内部用（生成事实抽取 prompt）' },
  ],
  sampleCall: {
    tool: 'sl_query',
    req: '{\n  "measures": ["orders.total_revenue"],\n  "dimensions": ["customers.region", "orders.order_month"],\n  "filters": [{ "expr": "orders.order_date >= \\"2025-01-01\\"" }],\n  "limit": 1000\n}',
    resp: '{\n  "sql": "WITH orders AS (...) SELECT region, order_month, SUM(total) ...",\n  "dialect": "bigquery",\n  "columns": ["region", "order_month", "total_revenue"],\n  "plan": { "anchor": "orders", "joins": [...], "fanout": false }\n}',
  },
  insights: {
    input: 'MCP 是 ktx 服务 Agent 的主入口——读 wiki/SL 是工具调用、跑 SL 是工具调用、跑只读 SQL 是工具调用；Agent 不需要理解仓库结构。',
    surface: [
      { icon: 'i-lucide-lock', title: 'read-only by design', body: 'sql_execution 只接 SELECT（daemon validate-read-only 黑名单 Insert/Update/Delete/Alter/Create…）；写入只走 memory_ingest 的 worktree 路径——MCP 不可能直接污染仓库。' },
      { icon: 'i-lucide-blocks', title: '工具即原语', body: '每个工具职责单一（搜 / 读 / 解析 / 跑 SQL / 学习），Agent 自己编排顺序——和 WrenAI 把正确性拆成原语相同的哲学，但 ktx 的载体是 MCP 工具。' },
    ],
    safety: [
      { icon: 'i-lucide-server', title: '本地、零外发', body: 'MCP 服务跑在本机，唯一外发的就是你配置的 LLM provider。schema/SQL/数据不上传任何托管服务。' },
    ],
  },
}

export const KTX_MODULES: Record<string, KtxModuleData> = {
  ingest: { id: 'ingest', accent: 'emerald', ingest: ingestArch },
  search: { id: 'search', accent: 'blue', search: searchArch },
  mcp: { id: 'mcp', accent: 'violet', mcp: mcpArch },
  sl: { id: 'sl', accent: 'amber', sl: slArch },
  memory: { id: 'memory', accent: 'amber', memory: memoryArch },
  daemon: { id: 'daemon', accent: 'rose', daemon: daemonArch },
  storage: { id: 'storage', accent: 'indigo', storage: storageArch },
}

export function getKtxModule(id: string | null): KtxModuleData | null {
  if (!id) return null
  return KTX_MODULES[id] ?? null
}
