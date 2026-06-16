// OKF (Open Knowledge Format) deck — GoogleCloudPlatform/knowledge-catalog.
//
// Scope: the OKF spec itself + a proof-of-concept enrichment_agent that
// produces OKF bundles from BigQuery / Web, plus a static HTML viewer
// that graph-renders the bundle as a navigable force-directed graph.
//
// Note: the project as a whole is *not* an NL2SQL agent — it is a
// *metadata enrichment* toolkit. OKF is the central artifact: a
// vendor-neutral format for representing catalog knowledge as
// plain markdown files with YAML frontmatter, git-versioned.
//
// Reference: https://github.com/GoogleCloudPlatform/knowledge-catalog
// Spec: okf/SPEC.md
// Agent: okf/src/enrichment_agent/{agent,runner,cli}.py
// Viewer: okf/src/enrichment_agent/viewer/generator.py

import { ACCENTS, type Accent, type AccentKey, type ArchLayer, type ArchNode } from '../../arch/model/architecture'
import type { Insight, NamedItem } from '../../arch/model/modules'

export { ACCENTS }
export type { Accent, AccentKey, ArchLayer, ArchNode, Insight, NamedItem }

/* ─── L0 panorama: the OKF stack — format + producer + consumer ─── */
export const OKF_LAYERS: ArchLayer[] = [
  {
    id: 'format',
    title: '① Format · 规范本体',
    subtitle: 'vendor-neutral 的元数据目录结构（YAML frontmatter + Markdown body）',
    icon: 'i-lucide-file-stack',
    accent: 'slate',
    cols: 3,
    nodes: [
      {
        id: 'concept',
        label: 'Concept · 一个 .md 一个知识单元',
        sublabel: 'YAML frontmatter(type/title/desc/resource/tags) + Markdown body',
        icon: 'i-lucide-file-text',
        accent: 'slate',
        flow: 'concept',
        span: 1,
        codeRefs: ['okf/SPEC.md §4', 'okf/bundles/ga4/tables/events_.md'],
      },
      {
        id: 'bundle',
        label: 'Bundle · 目录树',
        sublabel: '一组 concepts 组成 hierarchical 可分发的知识包；index.md / log.md 保留',
        icon: 'i-lucide-folder-tree',
        accent: 'slate',
        flow: 'bundle',
        span: 1,
        codeRefs: ['okf/SPEC.md §3', 'okf/src/enrichment_agent/bundle/'],
      },
      {
        id: 'linking',
        label: 'Cross-linking',
        sublabel: '绝对路径 /tables/... 链接表达概念间关系，broken link 容忍',
        icon: 'i-lucide-link-2',
        accent: 'slate',
        flow: 'linking',
        span: 1,
        codeRefs: ['okf/SPEC.md §5'],
      },
    ],
  },
  {
    id: 'producer',
    title: '② Producer · enrichment_agent',
    subtitle: 'Google ADK + Gemini · 从 BigQuery / Web 读源，写 OKF bundle',
    icon: 'i-lucide-bot',
    accent: 'violet',
    cols: 3,
    nodes: [
      {
        id: 'bq-agent',
        label: 'BQ Enrichment Agent',
        sublabel: 'ADK agent · 5 tools: list/read_concept/sample_rows/read_existing_doc/write_concept_doc',
        icon: 'i-lucide-database',
        accent: 'emerald',
        flow: 'bq-agent',
        span: 1,
        codeRefs: [
          'okf/src/enrichment_agent/agent.py::build_bq_agent',
          'okf/src/enrichment_agent/sources/bigquery.py',
          'okf/src/enrichment_agent/tools/bundle_tools.py',
        ],
      },
      {
        id: 'web-agent',
        label: 'Web Ingestion Agent',
        sublabel: '独立 agent · fetch_url + write_concept_doc · 受 max_pages/hosts/prefix 约束',
        icon: 'i-lucide-globe',
        accent: 'blue',
        flow: 'web-agent',
        span: 1,
        codeRefs: [
          'okf/src/enrichment_agent/agent.py::build_web_agent',
          'okf/src/enrichment_agent/tools/web_tools.py',
        ],
      },
      {
        id: 'runner',
        label: 'EnrichmentRunner',
        sublabel: '编排：list → enrich_all → web_pass → regenerate_indexes',
        icon: 'i-lucide-workflow',
        accent: 'violet',
        flow: 'runner',
        span: 1,
        codeRefs: [
          'okf/src/enrichment_agent/runner.py::EnrichmentRunner',
          'okf/src/enrichment_agent/cli.py',
        ],
      },
    ],
  },
  {
    id: 'storage',
    title: '③ Storage & Layout',
    subtitle: 'bundle 落盘 = 普通 git 仓库；文件 = 资产，frontmatter = 索引，index.md = 导航',
    icon: 'i-lucide-archive',
    accent: 'emerald',
    cols: 3,
    nodes: [
      {
        id: 'fs',
        label: '文件系统 / Git',
        sublabel: '无私有后端；diff / blame / PR 走标准 git 工作流',
        icon: 'i-lucide-git-branch',
        accent: 'emerald',
        flow: 'fs',
        span: 1,
        codeRefs: ['okf/SPEC.md §3'],
      },
      {
        id: 'index',
        label: 'index.md · 渐进式披露',
        sublabel: 'directory listing · 头部排序按 frontmatter.title/description',
        icon: 'i-lucide-list',
        accent: 'emerald',
        flow: 'index',
        span: 1,
        codeRefs: [
          'okf/src/enrichment_agent/bundle/index.py::regenerate_indexes',
          'okf/src/enrichment_agent/bundle/synthesizer.py',
        ],
      },
      {
        id: 'log',
        label: 'log.md · 变更追溯',
        sublabel: '按 ISO 日期分组 · Update / Creation / Deprecation',
        icon: 'i-lucide-history',
        accent: 'emerald',
        flow: 'log',
        span: 1,
        codeRefs: ['okf/SPEC.md §7'],
      },
    ],
  },
  {
    id: 'consumer',
    title: '④ Consumer · 谁来读',
    subtitle: '由 ① → ④ 一路下来，bundle 既是产物也是人 / Agent / 第三方都能消费的载体',
    icon: 'i-lucide-eye',
    accent: 'amber',
    cols: 3,
    nodes: [
      {
        id: 'human',
        label: '人类 · Obsidian / MkDocs',
        sublabel: 'cat / grep / vim / Obsidian 直接打开，git diff 评审',
        icon: 'i-lucide-user',
        accent: 'amber',
        flow: 'human',
        span: 1,
        codeRefs: ['okf/README.md (Why OKF?)'],
      },
      {
        id: 'agent',
        label: '下游 Agent · 注入 context',
        sublabel: '文件级粒度把相关 concept 塞进 prompt · 走 chunk/embed/FTS 都能消费',
        icon: 'i-lucide-cpu',
        accent: 'amber',
        flow: 'agent',
        span: 1,
        codeRefs: ['okf/README.md'],
      },
      {
        id: 'viewer',
        label: 'Graph Viewer · viz.html',
        sublabel: '静态生成 force-directed 图谱 · 看 concept 间 link 关系',
        icon: 'i-lucide-share-2',
        accent: 'amber',
        flow: 'viewer',
        span: 1,
        codeRefs: [
          'okf/src/enrichment_agent/viewer/generator.py',
          'okf/bundles/ga4/viz.html',
        ],
      },
    ],
  },
]

/* ─── L1 flow registry — wiring between L0 nodes and detailed diagrams ─── */
export interface OkfFlowDef {
  id: string
  label: string
  title: string
  subtitle: string
  icon: string
  accent: AccentKey
}

export const okfFlows: OkfFlowDef[] = [
  { id: 'concept',     label: 'Concept',     title: '概念文档 · 一个 .md 一份知识',     subtitle: 'YAML frontmatter(type/title/desc/resource/tags) + Markdown body. Conventional headings: # Schema / # Examples / # Citations.', icon: 'i-lucide-file-text',     accent: 'slate'   },
  { id: 'bundle',      label: 'Bundle',      title: 'Bundle 目录结构',                     subtitle: 'index.md / log.md 保留名；其它 .md 都是 concept. 一个 bundle = 一个 git 仓库（或 tarball）.', icon: 'i-lucide-folder-tree', accent: 'slate'   },
  { id: 'linking',     label: 'Linking',     title: '概念间链接',                          subtitle: '相对 / 绝对 markdown 链接 · broken link 容忍 · 周围 prose 描述关系类型',                    icon: 'i-lucide-link-2',       accent: 'slate'   },
  { id: 'bq-agent',    label: 'BQ Agent',    title: 'BigQuery enrichment agent',          subtitle: 'ADK + Gemini. 5 tools: list_concepts / read_concept_raw / sample_rows / read_existing_doc / write_concept_doc.', icon: 'i-lucide-database', accent: 'emerald' },
  { id: 'web-agent',   label: 'Web Agent',   title: 'Web ingestion agent',                subtitle: '独立 agent · 同样的 concept write 工具 + fetch_url · 受 max_pages/hosts/path_prefixes 约束',            icon: 'i-lucide-globe',        accent: 'blue'    },
  { id: 'runner',      label: 'Runner',      title: 'EnrichmentRunner 编排',              subtitle: 'list concepts → enrich_all → web_pass → regenerate_indexes. 单个 concept enrich = 一次新 ADK session', icon: 'i-lucide-workflow', accent: 'violet'  },
  { id: 'fs',          label: 'Filesystem',  title: 'bundle 落盘到普通 git 仓库',          subtitle: '无私有存储后端 · 文件 = 资产 · frontmatter = 索引 · index.md = 导航 · 走标准 git diff/ blame / PR', icon: 'i-lucide-git-branch', accent: 'emerald' },
  { id: 'index',       label: 'index.md',    title: '目录的渐进式披露',                    subtitle: '无 frontmatter · 用 section heading 分组 · 条目含 title + 来自 frontmatter.description 的简介',          icon: 'i-lucide-list',         accent: 'emerald' },
  { id: 'log',         label: 'log.md',      title: '变更日志',                            subtitle: 'ISO 8601 日期分组 · 顺序新→旧 · 引导词 Update / Creation / Deprecation 是约定非强约束',                  icon: 'i-lucide-history',      accent: 'emerald' },
  { id: 'human',       label: '人类消费',    title: '人直接 cat / Obsidian / MkDocs',      subtitle: '无 SDK 阻拦 · 普通 IDE 工具链即可 · git diff 提供评审界面',                                          icon: 'i-lucide-user',         accent: 'amber'   },
  { id: 'agent',       label: 'Agent 消费',  title: '下游 LLM Agent 把 concept 注入 prompt', subtitle: 'FTS / chunk / embed / link 任何召回策略都能消费 · 无需专用 SDK',                                    icon: 'i-lucide-cpu',         accent: 'amber'   },
  { id: 'viewer',      label: 'Graph viewer', title: 'viz.html force-directed graph',     subtitle: 'view/generator.py 把 bundle 渲染成单文件 HTML · 节点 = concept · 边 = link · 完全静态可托管',         icon: 'i-lucide-share-2',     accent: 'amber'   },
]

export function getOkfFlow(id: string | null): OkfFlowDef | null {
  if (!id) return null
  return okfFlows.find((f) => f.id === id) ?? null
}
