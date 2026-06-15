import type { StageArch } from './comm'

export const reasonArch: StageArch = {
  id: 'reason',
  abstract:
    '不论"端到端 LLM"还是"多步 Agent"，都能被还原成 Plan → Ground → Generate → Repair 四段。各家差异 = 把 LLM 放在哪几段、谁来展开 join/计算、错了怎么循环。',
  principles: [
    {
      name: '把"问题→SQL"切开',
      desc: '不要让一次 LLM 调用同时做意图分解、schema 落地、SQL 生成；切开后每段都可监控、可重试、可缓存。',
    },
    {
      name: '校验闭环胜过精度',
      desc: '"先生成 SQL 再 dry-run / 修复"循环，比"训出更聪明的生成模型"更可靠。',
    },
    {
      name: 'Schema linking 是核心难题',
      desc: '不是 LLM 能力问题，而是上下文层质量问题：列描述 / enum / 同义词不到位，再聪明的 LLM 都会绑错。',
    },
    {
      name: '让 LLM 声明、让引擎展开',
      desc: 'LLM 把意图映射到模型 / 指标；join 路径、RLAC 注入、防 fan-out 交给确定式引擎——LLM 直接写物理 SQL 是技术债。',
    },
  ],
  subQuestions: [
    /* ─────── Q1: 推理流水线骨架 ─────── */
    {
      id: 'shape',
      question: '推理流水线长什么样？',
      why: '系统的"骨架"决定了能做多复杂的查询、错了多容易修。',
      steps: [
        {
          id: 'reason-shape-1',
          name: '整体拓扑：把 LLM 放在哪几段',
          desc: '端到端单次 / 固定多步流水线 / Agent ReAct 循环 / 原语 + 外置 Agent——四种骨架决定可控性。',
          icon: 'i-lucide-workflow',
          takes: [
            {
              vendor: 'WrenAI',
              school: 'semantic-layer',
              primary: true,
              desc: '原语 + 外置 Agent：把 fetch / dry-plan / dry-run / generate 暴露为工具，Agent 编排，wren-core 引擎做确定式展开。',
              detail: {
                summary:
                  'WrenAI 的骨架是"语义引擎 + 可编排原语"——它不内置一条写死的流水线，而是把每个阶段做成独立、可被 Agent 调用的工具，真正的 join/计算展开交给 Rust 写的 wren-core 引擎。',
                bullets: [
                  {
                    label: 'Plan 在 Agent 侧',
                    icon: 'i-lucide-list-tree',
                    accent: 'violet',
                    body: 'Agent（或内置 pipeline）负责把业务问题分解、改写、路由——这一层是可替换的 LLM 编排，不绑死在引擎里。',
                  },
                  {
                    label: 'Ground 走召回原语',
                    icon: 'i-lucide-radar',
                    accent: 'amber',
                    body: 'retrieval pipeline 从 MDL 索引里 fetch 候选 model / column，缩小 LLM 的选择空间——schema linking 不靠"全塞 prompt 让 LLM 猜"。',
                  },
                  {
                    label: 'Generate 是"模型 SQL"',
                    icon: 'i-lucide-code-2',
                    accent: 'emerald',
                    body: 'LLM 只写引用 MDL model / 计算列的"逻辑 SQL"；wren-core 再把它展开成对物理表的真实 SQL（join 路径、CTE、RLAC 注入都在引擎里）。',
                  },
                  {
                    label: 'Repair 靠结构化错误',
                    icon: 'i-lucide-rotate-ccw',
                    accent: 'rose',
                    body: 'dry-plan / dry-run 失败返回带 phase + code 的结构化错误，Agent 据此精确修复，而不是吞一个字符串报错。',
                  },
                ],
                closing:
                  '一句话：WrenAI = 一个确定式语义引擎 + 一圈可被任意 Agent 编排的原语；LLM 越弱，引擎兜底越关键。',
              },
              code: [
                { repo: 'wrenai', path: 'wren-ai-service/src/pipelines', label: 'pipelines/' },
                { repo: 'wrenai', path: 'wren-core', label: 'wren-core (engine)' },
              ],
            },
            {
              vendor: 'Snowflake Cortex Analyst',
              school: 'managed-cloud',
              primary: true,
              desc: '固定多步流水线：任务路由 → 基于 Semantic View 的 schema 落地 → 受控生成 → 内置 self-correction，全部在云内托管。',
              detail: {
                summary:
                  'Cortex Analyst 是"黑盒里的固定流水线"——你看不到中间步骤的代码，但官方文档披露它内部按固定阶段走：理解问题 → 选 semantic model → 生成 SQL → 自校验，整条链路跑在 Snowflake 账户内、不出仓。',
                bullets: [
                  {
                    label: '阶段固定、不可编排',
                    icon: 'i-lucide-lock',
                    accent: 'blue',
                    body: '和 WrenAI 的"可编排原语"相反——你不能插入自定义步骤，只能通过 Semantic Model YAML + Verified Queries 影响它的行为。',
                  },
                  {
                    label: '语义层是唯一抓手',
                    icon: 'i-lucide-file-cog',
                    accent: 'emerald',
                    body: '生成质量几乎完全由 Semantic Model 的 description / synonyms / verified queries 决定——黑盒流水线本身不开放调参。',
                  },
                  {
                    label: '数据不出仓',
                    icon: 'i-lucide-shield',
                    accent: 'violet',
                    body: '整条推理链路在 Snowflake 边界内执行——这是 managed-cloud 学派最大的卖点（合规 / 安全）也是最大的锁定（不可迁移）。',
                  },
                ],
                closing:
                  '一句话：Cortex 用"封闭换稳定"——上手快、合规强，代价是你只能从语义层这一个旋钮去调它。',
              },
              refs: ['sf-cortex-overview'],
            },
            {
              vendor: 'ATLAS',
              school: 'agentic',
              desc: 'Agent ReAct 循环：Reason → Act(tool) → Observe 自治推进，步骤数自定；rewrite / linking / verify_sql 都是工具。',
              detail: {
                summary:
                  'ATLAS 走的是纯 Agent ReAct 路线——没有写死的流水线阶段，Agent 在一个"思考→调工具→看结果"的循环里自己决定下一步，适合长尾 / 复杂 / 多轮问题。',
                bullets: [
                  {
                    label: '步骤数动态',
                    icon: 'i-lucide-infinity',
                    accent: 'violet',
                    body: '简单问题一两步出 SQL；复杂问题可以多轮 rewrite → 召回 RC → verify_sql → 修复，循环直到通过。',
                  },
                  {
                    label: '工具即能力',
                    icon: 'i-lucide-wrench',
                    accent: 'amber',
                    body: 'rewrite / set_rich_context / verify_sql / dry-run 都是 Agent 工具——能力的增减 = 工具的增减，不改流水线代码。',
                  },
                  {
                    label: '代价：可预测性',
                    icon: 'i-lucide-dice-5',
                    accent: 'rose',
                    body: 'ReAct 的自由度换来的是"步骤不固定"——延迟和 token 成本比固定流水线更难预估，需要靠 verify 闸门 + 步数上限兜底。',
                  },
                ],
                closing: '一句话：ReAct 适合"问题形态不可枚举"的场景；用闸门和上限把它的自由度框住。',
              },
            },
            {
              vendor: 'ktx',
              school: 'open-context',
              desc: '原语 + BYO Agent：MCP 暴露 sl_query / wiki / search 工具，由用户自己的 Claude Code / Codex / Cursor 编排推理。',
              code: [{ repo: 'ktx', path: 'python/ktx-sl', label: 'semantic-layer planner' }],
            },
            {
              vendor: 'dbt SL',
              school: 'semantic-layer',
              desc: 'dbt SL 不做推理——只提供 MetricFlow 编译引擎；"问题→选指标"这步交给上游消费方（Hex / Tableau / 自建 Agent）。',
              notSupported:
                'dbt SL 自身没有 NL2SQL 推理流水线；它是被编排的"指标编译器"，Plan/Ground/Repair 都在消费方。',
            },
            {
              vendor: 'Databricks UC',
              school: 'managed-cloud',
              desc: 'AI/BI Genie：固定的托管流水线（理解 → 取 UC 元数据 → 生成 → 内置 retry），与 Cortex 同形态。',
              refs: ['dbx-genie'],
            },
            {
              vendor: 'Fabric DA',
              school: 'managed-cloud',
              desc: 'Fabric Data Agent：内置 Agent + 任务路由的托管流水线；绑定 Fabric / OneLake 生态。',
              refs: ['fabric-data-agent'],
            },
          ],
        },
      ],
      commonSense:
        '**固定流水线适合云厂商封闭场景；Agent 循环适合复杂 / 长尾问题；原语 + 外置 Agent 适合可定制场景**。最忌讳的是"端到端单 LLM 调用"——任何错都是黑盒，没法定位、没法局部重试。',
    },

    /* ─────── Q2: Plan · 意图分解 ─────── */
    {
      id: 'plan',
      question: '"意图分解"放不放？怎么放？',
      why: '业务问题往往含多个子查询（"上月营收 + 同比增长 + Top5 客户"），不分解就让 LLM 一次写出，容易丢条件。',
      steps: [
        {
          id: 'reason-plan-1',
          name: '问题改写 / 规范化',
          desc: '把口语化问题改写成"标准化"形式：把"上月"具体化成日期、把单位 / 时区注入。',
          icon: 'i-lucide-pencil-line',
          takes: [
            {
              vendor: 'WrenAI',
              school: 'semantic-layer',
              primary: true,
              desc: '生成前先跑 intent / question understanding pipeline：纠正措辞、补全省略、注入时间上下文。',
              code: [
                { repo: 'wrenai', path: 'wren-ai-service/src/pipelines/generation', label: 'generation/' },
              ],
            },
            {
              vendor: 'Snowflake Cortex Analyst',
              school: 'managed-cloud',
              primary: true,
              desc: '黑盒内部做问题理解 + 改写；用户侧只能通过 Semantic Model 的 synonyms / instructions 影响规范化结果。',
              refs: ['sf-cortex-overview'],
            },
            {
              vendor: 'ATLAS',
              school: 'agentic',
              desc: '独立 rewrite step：把"上月"消解为日期区间、把"客户"消歧到模型名、注入单位 / 时区。',
            },
          ],
        },
        {
          id: 'reason-plan-2',
          name: '任务分类 / 路由',
          desc: '判断是 NL2SQL / 描述查询 / 元数据问询 → 走不同 prompt / 不同子流程。',
          icon: 'i-lucide-split',
          takes: [
            {
              vendor: 'Snowflake Cortex Analyst',
              school: 'managed-cloud',
              primary: true,
              desc: '内部对问题类型分类后路由到不同处理路径（数据查询 vs 元数据 vs 闲聊拒答）。',
              refs: ['sf-cortex-overview'],
            },
            {
              vendor: 'Fabric DA',
              school: 'managed-cloud',
              desc: '同样做意图路由——区分"取数 / 解释 / 拒答"，绑定 Fabric 数据源。',
              refs: ['fabric-data-agent'],
            },
            {
              vendor: 'WrenAI',
              school: 'semantic-layer',
              desc: 'pipeline 内含 question classification——非数据问题（greeting / misleading）提前拦截，不进生成。',
              code: [
                { repo: 'wrenai', path: 'wren-ai-service/src/pipelines/generation', label: 'classification' },
              ],
            },
            {
              vendor: 'ATLAS',
              school: 'agentic',
              desc: 'Agent 在 ReAct 里隐式路由：根据问题自行决定调 NL2SQL 工具还是元数据 / 描述工具。',
              selfContained: true,
            },
          ],
        },
        {
          id: 'reason-plan-3',
          name: '多步分解 + 子查询合并',
          desc: '拆成子问题，分别生成 SQL 再 UNION / JOIN——复合问题的关键。',
          icon: 'i-lucide-git-fork',
          takes: [
            {
              vendor: 'ATLAS',
              school: 'agentic',
              desc: 'ReAct 天然支持多步：复合问题被拆成多轮工具调用，中间结果在 Agent 上下文里累积合并。',
            },
            {
              vendor: 'WrenAI',
              school: 'semantic-layer',
              desc: 'Agent 编排下可做子问题分解；wren-core 负责把多模型引用合并展开成单条 SQL。',
            },
            {
              vendor: 'Snowflake Cortex Analyst',
              school: 'managed-cloud',
              desc: '复合问题主要靠单条语义 SQL 表达（多 measure / 多 dimension）；显式"子查询分解"对用户不可见。',
              refs: ['sf-cortex-overview'],
            },
          ],
        },
      ],
      commonSense:
        '**至少要有"问题改写"——把"上月"具体化成日期、把"客户"消歧到模型名、把单位 / 时区注入**。哪怕只做这一步，也能挡住大部分的低级错误。"不分解、直接生成"只适合简单单语句。',
    },

    /* ─────── Q3: Ground · schema linking ─────── */
    {
      id: 'ground',
      question: 'Schema linking 怎么做？',
      why: '"客户" → 哪张表？哪一列？这步错了，后面全错。',
      steps: [
        {
          id: 'reason-ground-1',
          name: '候选召回：缩小 schema 范围',
          desc: '向量 / 关键词召回 top-k schema_items，把候选从"整库"缩到"十几个"，再交给 LLM 选。',
          icon: 'i-lucide-radar',
          takes: [
            {
              vendor: 'WrenAI',
              school: 'semantic-layer',
              primary: true,
              desc: 'retrieval pipeline 从 MDL 索引召回相关 model / column / relationship 作为候选，限定 LLM 选择空间。',
              code: [
                { repo: 'wrenai', path: 'wren-ai-service/src/pipelines/retrieval', label: 'retrieval/' },
              ],
            },
            {
              vendor: 'Snowflake Cortex Analyst',
              school: 'managed-cloud',
              primary: true,
              desc: 'Cortex Search service 托管召回相关表 / 列 / verified queries；大 schema 下按需混合 BM25。',
              refs: ['sf-cortex-search'],
            },
            {
              vendor: 'ATLAS',
              school: 'agentic',
              desc: 'linking pipeline 并发召回多类 Rich Context（同义词 / enum / 关系），MariaDB 向量 + FULLTEXT 双路。',
            },
            {
              vendor: 'ktx',
              school: 'open-context',
              desc: 'ktx-daemon 做 BM25 + 向量 hybrid 召回，命中 semantic-layer + wiki 后经 MCP 喂给 Agent。',
              code: [{ repo: 'ktx', path: 'python/ktx-daemon', label: 'ktx-daemon' }],
            },
          ],
        },
        {
          id: 'reason-ground-2',
          name: '取值剖析：消歧到具体列',
          desc: '看列里实际 distinct 值（"status 里有 paid/cancelled/refunded"）判断哪列才是用户说的那个。',
          icon: 'i-lucide-scan-search',
          takes: [
            {
              vendor: 'WrenAI',
              school: 'semantic-layer',
              primary: true,
              desc: 'value profiling：采样列的 distinct 值参与 schema linking，解决"按状态过滤"绑错列的问题。',
            },
            {
              vendor: 'ATLAS',
              school: 'agentic',
              desc: 'introspect 阶段就剖析列取值并写进 RC enum；linking 时直接用，无需现场采样。',
            },
            {
              vendor: 'Snowflake Cortex Analyst',
              school: 'managed-cloud',
              desc: 'Semantic Model 里可声明列的 sample values；生成时参考它消歧。',
              refs: ['sf-semantic-yaml'],
            },
          ],
        },
        {
          id: 'reason-ground-3',
          name: '强约束：strict mode 兜底',
          desc: '禁止裸物理表 / 未建模列——schema linking 失败时宁可报错也不让 LLM 编一个不存在的列。',
          icon: 'i-lucide-shield-alert',
          takes: [
            {
              vendor: 'WrenAI',
              school: 'semantic-layer',
              primary: true,
              desc: 'wren-core strict mode：只允许引用 MDL 已建模的 model / column；未建模引用直接拒绝。',
              code: [{ repo: 'wrenai', path: 'wren-core', label: 'wren-core' }],
            },
            {
              vendor: 'Cube',
              school: 'semantic-layer',
              desc: 'Cube 只暴露已定义的 measures / dimensions，物理列对查询不可见——天然 strict。',
              code: [{ repo: 'cube', path: 'packages/cubejs-schema-compiler', label: 'schema-compiler' }],
            },
            {
              vendor: 'ATLAS',
              school: 'agentic',
              desc: 'verify_sql 校验引用列是否存在于 catalog；不存在则结构化报错让 Agent 重绑。',
            },
            {
              vendor: 'Snowflake Cortex Analyst',
              school: 'managed-cloud',
              desc: '生成被约束在 Semantic Model 暴露的 tables / dimensions 范围内，越界即拒。',
              refs: ['sf-semantic-views'],
            },
          ],
        },
      ],
      commonSense:
        '**召回缩小候选 + 取值剖析消歧 + strict mode 兜底**——三层组合是 schema linking 在生产里能跑稳的最低配置。少任何一层，长尾错误率都会上升。',
    },

    /* ─────── Q4: Generate · SQL 生成主体 ─────── */
    {
      id: 'gen',
      question: 'SQL 生成的"主体"在哪？',
      why: '把 join、计算列、关系展开放在 LLM 还是引擎，决定了正确性的上限。',
      steps: [
        {
          id: 'reason-gen-1',
          name: 'join / 计算的展开权在谁手里',
          desc: 'LLM 直接写物理 SQL / LLM 写"模型 SQL"引擎展开 / LLM 只选 measures 引擎全权生成 / Agent 工具拼装——四档。',
          icon: 'i-lucide-code-2',
          takes: [
            {
              vendor: 'WrenAI',
              school: 'semantic-layer',
              primary: true,
              desc: 'LLM 写引用 MDL model / 计算列的"逻辑 SQL"，wren-core 引擎展开成物理 SQL（join 路径 / CTE / RLAC）。',
              detail: {
                summary:
                  'WrenAI 把生成职责一分为二：LLM 只负责"写一段引用逻辑 model 的 SQL"，真正难且容易错的部分——join 怎么连、计算列怎么算、权限怎么注入——全部由 Rust 引擎 wren-core 确定式展开。',
                bullets: [
                  {
                    label: 'LLM 写"模型 SQL"',
                    icon: 'i-lucide-bot',
                    accent: 'violet',
                    body: 'LLM 输出形如 `SELECT customer.name, SUM(orders.amount) FROM orders` 的逻辑 SQL——引用的是 MDL model 名和计算列，不是物理表。',
                  },
                  {
                    label: '引擎展开 join',
                    icon: 'i-lucide-git-merge',
                    accent: 'emerald',
                    body: 'wren-core 按 MDL relationship 自动补全 JOIN 路径——LLM 不需要知道 orders.customer_id = customers.id 这种物理细节。',
                  },
                  {
                    label: '引擎注入治理',
                    icon: 'i-lucide-shield-check',
                    accent: 'amber',
                    body: 'RLAC / CLAC / 计算列表达式都在引擎展开时注入——LLM 无法绕过权限，因为它根本没碰物理表。',
                  },
                  {
                    label: 'fan-out 由引擎防',
                    icon: 'i-lucide-alert-triangle',
                    accent: 'rose',
                    body: '1-N 关系下 SUM 重复计数的坑由 planner 按 grain 拆 CTE 处理——不指望 LLM 记得每个 measure 的粒度。',
                  },
                ],
                closing:
                  '一句话：LLM 负责"理解业务、声明意图"，引擎负责"正确性"——这是 WrenAI 把准确率天花板拉高的核心设计。',
              },
              example: {
                lang: 'sql',
                caption: 'LLM 写的"模型 SQL"（逻辑） → wren-core 展开为物理 SQL',
                code: `-- ① LLM 输出（引用 MDL model + 关系列，不含物理 join）
SELECT
  c.name,
  SUM(o.amount) AS revenue
FROM orders o
JOIN o.customer c            -- 关系列：写 join 句柄，不写 ON 条件
WHERE o.ordered_at >= '2026-01-01'
GROUP BY c.name;

-- ② wren-core 展开后的物理 SQL（节选，引擎自动补全）
SELECT c.name, SUM(o.amount) AS revenue
FROM jaffle.main.orders o
JOIN jaffle.main.customers c
  ON o.customer_id = c.usr_id        -- ← 引擎按 MDL relationship 补
WHERE o.ordered_at >= '2026-01-01'
  AND c.region = CURRENT_USER_REGION() -- ← 引擎按 RLAC 注入
GROUP BY c.name;`,
              },
              code: [{ repo: 'wrenai', path: 'wren-core', label: 'wren-core (展开引擎)' }],
            },
            {
              vendor: 'Snowflake Cortex Analyst',
              school: 'managed-cloud',
              primary: true,
              desc: 'LLM 选 measures / dimensions，由 Semantic View 引擎全权生成 SQL——用户拿到的是引擎产出的可信 SQL。',
              detail: {
                summary:
                  'Cortex Analyst 把"写 SQL"这件事几乎完全从 LLM 手里拿走——LLM 的产出更接近"我要 revenue 这个 measure，按 month 这个 dimension 分组"，真正的 SQL 由 Semantic View 引擎生成。',
                bullets: [
                  {
                    label: 'LLM 做声明',
                    icon: 'i-lucide-list-checks',
                    accent: 'blue',
                    body: 'LLM 把问题映射到 Semantic Model 里已定义的 measures / dimensions / filters——而不是逐字写 join 和聚合。',
                  },
                  {
                    label: '引擎生成可信 SQL',
                    icon: 'i-lucide-database',
                    accent: 'emerald',
                    body: 'measure 的聚合表达式、join 关系在 Semantic View 里预先定义好，引擎据此生成——同一个 measure 永远算法一致。',
                  },
                  {
                    label: 'verified queries 加持',
                    icon: 'i-lucide-badge-check',
                    accent: 'violet',
                    body: '命中 Verified Query Repository 时直接复用审过的 SQL，连生成都省了——准确率最高的一条路径。',
                  },
                ],
                closing:
                  '一句话：和 WrenAI 同philosophy（声明 vs 展开），区别只是 Cortex 把引擎 + 数据都关在 Snowflake 里。',
              },
              refs: ['sf-semantic-views', 'sf-vqr'],
            },
            {
              vendor: 'ATLAS',
              school: 'agentic',
              desc: 'Agent 工具拼装：用 verify_sql / set_rich_context 等工具增量推进，LLM 写物理 SQL 但每步都过校验闸门。',
              detail: {
                summary:
                  'ATLAS 当前更接近"LLM 写物理 SQL + 强校验兜底"——没有 wren-core 那样的逻辑→物理展开引擎，而是靠 ReAct 里的 verify_sql 工具把每次生成都过一遍闸门。',
                bullets: [
                  {
                    label: 'LLM 写物理 SQL',
                    icon: 'i-lucide-bot',
                    accent: 'violet',
                    body: 'Agent 直接生成针对物理表的 SQL——join / 聚合都由 LLM 写，借助召回到的 Rich Context（关系 / enum）来写对。',
                  },
                  {
                    label: 'verify_sql 闸门',
                    icon: 'i-lucide-shield-check',
                    accent: 'amber',
                    body: '生成后立刻 verify_sql：语法 / 列存在 / 只读 / dry-run；失败结构化报错，Agent 在循环里修。',
                  },
                  {
                    label: '取舍',
                    icon: 'i-lucide-scale',
                    accent: 'rose',
                    body: '没有引擎展开层 = 少一层抽象、上手直接；代价是 join/fan-out 正确性更依赖召回质量 + 校验闸门，而非引擎硬保证。',
                  },
                ],
                closing: '一句话：ATLAS 用"强校验闭环"替代"展开引擎"——闸门做得越严，物理 SQL 路线越稳。',
              },
            },
            {
              vendor: 'dbt SL',
              school: 'semantic-layer',
              desc: 'LLM 选 metrics，MetricFlow 编译成 SQL——和 Cortex 同档（声明 + 引擎展开），但引擎是开源的 MetricFlow。',
              code: [{ repo: 'metricflow', path: 'metricflow/engine', label: 'metricflow/engine' }],
            },
            {
              vendor: 'ktx',
              school: 'open-context',
              desc: 'sl_query 引擎：Agent 声明要的指标 / 维度，ktx-sl planner 生成 SQL——同样是"声明 + 展开"。',
              code: [{ repo: 'ktx', path: 'python/ktx-sl', label: 'sl_query planner' }],
            },
            {
              vendor: 'Databricks UC',
              school: 'managed-cloud',
              desc: 'Genie 基于 Metric View 生成；measure 的聚合在 Metric View 定义里，引擎据此产出。',
              refs: ['dbx-genie', 'dbx-mv-ref'],
            },
          ],
        },
      ],
      commonSense:
        '**让 LLM 做"声明"，让引擎做"展开"**——LLM 的价值是理解业务问题、把意图映射到模型；写 join 路径、注入 RLAC、防 fan-out 这些应该交给确定式引擎。LLM 直接写物理 SQL 是技术债（除非配强校验闭环兜底）。',
    },

    /* ─────── Q5: Repair · 错误循环 ─────── */
    {
      id: 'loop',
      question: '错了之后怎么循环？',
      why: '"一次写对"是奢望；如何把错误变成结构化反馈让系统自己修复，决定了稳定性。',
      steps: [
        {
          id: 'reason-loop-1',
          name: '错误反馈 + 自动修复',
          desc: '返回结构化错误（phase / code / 列名 / 期望类型）→ Agent / 内置链据此精确修复。',
          icon: 'i-lucide-rotate-ccw',
          takes: [
            {
              vendor: 'WrenAI',
              school: 'semantic-layer',
              primary: true,
              desc: 'WrenError(phase, code)：dry-plan / dry-run 失败返回带阶段 + 错误码的结构化错误，Agent 据 phase 定向修复。',
              code: [
                { repo: 'wrenai', path: 'wren-ai-service/src/pipelines/generation', label: 'sql_correction' },
              ],
            },
            {
              vendor: 'Snowflake Cortex Analyst',
              school: 'managed-cloud',
              primary: true,
              desc: '内置 self-correction：黑盒流水线在返回前自校验 + 重试若干轮，用户看不到中间过程。',
              refs: ['sf-cortex-overview'],
            },
            {
              vendor: 'ATLAS',
              school: 'agentic',
              desc: 'verify_sql 失败 → Agent 在 ReAct 里自然修：根据结构化报错调整列绑定 / join / 过滤，再次校验。',
            },
            {
              vendor: 'dbt SL',
              school: 'semantic-layer',
              desc: 'dbt SL 自身不做生成-修复循环；MetricFlow 编译报错由消费方处理。',
              notSupported: 'dbt SL 是编译器不是 Agent——修复循环在上游消费方，不在 SL 内。',
            },
          ],
        },
      ],
      commonSense:
        '**结构化错误是基础设施**——返回 phase / code / 列名 / 期望类型，Agent 才有办法定位。返回字符串错误信息（"syntax error near LIMIT"）等于没给反馈。"不循环、错就让用户复述"是基线，生产不可接受。',
    },
  ],
  insights: [
    {
      icon: 'i-lucide-cpu',
      title: '语义引擎是真相源',
      body: '"LLM + 数据" 不等于 "上下文层"——决定正确性的是"谁来执行 join / 计算列 / 权限注入"。把这件事放在确定式引擎（wren-core / SL planner）里，不论 LLM 多差都不会出错。',
    },
    {
      icon: 'i-lucide-route',
      title: '可解释性 ≈ 生成轨迹',
      body: 'dry-plan / 计划解释把 LLM 的"黑盒选择"展开成"用了哪些模型 / join / 计算列"——这就是 NL2SQL 系统该有的可解释性，比"看 LLM 在想什么"靠谱得多。',
    },
    {
      icon: 'i-lucide-sliders',
      title: 'Schema linking 准 ≠ LLM 强',
      body: 'schema linking 失败的根因 90% 是上下文层缺信息（列描述 / enum / 同义词），10% 才是 LLM 不懂。把精力放在 wiki 而非 prompt 工程上。',
    },
  ],
  matrix: {
    cols: ['Plan', 'Ground', 'Generate', 'Repair'],
    rows: [
      { vendor: 'ATLAS', school: 'agentic', cells: ['rewrite', 'agent ReAct', 'agent + verify_sql', 'tool 内修'] },
      { vendor: 'WrenAI', school: 'semantic-layer', cells: ['agent 自分解', 'fetch + recall', 'agent → wren-core 展开', 'WrenError + retry'] },
      { vendor: 'ktx', school: 'open-context', cells: ['agent skill', 'hybrid 召回', 'sl_query 引擎', '工具粒度修复'] },
      { vendor: 'dbt SL', school: 'semantic-layer', cells: ['由消费方', '由消费方', 'MetricFlow 编译', '由消费方'] },
      { vendor: 'Cortex Analyst', school: 'managed-cloud', cells: ['任务路由', '云内 schema', '基于 Semantic Views', 'self-correction'] },
      { vendor: 'Fabric DA', school: 'managed-cloud', cells: ['任务路由', 'Business Sem.', '内置生成', '内置 retry'] },
    ],
  },
}
