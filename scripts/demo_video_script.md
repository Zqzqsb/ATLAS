# ATLAS Demo Video 脚本

> **目标时长**: 最终版 5–6 分钟（第一版录长一点，后期剪）  
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

## 片头

**【操作】** 打开浏览器，展示系统首页 Landing Page，能看到数据库卡片网格。

**【中文】** ATLAS 是一个基于 MariaDB 原生向量能力的端到端 Text-to-SQL 系统。它用一张数据库统一存储 Schema、Rich Context 和向量索引，无需任何外部引擎。三个 Docker 容器，一条命令即可部署全栈。接下来，我们通过三个场景，展示系统的核心能力。

**【EN】** ATLAS is an end-to-end Text-to-SQL system built on MariaDB's native vector capabilities. It stores schema, Rich Context, and vector indexes within a single database — no external engines required. Three Docker containers, one command to deploy the full stack. We will walk through three scenarios to demonstrate the system's core capabilities.

---

## 场景一：Onboarding — Rich Context 生成

> 使用数据库: spider_tvshow (3 表)  
> 前置条件: Rich Context 已清空

**【操作】** 在首页点击 Spider Dataset 卡片，进入 workspace。  
**【中文】** 我们先用一个 3 张表的 TV 节目数据库来演示 Onboarding 流程。  
**【EN】** We start with a 3-table TV show database to demonstrate the Onboarding pipeline.

**【操作】** 点击 Sync Schema 按钮，等待同步完成。  
**【中文】** 首先同步数据库 Schema 到 Lakebase 存储层。  
**【EN】** First, we sync the database schema into the Lakebase storage layer.

**【操作】** 切换到 Context Tab，展示空状态页面。  
**【中文】** 当前数据库还没有任何 Rich Context。我们点击生成按钮。  
**【EN】** The database has no Rich Context yet. Let's click generate.

**【操作】** 点击 Generate Rich Context → 弹出 Generation Console → 点击 Start Generation。  
**【中文】** 系统启动 ReAct Agent，开始逐表分析。  
**【EN】** The system launches a ReAct Agent and begins analyzing each table.

**【操作】** 等待。Console 实时滚动日志，展示 Agent 的 thought → action → observation 循环。**让日志滚动几秒，让观众看清 ReAct 过程。**  
**【中文】** Agent 会自动采样数据、分析分布、发现语义关系。比如它发现 Episode 表的 rating 字段是 1 到 10 的观众评分。  
**【EN】** The agent automatically samples data, analyzes distributions, and discovers semantic relationships. For example, it identifies that the rating column in the Episode table represents audience scores from 1 to 10.

**【操作】** 生成完成后，关闭 Console。  
**【中文】** 几秒钟后，三张表的 Rich Context 全部生成完毕。  
**【EN】** Within seconds, Rich Context for all three tables is fully generated.

**【操作】** 在 Context Manager 界面，展开一张表，看到按列分组的 context 条目。鼠标悬停几个条目，展示 description / synonym / value mapping 等标签。  
**【中文】** 现在 Context Manager 按表和列分组展示所有条目。每个条目带有类型标签——description、example、synonym、value mapping、business rule 等等。这些 Rich Context 会在后续查询中被注入 LLM prompt，帮助模型理解业务语义。  
**【EN】** The Context Manager now displays all entries grouped by table and column. Each entry carries a type label — description, example, synonym, value mapping, business rule, and so on. These Rich Context entries will be injected into the LLM prompt during subsequent queries, helping the model understand domain semantics.

**【操作】** 停顿 2 秒。  
**【中文】** 这就是 Onboarding 管线——一键从零到完整的语义知识库。  
**【EN】** That's the Onboarding pipeline — one click from zero to a complete semantic knowledge base.

---

## 场景二：Rich Context 前后对比 — 语义消歧

> 使用数据库: spider_tvshow (3 表)  
> 前置条件: 场景一刚完成 Onboarding，RC 已生成  
> 关键点: 先关 RC 查一次（答错）→ 开 RC 查一次（答对），同一个问题形成对比

### Part A — 无 Rich Context（预期答错）

**【操作】** 切换到 Query Tab。关闭左侧 Rich Context 开关。  
**【中文】** 现在我们关闭 Rich Context，看看裸 Schema 能不能回答一个有歧义的问题。  
**【EN】** Now we turn off Rich Context and see whether the raw schema alone can handle an ambiguous question.

**【操作】** 在输入框中输入 `Which channel has the highest share?`，然后回车。  
**【中文】** 注意，我们只说了 share——这个词可以是收视份额、股份、市占率。裸 Schema 里 Share 只是一个 DECIMAL 列，没有任何注释，LLM 无从判断含义。  
**【EN】** Notice we only say "share" — this word could mean audience share, stock ownership, or market share. In the raw schema, Share is just a DECIMAL column with no comments, giving the LLM no way to determine its meaning.

**【操作】** 等待结果。**观察 SQL 是否选错了字段或聚合逻辑。** 停留几秒，让观众看到错误。  
**【中文】** 可以看到，模型选错了字段（或者生成了不合理的查询），因为它无法理解 Share 在电视行业的业务含义。  
**【EN】** As we can see, the model selects the wrong column or generates an unreasonable query, because it cannot understand what Share means in the TV broadcasting domain.

### Part B — 开启 Rich Context（预期答对）

**【操作】** 点击 Clear，打开左侧 Rich Context 开关。再次输入同一个问题 `Which channel has the highest share?`，回车。  
**【中文】** 现在打开 Rich Context，问完全相同的问题。  
**【EN】** Now we enable Rich Context and ask the exact same question.

**【操作】** 等待卡片 2 亮起：One-Shot Schema Linking。**稍作停留，让观众看清推理内容。**  
**【中文】** 看 Linking Agent 的推理——Rich Context 标注了 TV_series.Share 是"所有收看电视的家庭中，收看该节目的百分比"。有了这条标注，模型选对了字段，理解了按 Channel 分组取最大值。  
**【EN】** Look at the Linking Agent's reasoning — Rich Context annotates TV_series.Share as "the percentage of all TV households tuned to this program." With this annotation, the model selects the correct column and understands the aggregation should group by channel and take the maximum.

**【操作】** 等待 SQL 生成和执行结果。  
**【中文】** 同一个问题，加了 Rich Context 后结果完全正确。前后对比一目了然。  
**【EN】** Same question — with Rich Context, the result is completely correct. The before-and-after comparison speaks for itself.

**【操作】** 停顿 2 秒。  
**【中文】** 这就是 Rich Context 的核心价值——不靠改问题措辞，靠知识库消歧。  
**【EN】** This is the core value of Rich Context — disambiguation through a knowledge base, not prompt rewording.

---

## 场景三：Schema Evolution — Agent 自维持

> 使用数据库: lucid_evolution (2 表初始 → 5 阶段 DDL 演进)  
> 前置条件: 已 Reset 到 Stage 0/5

**【操作】** 返回首页，点击 Evolution Demo 卡片进入。  
**【中文】** 接下来演示 Agent 自维持机制。这个数据库初始只有 users 和 orders 两张表。  
**【EN】** Next, we demonstrate the agent self-maintenance mechanism. This database initially has only two tables — users and orders.

**【操作】** 切换到 Evolution Tab，展示 Stage 进度条。  
**【中文】** 我们会执行几个阶段的 DDL 变更，每次变更后 Agent 自动维护 Rich Context。  
**【EN】** We will execute several stages of DDL changes. After each change, the agent automatically maintains the Rich Context.

**【操作】** 点击 Execute Next Stage（Stage 1: Add User Phone Column）。等待实时日志流。  
**【中文】** Stage 1，给 users 表添加 phone 列。  
**【EN】** Stage 1 — adding a phone column to the users table.

**【操作】** **让日志滚动几秒**——DDL executing → Changes detected → Schema synced → Agent running。  
**【中文】** 执行完 DDL 后，系统自动 Detect Changes——检测到 column_added 变更，同步 schema，触发自维护管线。  
**【EN】** After executing the DDL, the system automatically detects changes — it identifies a column_added event, syncs the schema, and triggers the self-maintenance pipeline.

**【操作】** 日志中出现 Coordinator → Executor，LLM 生成新 context。等待完成。  
**【中文】** Coordinator 协调维护任务，Executor 用 LLM 为新列生成语义描述和示例值。  
**【EN】** The Coordinator dispatches maintenance tasks, and the Executor uses the LLM to generate semantic descriptions and sample values for the new column.

**【操作】** Stage 1 完成（绿色勾）。点击 Execute Next Stage（Stage 2: Create Products Table）。  
**【中文】** Stage 2，创建 products 新表。这次变更更大——整张表的 Rich Context 都需要从零生成。  
**【EN】** Stage 2 — creating a new products table. This is a larger change — Rich Context for the entire table must be generated from scratch.

**【操作】** 等待日志流：table_added → 全套 context 生成。  
**【中文】** Agent 检测到新表，自动为 products 的每一列生成 description、example、synonym 等完整 context。  
**【EN】** The agent detects the new table and automatically generates complete context for every column — descriptions, examples, synonyms, and more.

**【操作】** Stage 2 完成。点击 Execute Next Stage（Stage 3: Add FK）。  
**【中文】** Stage 3，给 orders 添加 product_id 外键。这会触发关系图更新。  
**【EN】** Stage 3 — adding a product_id foreign key to the orders table. This triggers a relationship graph update.

**【操作】** 等待日志：column_added + fk_added → context 更新。完成后展示进度条和 Change Log 面板。  
**【中文】** 系统不仅检测到新列，还检测到了外键关系，自动更新了关联表的 context。每一步变更都记录在 Change Log 中，包含 before/after 对比。整个过程全自动，无需人工干预。  
**【EN】** The system detects not only the new column but also the foreign key relationship, and automatically updates context for the related tables. Every change is recorded in the Change Log with before/after diffs. The entire process is fully automatic — no manual intervention required.

---

## 场景四：大规模 Adaptive Schema Linking

> 使用数据库: tpch_enterprise (517 表，21 个业务域)  
> 前置条件: Rich Context 已预生成

### Part A — Forest-Based Onboarding（快速展示）

**【操作】** 返回首页，点击 TPC-H Enterprise 卡片进入。  
**【中文】** 最后一个场景——517 张表的企业级数据库，跨越 HR、财务、CRM、供应链等 21 个业务域。  
**【EN】** The final scenario — a 517-table enterprise database spanning 21 business domains including HR, finance, CRM, and supply chain.

**【操作】** 进入 workspace，在 Context Tab 查看。  
**【中文】** 这个规模的数据库，传统方法要么超出 LLM 上下文窗口，要么检索效率极低。  
**【EN】** At this scale, traditional approaches either exceed the LLM context window or suffer from extremely low retrieval efficiency.

**【操作】** 点击 Generate → Console 弹出，显示 Forest 模式 banner（517 tables > threshold）。  
**【中文】** 系统检测到表数超过阈值，自动启用 Forest-Based Chunked Onboarding。Schema 被 FK 图分解成约 40 个子树。  
**【EN】** The system detects that the table count exceeds the threshold and automatically enables Forest-Based Chunked Onboarding. The schema is decomposed by the foreign-key graph into approximately 40 subtrees.

**【操作】** 展示 Treemap 预览。**停留 3-5 秒让观众看清。不需要真的跑完。**  
**【中文】** Treemap 可视化展示了 Forest 分解结果。每个色块是一个 FK 连通子图，面积与表数成正比。  
**【EN】** The treemap visualizes the forest decomposition. Each block represents an FK-connected subgraph, with area proportional to the number of tables.

### Part B — Two-Stage Linking 查询

**【操作】** 关闭 Console，切到 Query Tab。  
**【中文】** 现在我们直接查询，看两阶段 Linking 如何高效处理大规模 schema。  
**【EN】** Now let's query directly and see how two-stage linking efficiently handles a large-scale schema.

**【操作】** 输入 `Find the supplier with the highest profit`，回车。  
**【中文】** 问一个需要跨表关联的供应链分析问题。  
**【EN】** We ask a supply chain analysis question that requires joining multiple tables.

**【操作】** 等待卡片 1 亮起：Vector Search，显示约 20 个候选表。  
**【中文】** Stage 1，向量检索在约 1 秒内将 517 张表缩减到约 20 个候选。这一步避免了把全部 schema 塞进 LLM。  
**【EN】** Stage 1 — vector retrieval narrows 517 tables down to about 20 candidates in roughly one second. This avoids stuffing the entire schema into the LLM.

**【操作】** 等待卡片 2 亮起：One-Shot Schema Linking，显示 LLM 从 20 个中精选出的表和 reasoning。  
**【中文】** Stage 2，Linking Agent 对候选表做精确推理，最终选出真正相关的表。向量粗筛加 LLM 精选，两阶段配合。  
**【EN】** Stage 2 — the Linking Agent performs precise reasoning over the candidates and selects the truly relevant tables. Vector coarse filtering plus LLM fine selection — two stages working in tandem.

**【操作】** 等待卡片 3：ReAct SQL Generation → Generated SQL → 执行结果。  
**【中文】** SQL 正确生成并执行。从 517 张表到精准查询，整个过程实时透明。  
**【EN】** The SQL is correctly generated and executed. From 517 tables to a precise query, the entire process is transparent in real time.

---

## 片尾

**【操作】** 返回首页全景。停留 3 秒。

**【中文】** ATLAS 将 Schema、Rich Context 和向量索引统一存储在一张 MariaDB 中，实现了从 Onboarding 生成、查询推理、到 Schema 演进的完整闭环——端到端，无外挂，全自动。在 BIRD 基准上达到 75.55% 的执行准确率，在 500 张表以上的企业级 schema 上也能快速完成端到端查询。感谢观看。

**【EN】** ATLAS stores schema, Rich Context, and vector indexes within a single MariaDB instance, achieving a complete closed loop from onboarding generation, to query inference, to schema evolution — end-to-end, no external dependencies, fully automatic. It achieves 75.55% execution accuracy on the BIRD benchmark, and handles end-to-end queries efficiently on enterprise schemas with over 500 tables. Thank you for watching.

---

## 纯英文配音稿（TTS 直接使用）

以下为完整英文配音文稿，可直接复制给 TTS 工具。每个 `[pause]` 处插入 1.5 秒静音。

```
ATLAS is an end-to-end Text-to-SQL system built on MariaDB's native vector capabilities. It stores schema, Rich Context, and vector indexes within a single database — no external engines required. Three Docker containers, one command to deploy the full stack. We will walk through three scenarios to demonstrate the system's core capabilities.

[pause]

We start with a 3-table TV show database to demonstrate the Onboarding pipeline.

First, we sync the database schema into the Lakebase storage layer.

The database has no Rich Context yet. Let's click generate.

The system launches a ReAct Agent and begins analyzing each table.

The agent automatically samples data, analyzes distributions, and discovers semantic relationships. For example, it identifies that the rating column in the Episode table represents audience scores from 1 to 10.

Within seconds, Rich Context for all three tables is fully generated.

The Context Manager now displays all entries grouped by table and column. Each entry carries a type label — description, example, synonym, value mapping, business rule, and so on. These Rich Context entries will be injected into the LLM prompt during subsequent queries, helping the model understand domain semantics.

That's the Onboarding pipeline — one click from zero to a complete semantic knowledge base.

[pause]

Now we turn off Rich Context and see whether the raw schema alone can handle an ambiguous question.

Notice we only say "share" — this word could mean audience share, stock ownership, or market share. In the raw schema, Share is just a DECIMAL column with no comments, giving the LLM no way to determine its meaning.

As we can see, the model selects the wrong column or generates an unreasonable query, because it cannot understand what Share means in the TV broadcasting domain.

[pause]

Now we enable Rich Context and ask the exact same question.

Look at the Linking Agent's reasoning — Rich Context annotates TV_series.Share as "the percentage of all TV households tuned to this program." With this annotation, the model selects the correct column and understands the aggregation should group by channel and take the maximum.

Same question — with Rich Context, the result is completely correct. The before-and-after comparison speaks for itself.

This is the core value of Rich Context — disambiguation through a knowledge base, not prompt rewording.

[pause]

Next, we demonstrate the agent self-maintenance mechanism. This database initially has only two tables — users and orders.

We will execute several stages of DDL changes. After each change, the agent automatically maintains the Rich Context.

Stage 1 — adding a phone column to the users table.

After executing the DDL, the system automatically detects changes — it identifies a column_added event, syncs the schema, and triggers the self-maintenance pipeline.

The Coordinator dispatches maintenance tasks, and the Executor uses the LLM to generate semantic descriptions and sample values for the new column.

Stage 2 — creating a new products table. This is a larger change — Rich Context for the entire table must be generated from scratch.

The agent detects the new table and automatically generates complete context for every column — descriptions, examples, synonyms, and more.

Stage 3 — adding a product_id foreign key to the orders table. This triggers a relationship graph update.

The system detects not only the new column but also the foreign key relationship, and automatically updates context for the related tables. Every change is recorded in the Change Log with before/after diffs. The entire process is fully automatic — no manual intervention required.

[pause]

The final scenario — a 517-table enterprise database spanning 21 business domains including HR, finance, CRM, and supply chain.

At this scale, traditional approaches either exceed the LLM context window or suffer from extremely low retrieval efficiency.

The system detects that the table count exceeds the threshold and automatically enables Forest-Based Chunked Onboarding. The schema is decomposed by the foreign-key graph into approximately 40 subtrees.

The treemap visualizes the forest decomposition. Each block represents an FK-connected subgraph, with area proportional to the number of tables.

Now let's query directly and see how two-stage linking efficiently handles a large-scale schema.

We ask a supply chain analysis question that requires joining multiple tables.

Stage 1 — vector retrieval narrows 517 tables down to about 20 candidates in roughly one second. This avoids stuffing the entire schema into the LLM.

Stage 2 — the Linking Agent performs precise reasoning over the candidates and selects the truly relevant tables. Vector coarse filtering plus LLM fine selection — two stages working in tandem.

The SQL is correctly generated and executed. From 517 tables to a precise query, the entire process is transparent in real time.

[pause]

ATLAS stores schema, Rich Context, and vector indexes within a single MariaDB instance, achieving a complete closed loop from onboarding generation, to query inference, to schema evolution — end-to-end, no external dependencies, fully automatic. It achieves 75.55% execution accuracy on the BIRD benchmark, and handles end-to-end queries efficiently on enterprise schemas with over 500 tables. Thank you for watching.
```

---

## 录制前 Checklist

- [ ] Docker 环境正常运行（`docker compose -f deploy/docker-compose.yml up -d`）
- [ ] 确认所有数据库都已连接（首页卡片可见）
- [ ] spider_tvshow 的 Rich Context **已清空**（场景一从零生成）
- [ ] lucid_evolution **已 Reset**（Stage 0/5 起始状态）
- [ ] tpch_enterprise 的 Rich Context **已预生成**（场景四查询需要）
- [ ] LLM API Key 配置正常，网络稳定
- [ ] Chrome 全屏，1920×1080，缩放 100%
- [ ] 隐藏书签栏、插件图标、清除地址栏历史建议
- [ ] 录屏软件已开启，录制区域确认

## 录制技巧

- 鼠标移动要慢而稳，别抖
- 每个操作后**留 2-3 秒停顿**，给剪辑留余量
- 关键 UI 变化（卡片亮起、日志滚动）时暂停几秒让观众看清
- Evolution 的 Stage 执行等待时间较长（10-20秒），**不用着急**，后期加速
- **念错了不要停**，继续念完当前段落，后期用 AI 英文配音替换

## AI 配音注意事项

- 语速：中等偏慢，约每分钟 140-160 词（英文）
- 语气：专业但不冷淡，像在会议上给同行做 live demo
- 每个 `[pause]` 处插入 1.5 秒静音
- 每句配音之间自然停顿 0.5-0.8 秒
- 专有名词发音：ATLAS（/ˈætləs/）、ReAct（/riːˈækt/）、HNSW（逐字母 H-N-S-W）、Lakebase（/leɪkbeɪs/）
