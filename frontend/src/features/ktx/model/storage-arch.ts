import type { StorageArch } from './ktx'

export const storageArch: StorageArch = {
  id: 'storage',
  input: {
    label: 'ktx project directory',
    note: 'ktx setup 创建 · 解析顺序：KTX_PROJECT_DIR → 最近 ktx.yaml → cwd · --project-dir 可脚本化',
  },
  paths: [
    {
      path: 'ktx.yaml',
      role: '项目配置',
      commit: '✓',
      icon: 'i-lucide-settings',
      accent: 'slate',
      desc: 'connections / llm / ingest / scan / storage 声明式配置',
    },
    {
      path: 'semantic-layer/<conn>/',
      role: 'YAML 语义源',
      commit: '✓',
      icon: 'i-lucide-box',
      accent: 'amber',
      desc: 'models / relationships / measures / segments · _schema/ manifest 分片',
    },
    {
      path: 'wiki/global/ · wiki/user/<id>/',
      role: 'Markdown 业务知识',
      commit: '✓',
      icon: 'i-lucide-book-marked',
      accent: 'emerald',
      desc: 'global 共享 · user 个人笔记 · 矛盾标注供人工 review',
    },
    {
      path: 'raw-sources/<conn>/<source>/<sync>/',
      role: 'Ingest 痕迹',
      commit: '✓',
      icon: 'i-lucide-archive',
      accent: 'blue',
      desc: 'scan-report.json · per-table JSON · enrichment/relationship-profile.json',
    },
    {
      path: '.ktx/db.sqlite',
      role: '本地索引 + 状态',
      commit: '✗',
      icon: 'i-lucide-database',
      accent: 'indigo',
      desc: 'FTS5 · embeddings cache · ingest runs · memory runs · SL 字典',
    },
    {
      path: '.ktx/worktrees/',
      role: '隔离分支',
      commit: '✗',
      icon: 'i-lucide-git-fork',
      accent: 'violet',
      desc: 'ingest WU / memory agent 的 session worktree',
    },
  ],
  ktxYaml: `# ktx.yaml (commit ✓)
connections:
  warehouse:
    driver: postgres
    host: \${PG_HOST}
    database: analytics
sources:
  dbt_main:
    type: dbt
    connection: warehouse
llm:
  provider: anthropic
  model: claude-sonnet-4-6
embeddings:
  provider: openai
  model: text-embedding-3-small`,
  sqliteTables: [
    { name: 'wiki_chunks_fts', desc: 'FTS5 倒排索引（wiki 页切片）' },
    { name: 'wiki_chunks_emb', desc: 'wiki chunk 嵌入 sidecar' },
    { name: 'sl_entities', desc: 'SL model/column/measure 字典 + embedding' },
    { name: 'ingest_runs', desc: 'ingest bundle 运行记录 + 状态' },
    { name: 'memory_runs', desc: 'memory agent 异步 run 状态 + captured actions' },
    { name: 'enrichment_state', desc: 'scan 增量丰富进度（避免重复 LLM 调用）' },
  ],
  resolution: [
    { name: 'KTX_PROJECT_DIR', desc: '环境变量优先' },
    { name: 'nearest ktx.yaml', desc: '向上遍历目录树' },
    { name: '--project-dir <path>', desc: 'CLI 显式指定（脚本化）' },
  ],
  insights: {
    input: 'ktx 的存储模型是"Git 化语义层 + 本地索引"——可评审的工件进仓库，索引和状态留 .ktx/（git-ignored）。',
    git: [
      { icon: 'i-lucide-git-pull-request', title: '语义层像代码一样评审', body: 'wiki/ + semantic-layer/ 全是可读文本（Markdown / YAML），变更走 git diff；矛盾标注、measure 定义、业务规则都可被 data team review。' },
      { icon: 'i-lucide-key-round', title: '凭据不进仓库', body: '.ktx/ 和连接密钥全部 git-ignored；项目目录可安全共享给团队，每人本地配自己的 warehouse 连接。' },
    ],
    state: [
      { icon: 'i-lucide-database', title: 'SQLite 零运维索引', body: 'FTS5 + embedding cache + run 记录全在 .ktx/db.sqlite；无外部向量库、无 Redis——ktx 开箱即用，索引随 wiki/SL 增量重建。' },
    ],
  },
}
