/**
 * Dataflow definitions for the module (L1) drill-down views.
 * Each flow is a sequence of steps describing how data moves through a pipeline.
 * Kept separate from architecture.ts so module visualizations stay data-driven.
 */
import type { AccentKey } from './architecture'

export interface FlowArtifact {
  /** what the step reads */
  input?: string
  /** what the step produces */
  output?: string
  /** storage table / tool it touches */
  store?: string
  /** representative code / SQL snippet */
  code?: string
  /** language hint for the snippet */
  lang?: 'sql' | 'json' | 'text'
}

export interface FlowStep {
  id: string
  title: string
  subtitle: string
  icon: string
  accent: AccentKey
  /** one-line summary shown in the rail */
  summary: string
  /** longer explanation shown in the detail panel */
  detail: string
  artifact: FlowArtifact
}

export interface FlowDef {
  id: string
  label: string
  title: string
  subtitle: string
  icon: string
  accent: AccentKey
  steps: FlowStep[]
}

const onboarding: FlowDef = {
  id: 'onboarding',
  label: 'Onboarding',
  title: 'Database Onboarding',
  subtitle: '接入一个新数据库时，ATLAS 如何从零自动构建 Rich Context 与向量索引',
  icon: 'i-lucide-database-zap',
  accent: 'emerald',
  steps: [
    {
      id: 'introspect',
      title: 'Schema Introspection',
      subtitle: '物理结构抽取',
      icon: 'i-lucide-scan-search',
      accent: 'emerald',
      summary: '从 INFORMATION_SCHEMA 读取表 / 列 / 外键',
      detail:
        '连接目标库，遍历 INFORMATION_SCHEMA 抽取所有表、列（类型 / 可空 / 主键 / 外键）以及外键关系，落地为结构化元数据，作为后续一切处理的骨架。',
      artifact: {
        input: '目标数据库连接',
        output: 'TableInfo / ColumnInfo / Relation',
        store: 'rc_tables · rc_columns · rc_relations',
        lang: 'sql',
        code: `SELECT table_name, column_name, data_type,
       is_nullable, column_key
FROM   information_schema.columns
WHERE  table_schema = :db;`,
      },
    },
    {
      id: 'decompose',
      title: 'Forest Decompose',
      subtitle: '按外键图拆分子簇',
      icon: 'i-lucide-git-fork',
      accent: 'emerald',
      summary: '外键图 → 连通分量 → 表簇 + 预算',
      detail:
        '把外键关系建成无向图，用 BFS 求连通分量，将数据库拆成若干「表簇」；孤立表按批合并，并为每个簇计算 ReAct 迭代预算（~3 次/表）。大库由此被切成可并行、可控成本的小块。',
      artifact: {
        input: 'Relation[] (外键)',
        output: 'TableCluster[] + 迭代预算',
        store: 'in-memory',
        lang: 'text',
        code: `FK Graph ──BFS──▶ 连通分量
  cluster#0  orders─order_items─products
  cluster#1  users─addresses
  isolated   logs, configs … → 按批合并
budget = tables*3 + 10  (cap 60~500)`,
      },
    },
    {
      id: 'explore',
      title: 'ReAct Exploration',
      subtitle: 'Agent 实地探查数据',
      icon: 'i-lucide-bot',
      accent: 'emerald',
      summary: '每簇内 Reason→Act→Observe 迭代探查',
      detail:
        '为每个表簇构建 Onboarding ReAct Engine，Agent 用 execute_sql 工具实地探查：SELECT * LIMIT 5 看样例、GROUP BY 看取值分布、检查 NULL / 空白等数据质量问题。绝不凭空猜测，先看真实数据再写上下文。',
      artifact: {
        input: 'TableCluster + system prompt',
        output: '探查观测 (样例行 / 取值分布)',
        store: 'tool: execute_sql',
        lang: 'sql',
        code: `SELECT * FROM orders LIMIT 5;
SELECT status, COUNT(*)
FROM   orders GROUP BY status;`,
      },
    },
    {
      id: 'richcontext',
      title: 'Rich Context Generation',
      subtitle: '生成语义上下文',
      icon: 'i-lucide-sparkles',
      accent: 'emerald',
      summary: 'set_rich_context 批量写 5 类语义信息',
      detail:
        '基于探查结果，Agent 调 set_rich_context（批量模式）写入五类 Rich Context：表描述、列描述、样例值、列同义词、业务术语。这些语义信息让 Text-to-SQL 真正「读懂」业务，而非只看裸表名。',
      artifact: {
        input: '探查观测',
        output: '5 类 Rich Context',
        store: 'rc_business_context',
        lang: 'json',
        code: `[
  {"type":"table_description","table":"orders","value":"订单主表…"},
  {"type":"column_description","table":"orders","column":"status","value":"订单状态"},
  {"type":"column_sample_values","table":"orders","column":"status","value":"paid, shipped, refunded"},
  {"type":"column_synonyms","table":"orders","column":"amount","value":"金额, 订单额"}
]`,
      },
    },
    {
      id: 'embed',
      title: 'Vector Embedding',
      subtitle: '向量化并建索引',
      icon: 'i-lucide-radar',
      accent: 'emerald',
      summary: 'Doubao 嵌入 2048d → HNSW 索引',
      detail:
        '把每条 Rich Context 用 Doubao Embedding 编码成 2048 维向量，写入 rc_embeddings（MariaDB 原生 VECTOR 列），构建 COSINE HNSW 索引。至此新库完成接入，推理阶段即可亚毫秒级向量召回相关表列。',
      artifact: {
        input: 'rc_business_context 文本',
        output: 'VECTOR(2048) + HNSW 索引',
        store: 'rc_embeddings',
        lang: 'sql',
        code: `INSERT INTO rc_embeddings (entity, content, vec)
VALUES (:entity, :content, VEC_FromText(:v));
-- 原生 HNSW：COSINE 距离，亚毫秒召回`,
      },
    },
  ],
}

export const flows: FlowDef[] = [onboarding]

export function getFlow(id: string | null): FlowDef | null {
  if (!id) return null
  return flows.find((f) => f.id === id) ?? null
}
