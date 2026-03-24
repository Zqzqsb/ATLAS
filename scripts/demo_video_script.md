# ATLAS Demo Video 脚本

> **总时长**: 约 5–6 分钟  
> **格式**: 屏幕录制 + AI 配音  
> **分辨率**: 1920×1080，浏览器全屏  
> **配音风格**: 专业、简洁、有节奏感，像产品发布会

---

## 🎬 片头 (0:00 – 0:15)

**【画面】** 黑屏渐入 → 系统首页 Landing Page，能看到数据库卡片网格

**【配音】**
> ATLAS 是一个基于 MariaDB 原生向量能力的端到端 Text-to-SQL 系统。
> 它用一张数据库统一存储 Schema、Rich Context 和向量索引，无需任何外部引擎。
> 接下来，我们通过四个场景，展示系统的核心能力。

---

## 📦 场景一：Onboarding — Rich Context 生成 (0:15 – 1:30)

> **对应论文**: Scenario 1: Onboarding Analysis (Rich Context lifecycle)  
> **使用数据库**: spider_tvshow (3 表)

### 操作步骤

1. **点击 Spider Dataset 卡片** → 进入 workspace
2. **切换到 "Context" Tab** → 此时页面显示 "No context yet"
3. **点击 "Generate Rich Context" 按钮** → 弹出 Generation Console
4. **点击 "Start Generation"** → 开始实时流式生成

### 镜头要点

| 时间 | 画面 | 配音 |
|------|------|------|
| 0:15 | 点击 Spider Dataset 卡片进入 workspace | 我们先用一个 3 张表的 TV 节目数据库来演示 Onboarding 流程。 |
| 0:22 | 切到 Context Tab，显示空状态 | 当前数据库还没有任何 Rich Context。我们点击生成按钮。 |
| 0:28 | 弹出 Generation Console → 点击 Start | 系统启动 ReAct Agent，开始逐表分析。 |
| 0:33 | Console 实时滚动日志，Agent 的 thought → action → observation 循环 | Agent 会自动采样数据、分析分布、发现语义关系。比如它发现 Episode 表的 rating 字段是 1 到 10 的观众评分。 |
| 0:50 | 生成完成，关闭 Console | 几秒钟后，三张表的 Rich Context 全部生成完毕。 |
| 0:55 | Context Manager 界面，展开表，看到按列分组的 context 条目 | 现在 Context Manager 按表和列分组展示所有条目。每个条目带有类型标签——description、example、synonym、value mapping、business rule 等等。 |
| 1:10 | 鼠标悬停几个条目，展示 desc / syn / map 等标签 | 这些 Rich Context 会在后续查询中被注入 LLM prompt，帮助模型理解业务语义。 |
| 1:20 | 简短停顿 | 这就是 Onboarding 管线——一键从零到完整的语义知识库。 |

---

## 🔍 场景二：Context-Enhanced SQL 生成 (1:30 – 3:00)

> **对应论文**: Scenario 2: Context-Enhanced SQL Generation  
> **使用数据库**: spider_tvshow (3 表，已有 Rich Context)

### 操作步骤

1. **切换到 "Query" Tab**
2. **在输入框输入自然语言问题**（建议: `Which channel has the highest audience share?`）
3. **点击发送 / 回车** → 观察三阶段实时卡片

### 镜头要点

| 时间 | 画面 | 配音 |
|------|------|------|
| 1:30 | 切到 Query Tab，左侧参数面板 + 右侧执行区 | 现在进入推理管线。左侧可以选择模型和 Linking Mode。 |
| 1:38 | 输入 "Which channel has the highest audience share?" → 回车 | 我们问一个关键问题——"哪个频道的观众份额最高？" |
| 1:42 | **卡片 1: Vector Search / Schema Loaded** 亮起，显示 3 tables | 因为只有 3 张表，系统走 Small Scale 路径，直接加载全量 Schema。 |
| 1:48 | **卡片 2: One-Shot Schema Linking** 亮起，显示 LLM 选出的表和列，以及 reasoning | Schema Linking Agent 分析后选出了相关表和字段。注意它的推理过程——Rich Context 告诉模型 "Share" 是观众份额百分比，而不是股份或股权。 |
| 2:05 | **卡片 3: ReAct SQL Generation** 展开，thought/action/observation 循环 | SQL Generator 用 ReAct 模式逐步构建查询。它先用 verify_sql 做 dry-run 检查执行计划，确认无误后输出最终 SQL。 |
| 2:25 | **Generated SQL 区域** 显示完整 SQL + 自动执行结果表格 | SQL 自动执行，返回了正确的结果。没有 Rich Context 的话，LLM 很可能会把 Share 理解成"所有权百分比"，生成完全错误的查询。 |
| 2:40 | 稍微滚动回顾整个流程 | 从自然语言到正确 SQL，整个过程实时透明——每一步的推理和决策都清晰可见。 |

---

## 🔄 场景三：Schema Evolution — Agent 自维持 (3:00 – 4:30)

> **对应论文**: Scenario 3: Schema Evolution  
> **使用数据库**: lucid_evolution (2 表初始 → 5 阶段 DDL 演进)

### 操作步骤

1. **返回首页** → 点击 **Evolution Demo** 卡片进入
2. **切换到 "Evolution" Tab**
3. **点击 "Execute Next Stage"** → 观察实时日志
4. **重复 2–3 个 Stage**，展示不同类型的 DDL 变更

### 镜头要点

| 时间 | 画面 | 配音 |
|------|------|------|
| 3:00 | 返回首页，点击 Evolution Demo 卡片 | 接下来演示 Agent 自维持机制。这个数据库初始只有 users 和 orders 两张表。 |
| 3:08 | Evolution Tab，显示 Stage 1/5，进度条 | 我们会执行 5 个阶段的 DDL 变更，每次变更后 Agent 自动维护 Rich Context。 |
| 3:15 | 点击 Execute Next Stage → Stage 1: Add User Phone Column | **Stage 1**：给 users 表添加 phone 列。 |
| 3:18 | 实时日志流：DDL executing → Changes detected → Schema synced → Agent running | 系统自动检测到 column_added 变更，同步 schema，然后触发维护 Agent。 |
| 3:30 | 日志中出现 Coordinator → Executor，LLM 生成新 context | Coordinator 协调维护任务，Executor 用 LLM 为新列生成语义描述和示例值。 |
| 3:40 | Stage 1 完成，绿色勾 → 点击 Execute Next Stage → Stage 2 | **Stage 2**：创建 products 新表。这次变更更大——整张表的 Rich Context 都需要从零生成。 |
| 3:50 | 日志流：table_added → 全套 context 生成 | Agent 检测到新表，自动为 products 的每一列生成 description、example、synonym 等完整 context。 |
| 4:00 | Stage 2 完成 → 点击 Stage 3: Add FK | **Stage 3**：给 orders 添加 product_id 外键。这会触发关系图更新。 |
| 4:08 | 日志：column_added + fk_added → context 更新 | 系统不仅检测到新列，还检测到了外键关系，自动更新了关联表的 context。 |
| 4:20 | 展示进度条 3/5，回顾 Change Log 面板 | 每一步变更都记录在 Change Log 中，包含 before/after 对比。整个过程全自动，无需人工干预。 |

---

## 🏢 场景四：大规模 Adaptive Schema Linking (4:30 – 5:45)

> **对应论文**: Scenario 4: Large-Scale Adaptive Linking  
> **使用数据库**: tpch_enterprise (517 表，21 个业务域)

### 操作步骤

**Part A — Forest-Based Onboarding (可选快速展示)**
1. 进入 TPC-H Enterprise → Context Tab
2. 点击 Generate → 弹出 Console，显示 Forest 模式预览 Treemap
3. 展示预览即可，**不需要真的跑完**（太耗时）

**Part B — Two-Stage Linking 查询**
1. 切到 Query Tab → 输入 `Which supplier has the highest account balance?`
2. 观察三张实时卡片的不同表现

### 镜头要点

| 时间 | 画面 | 配音 |
|------|------|------|
| 4:30 | 返回首页，点击 TPC-H Enterprise 卡片 | 最后一个场景——517 张表的企业级数据库，跨越 HR、财务、CRM、供应链等 21 个业务域。 |
| 4:38 | 进入 workspace，Context Tab | 这个规模的数据库，传统方法要么超出 LLM 上下文窗口，要么检索效率极低。 |
| 4:43 | 点击 Generate → Console 弹出，显示 Forest 模式 banner（517 tables > threshold） | 系统检测到表数超过阈值，自动启用 Forest-Based Chunked Onboarding。Schema 被 FK 图分解成约 40 个子树。 |
| 4:52 | 展示 Treemap 预览——色块大小 = 子树表数，颜色 = 状态 | Treemap 可视化展示了 Forest 分解结果。每个色块是一个 FK 连通子图，面积与表数成正比。 |
| 5:00 | 关闭 Console → 切到 Query Tab | 现在我们直接查询，看两阶段 Linking 如何高效处理大规模 schema。 |
| 5:05 | 输入 "Which supplier has the highest account balance?" → 回车 | 问一个供应商相关的问题。 |
| 5:10 | **卡片 1: Vector Search** 亮起，显示 ~20 个候选表 + 耗时 ~400ms | **Stage 1**：HNSW 向量检索在 400 毫秒内将 517 张表缩减到约 20 个候选。这一步避免了把全部 schema 塞进 LLM。 |
| 5:22 | **卡片 2: One-Shot Schema Linking** 显示 LLM 从 20 个中精选 5 个 + reasoning | **Stage 2**：Linking Agent 对候选表做精确推理，最终选出 5 张真正相关的表。向量粗筛加 LLM 精选，两阶段配合。 |
| 5:35 | **卡片 3: ReAct SQL Generation** → Generated SQL → 执行结果 | SQL 正确生成并执行。从 517 张表到精准查询，整个过程约 15 秒，实时透明。 |

---

## 🎬 片尾 (5:45 – 6:00)

**【画面】** 返回首页全景 → 渐暗

**【配音】**
> ATLAS 将 Schema、Rich Context 和向量索引统一存储在一张 MariaDB 中，
> 实现了从 Onboarding 生成、查询推理、到 Schema 演进的完整闭环——
> 端到端，无外挂，全自动。
> 感谢观看。

---

## 📝 录制注意事项

### 环境准备
- [ ] Docker 环境正常运行（`docker compose up -d`）
- [ ] 确认 5 个数据库都已连接（首页卡片全部绿灯）
- [ ] spider_tvshow 的 Rich Context **先清空**（演示场景一时从零生成）
- [ ] lucid_evolution **先 Reset**（确保 Stage 0/5 起始状态）
- [ ] tpch_enterprise 的 Rich Context **已预生成**（场景四查询需要）
- [ ] LLM API Key 配置正常，网络稳定

### 浏览器设置
- Chrome，1920×1080，缩放 100%
- 隐藏书签栏、插件图标
- 清除地址栏历史建议

### 录制技巧
- 鼠标移动要慢而稳，别抖
- 每个操作之间留 1-2 秒停顿，给后期剪辑留余量
- 关键 UI 变化（卡片亮起、日志滚动）时暂停几秒让观众看清
- Evolution 的 Stage 执行可能需要 10-20 秒等待，后期可以适当加速

### AI 配音建议
- 语速：中等偏慢，每分钟 150-180 字
- 语气：专业但不冷淡，像在会议上给同行做 live demo
- 关键术语读英文原文：Rich Context、Schema Linking、ReAct、HNSW、Onboarding
- 数字要清晰：517 张表、400 毫秒、20 个候选、5 张表
