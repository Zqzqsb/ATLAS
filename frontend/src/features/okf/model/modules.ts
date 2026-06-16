// L1 module data — internal architecture for each OKF drill-down.
// Shape mirrors arch/model/modules.ts (insights + named items + sections).

import type { Insight, NamedItem } from '../../arch/model/modules'

/** All L1 stages keyed by OkfFlowDef.id. */
export const OKF_MODULES: Record<string, OkfModuleArch> = {
  /* ────────────────── FORMAT layer ────────────────── */
  concept: {
    id: 'concept',
    abstract: 'OKF 最小知识单元：一份 .md 文件 = 一个 concept. 由 frontmatter + body 组成，是所有上层能力的基元。',
    principles: [
      { name: 'Frontmatter is the contract',  desc: 'type 必填（消费方路由键）；title/desc/resource/tags 推荐；未知字段保留不报错。' },
      { name: 'Body 是结构化 Markdown',        desc: '提倡 # Schema / # Examples / # Citations 等约定 heading；freeform 也接受。' },
      { name: 'Path is identity',              desc: '概念 ID = 文件相对 bundle 路径去 .md 后缀；可重命名文件，跨文件 link 用绝对 /path/... 仍能命中。' },
      { name: 'Extension-first',               desc: '消费者必须优雅容忍未知 type / 未知 frontmatter 字段，不能因 schema 升级而拒绝旧 bundle。' },
    ],
    sections: [
      {
        kind: 'anatomy',
        title: 'Anatomy · 一个 concept 长什么样',
        rows: [
          { label: 'frontmatter',     detail: '--- 分隔的 YAML 块；type 必填，其余可选。', accent: 'slate'   },
          { label: 'body',            detail: '自由 Markdown；推荐结构化（heading / table / code fence）。',  accent: 'emerald' },
          { label: 'conventional headings', detail: '# Schema · 列与类型；# Examples · 用法示例；# Citations · 外部证据。', accent: 'amber' },
        ],
      },
      {
        kind: 'example',
        title: '真实 concept · 来自 bundles/ga4/tables/events_.md',
        code: `---
type: BigQuery Table
title: GA4 Events
description: One row per GA4 event (pageview, purchase, …) from an obfuscated e-commerce sample.
resource: bigquery://ga4-obfuscated-sample-ecommerce.analytics_249147898.events_
tags: [ga4, events, web-analytics]
timestamp: 2026-05-28T14:30:00Z
---

# Schema

| Column         | Type      | Description                                  |
|----------------|-----------|----------------------------------------------|
| event_name     | STRING    | e.g. "page_view", "purchase", "add_to_cart". |
| event_date     | DATE      | UTC date the event was logged.               |
| user_pseudo_id | STRING    | GA4 client id (pseudonymous).                |
| ...            | ...       | ...                                          |

# Joins

Joined with [users](/tables/users.md) on \`user_pseudo_id\`.

# Citations

[1] [GA4 BigQuery export schema](https://support.google.com/analytics/answer/7029846)`,
      },
    ],
    insights: [
      { icon: "i-lucide-sparkles", title: '一个文件 = 一个 entity', body: 'OKF 的"实体"边界就是文件系统边界——这跟传统元数据 catalog 把 entity 塞进一行 SQL 行不同，能直接 cat / grep / diff，' +
        '所有版本控制 / 审计 / 工具链白送。' },
      { icon: "i-lucide-sparkles", title: 'Path-as-ID',                 body: '概念 ID = 文件路径。这意味着 git mv / refactor 全部天然支持，不会出现传统 catalog 里 rename 改 ID 然后所有 link 断掉的灾难。' },
      { icon: "i-lucide-sparkles", title: 'Frontmatter 是"小 schema"',   body: 'type 是路由键（消费方按 type 决定怎么渲染 / 怎么索引进 embedding），其它字段是 KV 扩展——这是 schema-on-read 而不是 schema-on-write。' },
    ],
  },

  bundle: {
    id: 'bundle',
    abstract: 'Bundle = 一棵 .md 树，是 OKF 的可分发单元。Reserved file: index.md（目录列表）/ log.md（变更日志）。',
    principles: [
      { name: '目录即层级',                  desc: '父子关系 = 目录嵌套；不强制任何固定层级。' },
      { name: '两个保留文件名',              desc: 'index.md / log.md 在任何层都有意义，不能用作 concept 文件名。' },
      { name: '分发形态任意',                desc: 'git 仓库（推荐）/ tarball / 子目录；bundle 本身自描述。' },
      { name: 'Unknown type 容忍',            desc: '生产者可加任意 type（Metric / Playbook / …），消费者必须能优雅处理未知 type。' },
    ],
    sections: [
      {
        kind: 'tree',
        title: 'Tree · bundle 的真实结构（bundles/ga4/）',
        tree: `ga4/
├── index.md
├── datasets/
│   ├── index.md
│   └── ga4_obfuscated_sample_ecommerce.md
├── tables/
│   ├── index.md
│   └── events_.md
├── references/
│   └── index.md
└── viz.html                ← 由 viewer/generator.py 静态生成`,
      },
    ],
    insights: [
      { icon: "i-lucide-sparkles", title: 'Bundle = 完整 self-contained 包', body: '把整个 catalog 装在一个目录里 → 拷贝 / 打包 / 跨组织交换变成 0 摩擦动作（tar 一下就完事）。' },
      { icon: "i-lucide-sparkles", title: 'Subdirectory 自由',              body: '你按业务怎么舒服怎么组织——按表 / 按域 / 按团队，没有 schema 强约束。这就是 OKF 跟传统 metadata catalog 的关键差异。' },
      { icon: "i-lucide-sparkles", title: 'viz.html = 第一公民 artifact',  body: '一个静态 HTML 就是 bundle 的可视化入口，可托管在 GCS / S3 / Pages 上。无需启动后端服务。' },
    ],
  },

  linking: {
    id: 'linking',
    abstract: 'OKF 概念间的关系表达：标准 markdown link + 周围 prose 描述关系类型。',
    principles: [
      { name: 'Markdown link 即关系',          desc: '标准 `[A](/tables/x.md)` 形式；不引入新语法。' },
      { name: '推荐绝对 / 路径',              desc: '以 / 开头，bundle 根为基准，重命名目录时 link 不破。' },
      { name: '关系类型由 prose 决定',          desc: '"joins on" / "depends on" / "see also" 由周围文字表达，link 本身无语义。' },
      { name: 'Broken link 不报错',            desc: '链接的目标可能还没写——消费者 MUST 容忍 broken link。' },
    ],
    sections: [],
    insights: [
      { icon: "i-lucide-sparkles", title: '无 link 类型系统',     body: 'OKF 故意不做"typed edge"——这跟传统知识图谱 / KG 看起来很弱，但好处是' +
        '任何 markdown 工具都能消费 bundle，无需专用渲染器。' },
      { icon: "i-lucide-sparkles", title: 'Broken link = TODO 语义', body: '写文档时引用一个还没建的概念是常见场景——OKF 把这种"未来知识"作为一等公民，而不是错误。' },
    ],
  },

  /* ────────────────── PRODUCER layer ────────────────── */
  'bq-agent': {
    id: 'bq-agent',
    abstract: 'BQ enrichment agent = Google ADK + Gemini，给一个 BigQuery dataset / 表生成 OKF concept 文档。',
    principles: [
      { name: 'ADK 5 tools 模式',              desc: 'list_concepts / read_concept_raw / sample_rows / read_existing_doc / write_concept_doc — 5 个 tool 就能循环 ① 查 ② 读 ③ 写。' },
      { name: '每个 concept 一次 session',       desc: 'enrich_concept() 为每个 ref 创建独立 ADK session，session_id 唯一。失败/重试不污染其它 concept。' },
      { name: '真实数据驱动',                  desc: 'sample_rows + INFORMATION_SCHEMA 提供 ground truth；LLM 不是凭空"猜"schema。' },
      { name: '写 = 落盘 + frontmatter',        desc: 'write_concept_doc 接收 title/description/tags/type，agent 负责决定其余字段。' },
    ],
    sections: [
      {
        kind: 'loop',
        title: 'Agent Loop · enrich_concept 内部',
        steps: [
          { name: 'list_concepts',     desc: '从 Source 拿所有 ConceptRef（dataset / table / …）' },
          { name: 'read_concept_raw',  desc: 'INFORMATION_SCHEMA schema + 类型 + nullable 等元数据' },
          { name: 'sample_rows',       desc: 'LIMIT N 真实样本，避免 LLM 瞎编枚举值 / 字符串格式' },
          { name: 'read_existing_doc', desc: '若文件已存在，agent 看到的是"修订"而不是"新建"' },
          { name: 'write_concept_doc', desc: '生成 frontmatter + body 写回；类型 = ConceptRef.type 锁定' },
        ],
      },
    ],
    insights: [
      { icon: "i-lucide-sparkles", title: 'Tool 设计是 ADK 范式',     body: '5 个 tool 全部是 FunctionTool，没有任何魔法——可以直接打开 agent.py 看到全部 wiring，方便 debug / unit test。' },
      { icon: "i-lucide-sparkles", title: '为什么 sample_rows 必要',    body: 'INFORMATION_SCHEMA 给的是结构，sample_rows 给的是"这一列的取值长什么样"（enum 集合 / 字符串前缀 / 时间格式）。' +
        '两者结合 → LLM 写出的 description 才不会"想了 5 秒然后胡编"。' },
    ],
  },

  'web-agent': {
    id: 'web-agent',
    abstract: 'Web ingestion agent = 独立 ADK agent，从 seed URL 出发按白名单爬取 + 写入 references/。',
    principles: [
      { name: '独立 agent',                  desc: 'build_web_agent 与 bq_agent 是两个 ADK agent，工具集不完全重合（少了 sample_rows，多了 fetch_url）。' },
      { name: '硬限制内置在 tool 内',          desc: 'fetch_url 内部 enforce max_pages / allowed_hosts / path_prefixes / denied_substrings — 不是 prompt 约束。' },
      { name: 'Web pass = 二阶段',            desc: 'enrich_all() 全部跑完后才 run_web_pass()，避免 web 拉来的内容污染 bq 概念。' },
      { name: 'Boundary 优先',                desc: 'agent 主动判断"这一页是 enrich 现有 concept / 还是新 references/<slug> / 还是 skip"，skip 优于 borderline fetch。' },
    ],
    sections: [],
    insights: [
      { icon: "i-lucide-sparkles", title: 'Tool 内置 hard limit',  body: '把安全 / 配额约束放进 fetch_url 内部实现而不是 prompt，是 OKF / enrichment_agent 的一个最佳实践——' +
        'LLM 不会试图绕过自己。' },
      { icon: "i-lucide-sparkles", title: 'References 是 OKF 一等公民', body: 'references/<slug>.md 跟 tables/<x>.md 在 OKF 里是平等的 concept，只是 type 不同。这让 web 信息自然进入同一种召回机制。' },
    ],
  },

  runner: {
    id: 'runner',
    abstract: 'EnrichmentRunner = CLI 入口的编排层：list concepts → 逐个 enrich → web pass → regenerate indexes。',
    principles: [
      { name: '线性 4 段',                    desc: 'list → enrich_all → web_pass → regenerate_indexes，阶段间有清晰边界。' },
      { name: 'Per-concept session',            desc: 'enrich_concept() 每次创建新 ADK session（uuid session_id），失败可重试不污染。' },
      { name: 'Verbose 控制',                  desc: 'log 默认 compact（函数签名 + 返回 compact），--verbose 切 full JSON。' },
      { name: 'CLI thin wrapper',              desc: 'cli.py 大部分是 argparse + dispatch 到 runner，几乎不持业务逻辑。' },
    ],
    sections: [
      {
        kind: 'flow',
        title: 'Pipeline · enrich_all() 实际步骤',
        steps: [
          { name: 'list concepts',           desc: 'source.list_concepts() → List[ConceptRef]' },
          { name: 'for each ref',            desc: 'enrich_concept(ref)：新 session → 5 tool loop → 写盘' },
          { name: 'web pass (if seeds)',     desc: 'run_web_pass()：新 session → fetch_url + write_concept_doc 循环' },
          { name: 'regenerate indexes',      desc: '对每个目录用 LLM 重新生成 index.md（基于 frontmatter + 关系）' },
        ],
      },
    ],
    insights: [
      { icon: "i-lucide-sparkles", title: '单线程 sequential',     body: 'enrich_all 默认按 list 顺序逐个 enrich，没有并发——这是有意为之：避免多个 agent 并发写同一 bundle 引发的 IO / index 竞态。' },
      { icon: "i-lucide-sparkles", title: 'Index 重生成是 LLM 任务', body: 'index.md 不是简单的"列文件"——是 LLM 看到所有 frontmatter 后重新组织 sections / 描述。这比 ls | sort 智能很多。' },
    ],
  },

  /* ────────────────── STORAGE layer ────────────────── */
  fs: {
    id: 'fs',
    abstract: 'Bundle 落盘到普通文件系统 / git 仓库。无私有后端，git diff / blame / PR 全部白送。',
    principles: [
      { name: '文件 = 资产',                  desc: '一张表 / 一个 metric / 一个 playbook = 一个 .md，无元数据表外存储。' },
      { name: 'Frontmatter = 索引',           desc: '消费方不需要数据库就能 grep / jq 出所有 type=Metric 的概念。' },
      { name: 'git = 版本控制',                desc: 'PR 评审 / blame / revert / 跨 fork 全部走标准工具链。' },
      { name: '可移植',                       desc: 'tar 一下就带走；mount 一下就分享；sync 一下就跨组织交换。' },
    ],
    sections: [],
    insights: [
      { icon: "i-lucide-sparkles", title: '0 vendor lock-in', body: '传统 metadata catalog 一旦不用了，元数据要 ETL 出来才能带走；OKF bundle 永远在文件层，' +
        '迁移到任何系统都只需写一次 consumer 适配。' },
    ],
  },

  index: {
    id: 'index',
    abstract: 'index.md = bundle 目录的渐进式披露入口。无 frontmatter，用 heading 分组列出当前目录的 concept。',
    principles: [
      { name: '无 frontmatter',                desc: 'index.md 自身不是 concept，不能被 list_concepts 当成 entity 返回。' },
      { name: 'Section 标题 = 分组',            desc: '一个 # heading 下一组 list 条目；可多 section。' },
      { name: 'description 必填',              desc: '条目格式 `* [Title](url) - short desc`，desc 来自 concept frontmatter。' },
      { name: '可 LLM 生成',                  desc: 'regenerate_indexes() 用同一个 model 重新生成 index.md，保证描述跟实际 frontmatter 同步。' },
    ],
    sections: [
      {
        kind: 'example',
        title: '真实 index.md · bundles/ga4/tables/index.md',
        code: `# Tables

* [Events](events_.md) - One row per GA4 event (pageview, purchase, add_to_cart, ...) from the obfuscated e-commerce sample.
* [Users](users.md) - One row per pseudonymous GA4 user with first/last seen timestamps.
* [Sessions](sessions.md) - One row per GA4 session, derived from events by sessionStart.

# References

* [GA4 BigQuery export schema](references/ga4-bq-export.md) - Canonical column reference for \`events_*\` tables.`,
      },
    ],
    insights: [
      { icon: "i-lucide-sparkles", title: 'Index 是 LLM 写的',         body: 'regenerate_indexes 让 description 跟 frontmatter 永远同步——不会出现"index 上写的是旧描述，文件本身是新描述"的不一致。' },
      { icon: "i-lucide-sparkles", title: '无 frontmatter 的刻意设计',   body: 'index.md 不算 concept，是为了"它永远反映目录当前状态"——你不能"在 index 里写旧描述而不被发现"。' },
    ],
  },

  log: {
    id: 'log',
    abstract: 'log.md = bundle 目录的变更历史。ISO 8601 日期分组，顺序新→旧，引导词 Update / Creation 是约定。',
    principles: [
      { name: '可选文件',                     desc: '任何层都可放 log.md，无则不强制。' },
      { name: 'ISO 日期分组',                 desc: '## 2026-05-22 这种 heading 作为时间锚点。' },
      { name: 'Prose 条目',                    desc: '* **Update**: ... 这种 bullet 形式；bold 引导词是约定不是要求。' },
      { name: '新→旧 顺序',                    desc: 'head 里是最新的，越往下越旧；consumer 容易展示最近变更。' },
    ],
    sections: [],
    insights: [
      { icon: "i-lucide-sparkles", title: 'changelog as a concept',  body: '把"什么时候改了什么"也当成 first-class artifact 落进 git，而不是 metadata DB 里的一个 audit log 表——保证 audit 数据跟内容一起被 review / revert。' },
    ],
  },

  /* ────────────────── CONSUMER layer ────────────────── */
  human: {
    id: 'human',
    abstract: '人类消费 OKF = 直接 cat / Obsidian / MkDocs。所有 IDE 工具链白送，git diff 是评审界面。',
    principles: [
      { name: 'cat 即读',                       desc: '无 SDK / query language 阻拦，工程师可以 cat 任何 concept 看到全文。' },
      { name: 'Obsidian / Notion 兼容',          desc: 'markdown + YAML frontmatter 是这俩工具的 native 格式；双击 .md 即可编辑。' },
      { name: 'MkDocs / Docusaurus',             desc: '直接把 bundle 挂上去就是一个静态文档站。' },
      { name: 'git PR 工作流',                   desc: 'enrich 一次 / 改一个 concept → PR → 评审 → merge，全部走标准代码评审。' },
    ],
    sections: [],
    insights: [
      { icon: "i-lucide-sparkles", title: '0 学习曲线',         body: '团队成员只要会 markdown + git 就能消费 + 贡献；不需要学"catalog 怎么用"——这是 OKF 拉低 metadata 贡献门槛的关键。' },
    ],
  },

  agent: {
    id: 'agent',
    abstract: '下游 LLM Agent 消费 OKF = 把 concept 注入 prompt。FTS / chunk / embed / link 任何召回策略都能用。',
    principles: [
      { name: '文件级粒度',                     desc: '消费方一次只需读相关 concept，不需要 catalog-wide 加载。' },
      { name: 'Frontmatter 加速过滤',           desc: '读 frontmatter 判断 type/title/tags → 决定是否深读 body。' },
      { name: 'No SDK',                        desc: '任何语言 / 任何 LLM 框架都能消费；只需一个文件读取器。' },
      { name: '链接即 context',                  desc: '从入口 concept 沿 [A](/path) 链遍历 = 多跳 context expansion。' },
    ],
    sections: [],
    insights: [
      { icon: "i-lucide-sparkles", title: 'OKF = "RAG 的 R 友好"', body: 'RAG 的 Retrieval 阶段不需要学"OKF 协议"——把 .md 当成普通文本 / markdown 喂给任何 chunker / embedder 即可。' +
        '这是 OKF 比专用 vector DB 更通用的关键。' },
    ],
  },

  viewer: {
    id: 'viewer',
    abstract: 'viz.html = viewer/generator.py 静态生成的单文件 HTML force-directed graph，看 concept 之间的 link 关系。',
    principles: [
      { name: '单文件 artifact',                desc: '一个 .html，无后端、无 CDN 依赖、本地双击能开。' },
      { name: '节点 = concept',                  desc: '节点色按 type 区分；label 用 frontmatter.title。' },
      { name: '边 = link',                       desc: '所有 markdown link 在 graph 中都呈现为有向边。' },
      { name: '可托管到 GCS / Pages',            desc: '静态文件 + 静态文件 = 0 运维。' },
    ],
    sections: [
      {
        kind: 'example',
        title: '真实 viz · bundles/ga4/viz.html 由 generator.py 渲染',
        code: `# 内部走法（viewer/generator.py）
# 1. 扫描 bundle 树，收集所有 .md
# 2. parse frontmatter → 节点 (id, type, title)
# 3. parse body → 提取 [text](url) 形式 link → 边
# 4. 把 nodes/edges JSON 内联到一个 <script> 标签
# 5. 用 cytoscape / d3 渲染 force-directed graph
# 输出 → 单文件 viz.html (no server, no build step)`,
      },
    ],
    insights: [
      { icon: "i-lucide-sparkles", title: 'viz = OKF 自身 OK',   body: 'viewer 走完一遍"扫描 frontmatter + parse link → graph"——本质上就是给 OKF 写了一个 consumer 实例，' +
        '证明 OKF 真的能被任何工具消费而不依赖 producer 私有 API。' },
    ],
  },
}

export interface OkfModuleArch {
  id: string
  abstract: string
  principles: NamedItem[]
  sections: OkfSection[]
  insights: Insight[]
}

export type OkfSection =
  | { kind: 'anatomy';  title: string; rows: { label: string; detail: string; accent?: any }[] }
  | { kind: 'loop';     title: string; steps: { name: string; desc: string }[] }
  | { kind: 'flow';     title: string; steps: { name: string; desc: string }[] }
  | { kind: 'tree';     title: string; tree: string }
  | { kind: 'example';  title: string; code: string }

export function getOkfModule(id: string | null): OkfModuleArch | null {
  if (!id) return null
  return OKF_MODULES[id] ?? null
}
