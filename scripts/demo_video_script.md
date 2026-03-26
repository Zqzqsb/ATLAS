# ATLAS Demo Video 完整脚本

> **目标时长**: 5–7 分钟（第一版录长一点，后期剪）  
> **格式**: 屏幕录制 + AI 英文配音替换  
> **分辨率**: 1920×1080，Chrome 全屏  
> **语言**: 录制时念中文，后期 AI 替换为英文配音

---

## 使用方法

- **【操作】** = 你在屏幕上做什么（录制时看这个）
- **【中文】** = 你录制时念的中文旁白
- **【EN】** = 后期 AI 英文配音稿
- 每步操作之间**多留 2-3 秒空白**，方便剪辑
- LLM 等待时间较长的地方，后期加速处理

---

## 第一部分：四项创新介绍（Features Page）

> 在 Features 页面（/features）逐屏滚动讲解，每屏停留 10-15 秒让动画播完

### 1.0 片头 — Hero

**【操作】** 打开浏览器，进入 Features 页面，停留在首屏 Hero。四个彩色 pill 按钮可见。

**【中文】** ATLAS 是一个基于 MariaDB 原生向量能力的端到端 Text-to-SQL 系统。它的核心设计围绕四项创新——针对传统方案依赖外挂向量库、手工维护元数据、大规模 schema linking 效率低这三大痛点，ATLAS 全部在一个数据库内统一解决。接下来我先逐一介绍这四项创新，然后通过四个 live demo 场景做实际演示。

**【EN】** ATLAS is an end-to-end Text-to-SQL system built on MariaDB's native vector capabilities. Its design centers on four innovations — addressing three pain points of traditional systems: reliance on external vector stores, manually maintained metadata, and poor schema linking efficiency at scale. ATLAS solves all of these within a single database. I'll first walk through each innovation, then demonstrate them in four live scenarios.

---

### 1.1 Lakebase Storage — 湖基统一存储

**【操作】** 滚动到 Lakebase 页。等待动画展示五层统一存储 vs 传统分散架构。

**【中文】** 第一项创新——湖基统一存储。传统方案需要多个独立组件：关系型数据库存 schema、外部向量库存 embeddings、文件系统存文档、再加胶水代码做同步。ATLAS 把五层数据——Schema Metadata、Rich Context、Relationship Graph、Vector Embeddings、Change Audit Log——全部存在同一个 MariaDB 实例里。利用原生 VECTOR 列和 HNSW 索引，一条 SQL 就能同时做向量检索和关系型过滤。所有数据在 ACID 事务保护下，不会出现不一致。

**【EN】** The first innovation — unified in-database storage. Traditional systems need multiple components: a relational database for schema, an external vector store for embeddings, a file system for docs, plus glue code. ATLAS consolidates five data layers — Schema Metadata, Rich Context, Relationship Graph, Vector Embeddings, and Change Audit Log — within a single MariaDB instance. With native VECTOR columns and HNSW index, one SQL query handles both vector retrieval and relational filtering. All under ACID guarantees — no consistency issues.

---

### 1.2 Two-Stage Adaptive Schema Linking — 两阶段自适应 Linking

**【操作】** 滚动到 Schema Linking 页。等待动画展示 30 个表 → 向量筛选 → LLM 精选。

**【中文】** 第二项创新——两阶段自适应 Schema Linking。小库（30 张表以内）直接 one-shot 全量发给 LLM。大库（比如 517 张表）塞不进上下文窗口，系统自动切换两阶段：第一阶段向量检索，1 秒内把数百张表缩减到约 20 个候选；第二阶段 LLM 精确推理，选出真正需要的表。粗筛加精选，既不超窗口，又保证召回。

**【EN】** The second innovation — two-stage adaptive schema linking. Small databases go directly to the LLM as one-shot. Large databases — like 517 tables — won't fit the context window, so the system automatically switches to two stages: vector retrieval narrows hundreds of tables to about 20 candidates in one second; then the LLM performs precise reasoning to select the truly relevant ones. Coarse filtering plus fine selection — within the window, with high recall.

---

### 1.3 Rich Context Lifecycle — Rich Context 生命周期

**【操作】** 滚动到 Context Lifecycle 页。等待动画依次展示 Onboarding → Query-Time → Evolution 三个阶段。

**【中文】** 第三项创新——Rich Context 生命周期。Rich Context 不是一次性静态标注，它有三个阶段。Onboarding：接入新库时，ReAct Agent 自动采样数据、分析分布，为每张表每列生成语义描述、同义词、业务规则，嵌入 HNSW 索引。Query-Time：用户提问时，向量检索召回相关 context 注入 LLM prompt，帮模型理解业务语义。Evolution：schema 变更时自动检测、标记过时 context、LLM 重新生成、更新向量。三阶段闭环——生成、消费、刷新，全自动。

**【EN】** The third innovation — Rich Context lifecycle. Rich Context is not static — it has three phases. Onboarding: a ReAct Agent samples data and generates descriptions, synonyms, and rules for every column, embedded into HNSW. Query-time: vector retrieval injects relevant context into the LLM prompt. Evolution: on DDL changes, stale context is detected, regenerated, and re-embedded. Three phases in a closed loop — generate, consume, refresh — fully automatic.

---

### 1.4 Agent-Driven Self-Maintenance — Agent 自维持

**【操作】** 滚动到 Agent Self-Maintain 页。等待动画展示 DDL 变更 → 检测 → 标记 → 刷新 → 重嵌入五步。

**【中文】** 第四项创新——Agent 驱动的自维持。生产环境 schema 频繁变更：加列、建表、改外键。传统系统靠人工重新标注，几百张表根本不可行。ATLAS 用 Coordinator-Executor 两阶段 Agent 自动维护：Coordinator 分析 DDL diff 规划任务，Executor 执行——生成描述、更新关联 context、重新嵌入向量。全程记录 Change Log，支持 before/after 审计。Schema 怎么变，Rich Context 跟着变。

**【EN】** The fourth innovation — agent-driven self-maintenance. Production schemas change frequently. Manual re-annotation is infeasible at scale. ATLAS uses a Coordinator-Executor agent: the Coordinator plans tasks from the DDL diff; the Executor generates descriptions, updates related context, and re-embeds vectors. Everything logged in the Change Log with before/after auditing. Schema evolves, context follows.

---

### 1.5 过渡

**【操作】** 滚动到 CTA 页。点击 "Go to Databases" 跳转到首页。

**【中文】** 以上是四项核心创新。接下来我们切到实际数据库，做四个 live demo。

**【EN】** Those are the four core innovations. Now let's switch to real databases for four live demonstrations.

---

## 第二部分：四个 Live Demo 场景

### 场景一：Onboarding — Rich Context 生成

> 使用数据库: spider_tvshow (3 表)  
> 前置条件: Rich Context 已清空  
> 对应创新: #3 Rich Context Lifecycle（Onboarding 阶段）

**【操作】** 在首页点击 Spider Dataset 卡片，进入 workspace。  
**【中文】** 第一个场景——用一个 3 张表的 TV 节目数据库演示 Onboarding。  
**【EN】** Scenario 1 — we use a 3-table TV show database to demonstrate Onboarding.

**【操作】** 点击 Sync Schema，等待完成。  
**【中文】** 先同步 Schema 到 Lakebase。  
**【EN】** First, sync the schema into Lakebase.

**【操作】** 切换 Context Tab，展示空状态。  
**【中文】** 当前没有任何 Rich Context，点击生成。  
**【EN】** No Rich Context yet. Let's generate.

**【操作】** 点击 Generate Rich Context → Console 弹出 → Start Generation。  
**【中文】** ReAct Agent 启动，逐表分析。  
**【EN】** The ReAct Agent launches and analyzes each table.

**【操作】** 让日志滚动几秒，展示 thought → action → observation 循环。  
**【中文】** Agent 自动采样数据、分析分布、发现语义关系。比如它识别出 rating 是 1-10 的观众评分。  
**【EN】** The agent samples data, analyzes distributions, and discovers relationships. For instance, it identifies rating as audience scores from 1 to 10.

**【操作】** 完成后关闭 Console。展开 Context Manager，悬停几个条目看标签。  
**【中文】** 三张表的 Rich Context 全部生成。Context Manager 按表和列分组，每个条目有 description、synonym、value mapping 等类型标签。这就是 Onboarding——一键从零到完整语义知识库。  
**【EN】** Rich Context for all three tables is generated. The Context Manager groups entries by table and column, each with type labels. That's Onboarding — one click from zero to a complete semantic knowledge base.

---

### 场景二：Rich Context 消歧 — 前后对比

> 使用数据库: spider_tvshow (3 表)  
> 前置条件: 场景一刚完成，RC 已生成  
> 对应创新: #3 Rich Context Lifecycle（Query-Time 阶段）

#### Part A — 关闭 Rich Context（预期答错）

**【操作】** 切换 Query Tab。将 Linking Mode 切换到 Schema Only。  
**【中文】** 现在关闭 Rich Context，用裸 Schema 问一个有歧义的问题。  
**【EN】** Now we disable Rich Context and ask an ambiguous question with raw schema only.

**【操作】** 输入 `Which channel has the highest share?`，回车。  
**【中文】** share 这个词可以是收视份额、股份、市占率。裸 Schema 里它只是一个 DECIMAL 列，没有注释，LLM 无从判断。  
**【EN】** The word "share" could mean audience share, stock, or market share. In the raw schema it's just a DECIMAL column with no comments — the LLM has no way to tell.

**【操作】** 等待结果，停留几秒让观众看到错误。  
**【中文】** 模型选错了字段，因为它不理解 Share 在电视行业的含义。  
**【EN】** The model picks the wrong column — it can't understand what Share means in broadcasting.

#### Part B — 开启 Rich Context（预期答对）

**【操作】** 点击 Clear，将 Linking Mode 切换回 Rich Context。输入同一问题，回车。  
**【中文】** 现在开启 Rich Context，问完全相同的问题。  
**【EN】** Now we enable Rich Context and ask the exact same question.

**【操作】** 等待 Schema Linking 卡片亮起，稍作停留看推理内容。  
**【中文】** Rich Context 标注了 TV_series.Share 是"收看该节目的家庭占所有电视家庭的百分比"。有了这条标注，模型选对了字段，正确生成了按 Channel 分组取最大值的查询。  
**【EN】** Rich Context annotates TV_series.Share as "the percentage of TV households tuned to this program." With this, the model selects the right column and generates the correct aggregation.

**【操作】** 等待执行结果。停顿 2 秒。  
**【中文】** 同一个问题，加 Rich Context 后完全正确。这就是 Context 的核心价值——靠知识库消歧，不靠改措辞。  
**【EN】** Same question — with Rich Context, completely correct. That's the core value: disambiguation through knowledge, not prompt rewording.

---

### 场景三：Schema Evolution — Agent 自维持

> 使用数据库: lucid_evolution (2 表初始 → 3 阶段 DDL 演进)  
> 前置条件: 已 Reset 到 Stage 0  
> 对应创新: #4 Agent-Driven Self-Maintenance

**【操作】** 返回首页，点击 Evolution Demo 卡片。  
**【中文】** 第三个场景——Agent 自维持。这个库初始只有 users 和 orders 两张表。  
**【EN】** Scenario 3 — agent self-maintenance. This database starts with just users and orders.

**【操作】** 切换 Evolution Tab，展示 Stage 进度条。  
**【中文】** 我们执行三个阶段的 DDL 变更，看 Agent 如何自动维护。  
**【EN】** We'll execute three DDL stages and watch the agent maintain everything automatically.

**【操作】** Execute Next Stage（Stage 1: Add phone column）。让日志滚动。  
**【中文】** Stage 1——给 users 加 phone 列。系统检测到 column_added，同步 schema，触发自维护管线。Coordinator 分配任务，Executor 用 LLM 生成语义描述。  
**【EN】** Stage 1 — adding a phone column. The system detects column_added, syncs schema, and triggers self-maintenance. The Coordinator dispatches tasks; the Executor generates descriptions via LLM.

**【操作】** Stage 1 完成。Execute Next Stage（Stage 2: Create products table）。  
**【中文】** Stage 2——创建 products 新表。变更更大，整张表的 Rich Context 从零生成。  
**【EN】** Stage 2 — creating a products table. A larger change — full context generated from scratch.

**【操作】** 等待完成。Execute Next Stage（Stage 3: Add FK）。  
**【中文】** Stage 3——给 orders 加 product_id 外键。这会触发关系图更新。  
**【EN】** Stage 3 — adding a foreign key. This triggers relationship graph updates.

**【操作】** 等待完成。展示 Change Log 面板。  
**【中文】** 系统检测到外键关系，自动更新关联表 context。每步变更记录在 Change Log 中，支持 before/after 审计。全自动，无人工干预。  
**【EN】** The system detects the FK relationship and updates related context. Every change is logged with before/after diffs. Fully automatic, no manual intervention.

---

### 场景四：大规模 Two-Stage Linking

> 使用数据库: tpch_enterprise (517 表，21 个业务域)  
> 前置条件: Rich Context 已预生成  
> 对应创新: #1 Lakebase + #2 Two-Stage Linking

#### Part A — Forest-Based Onboarding（快速展示）

**【操作】** 返回首页，点击 TPC-H Enterprise 卡片。  
**【中文】** 最后一个场景——517 张表的企业级数据库，21 个业务域。  
**【EN】** Final scenario — a 517-table enterprise database, 21 business domains.

**【操作】** 点击 Generate → Console 弹出，显示 Forest 模式 banner。  
**【中文】** 系统检测到表数超阈值，自动启用 Forest-Based Chunked Onboarding，把 FK 图分解成约 40 个子树。  
**【EN】** The system detects the table count exceeds the threshold and enables Forest-Based Chunked Onboarding, decomposing the FK graph into about 40 subtrees.

**【操作】** 展示 Treemap 预览。停留 3-5 秒。  
**【中文】** Treemap 可视化展示分解结果，每个色块是一个 FK 连通子图。  
**【EN】** The treemap shows the decomposition — each block is an FK-connected subgraph.

#### Part B — Two-Stage Linking 查询

**【操作】** 关闭 Console，切到 Query Tab。输入 `Find the supplier with the highest profit`，回车。  
**【中文】** 问一个跨表关联的供应链问题，看两阶段 Linking 如何处理。  
**【EN】** A cross-table supply chain question — let's see two-stage linking in action.

**【操作】** 等待卡片 1：Vector Search，约 20 个候选。  
**【中文】** Stage 1——向量检索，1 秒内从 517 张表缩减到约 20 个候选。  
**【EN】** Stage 1 — vector retrieval narrows 517 tables to about 20 candidates in one second.

**【操作】** 等待卡片 2：One-Shot Schema Linking。  
**【中文】** Stage 2——LLM 对候选精确推理，选出真正相关的表。粗筛加精选，两阶段配合。  
**【EN】** Stage 2 — the LLM reasons over candidates and selects the relevant tables. Coarse filtering plus fine selection.

**【操作】** 等待 SQL 生成和执行结果。  
**【中文】** SQL 正确生成执行。517 张表到精准查询，全程实时透明。  
**【EN】** SQL generated and executed correctly. From 517 tables to a precise query, fully transparent.

---

## 第三部分：收尾

**【操作】** 返回首页全景。停留 3 秒。

**【中文】** ATLAS 把 Schema、Rich Context 和向量索引统一存储在一个 MariaDB 中，实现从 Onboarding、查询推理到 Schema 演进的完整闭环。端到端，无外挂，全自动。在 BIRD 基准上达到 75.55% 执行准确率，在 500 张表以上的企业级 schema 上高效完成端到端查询。感谢观看。

**【EN】** ATLAS stores schema, Rich Context, and vector indexes within a single MariaDB instance — a complete closed loop from onboarding, to query inference, to schema evolution. End-to-end, no external dependencies, fully automatic. It achieves 75.55% execution accuracy on BIRD, and handles enterprise schemas with over 500 tables efficiently. Thank you for watching.

---

## 纯英文配音稿（TTS 直接使用）

以下为完整英文配音文稿，可直接复制给 TTS 工具。每个 `[pause]` 处插入 1.5 秒静音。

```
ATLAS is an end-to-end Text-to-SQL system built on MariaDB's native vector capabilities. Its design centers on four innovations — addressing three pain points of traditional systems: reliance on external vector stores, manually maintained metadata, and poor schema linking efficiency at scale. ATLAS solves all of these within a single database. I'll first walk through each innovation, then demonstrate them in four live scenarios.

[pause]

The first innovation — unified in-database storage. Traditional systems need multiple components: a relational database for schema, an external vector store for embeddings, a file system for docs, plus glue code. ATLAS consolidates five data layers — Schema Metadata, Rich Context, Relationship Graph, Vector Embeddings, and Change Audit Log — within a single MariaDB instance. With native VECTOR columns and HNSW index, one SQL query handles both vector retrieval and relational filtering. All under ACID guarantees — no consistency issues.

[pause]

The second innovation — two-stage adaptive schema linking. Small databases go directly to the LLM as one-shot. Large databases — like 517 tables — won't fit the context window, so the system automatically switches to two stages: vector retrieval narrows hundreds of tables to about twenty candidates in one second; then the LLM performs precise reasoning to select the truly relevant ones. Coarse filtering plus fine selection — within the window, with high recall.

[pause]

The third innovation — Rich Context lifecycle. Rich Context is not static — it has three phases. Onboarding: a ReAct Agent samples data and generates descriptions, synonyms, and rules for every column, embedded into HNSW. Query-time: vector retrieval injects relevant context into the LLM prompt. Evolution: on DDL changes, stale context is detected, regenerated, and re-embedded. Three phases in a closed loop — generate, consume, refresh — fully automatic.

[pause]

The fourth innovation — agent-driven self-maintenance. Production schemas change frequently. ATLAS uses a Coordinator-Executor agent: the Coordinator plans tasks from the DDL diff; the Executor generates descriptions, updates related context, and re-embeds vectors. Everything logged in the Change Log with before-and-after auditing. Schema evolves, context follows.

[pause]

Those are the four core innovations. Now let's switch to real databases for four live demonstrations.

[pause]

Scenario one — we use a three-table TV show database to demonstrate Onboarding. First, sync the schema into Lakebase. No Rich Context yet — let's generate. The ReAct Agent launches and analyzes each table. It samples data, analyzes distributions, and discovers relationships. Within seconds, Rich Context for all three tables is generated. The Context Manager groups entries by table and column, each with type labels. That's Onboarding — one click from zero to a complete semantic knowledge base.

[pause]

Now we disable Rich Context and ask an ambiguous question with raw schema only. "Which channel has the highest share?" The word share could mean audience share, stock, or market share. In the raw schema it's just a DECIMAL column with no comments. The model picks the wrong column.

[pause]

Now we enable Rich Context and ask the exact same question. Rich Context annotates TV_series.Share as "the percentage of TV households tuned to this program." The model selects the right column and generates the correct aggregation. Same question — with Rich Context, completely correct. Disambiguation through knowledge, not prompt rewording.

[pause]

Scenario three — agent self-maintenance. This database starts with just users and orders. Stage one — adding a phone column. The system detects the change, syncs schema, and the Coordinator-Executor pipeline generates descriptions automatically. Stage two — creating a products table, with full context generated from scratch. Stage three — adding a foreign key, triggering relationship graph updates. Every change logged with before-and-after diffs. Fully automatic.

[pause]

Final scenario — a 517-table enterprise database, 21 business domains. The system enables Forest-Based Chunked Onboarding, decomposing the FK graph into about 40 subtrees. The treemap visualizes the decomposition. Now a cross-table query: "Find the supplier with the highest profit." Stage one — vector retrieval narrows 517 tables to about 20 candidates in one second. Stage two — the LLM selects the relevant tables. SQL generated and executed correctly. From 517 tables to a precise query, fully transparent.

[pause]

ATLAS stores schema, Rich Context, and vector indexes within a single MariaDB instance — a complete closed loop from onboarding, to query inference, to schema evolution. End-to-end, no external dependencies, fully automatic. It achieves 75.55 percent execution accuracy on BIRD, and handles enterprise schemas with over 500 tables efficiently. Thank you for watching.
```

---

## 录制前 Checklist

- [ ] Docker 环境正常运行（`docker compose -f deploy/docker-compose.yml up -d`）
- [ ] 确认所有数据库都已连接（首页卡片可见）
- [ ] spider_tvshow 的 Rich Context **已清空**（场景一从零生成）
- [ ] lucid_evolution **已 Reset**（Stage 0 起始状态）
- [ ] tpch_enterprise 的 Rich Context **已预生成**（场景四查询需要）
- [ ] LLM API Key 配置正常，网络稳定
- [ ] Chrome 全屏，1920×1080，缩放 100%
- [ ] 隐藏书签栏、插件图标、清除地址栏历史建议
- [ ] 录屏软件已开启，录制区域确认

## 录制技巧

- 鼠标移动要慢而稳，别抖
- 每个操作后**留 2-3 秒停顿**，给剪辑留余量
- 关键 UI 变化（卡片亮起、日志滚动）时暂停几秒让观众看清
- Features 页面每屏**等动画播完再翻**，别太快
- Evolution 的 Stage 等待时间较长（10-20秒），**不用着急**，后期加速
- **念错了不要停**，继续念完当前段落，后期用 AI 英文配音替换

## AI 配音注意事项

- 语速：中等偏慢，约每分钟 140-160 词（英文）
- 语气：专业但不冷淡，像在会议上给同行做 live demo
- 每个 `[pause]` 处插入 1.5 秒静音
- 每句配音之间自然停顿 0.5-0.8 秒
- 专有名词发音：ATLAS（/ˈætləs/）、ReAct（/riːˈækt/）、HNSW（逐字母 H-N-S-W）、Lakebase（/leɪkbeɪs/）
