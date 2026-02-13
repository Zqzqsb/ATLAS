---
name: lucid-project-context
description: >
  LUCID 项目全景知识 — VLDB Demo Track 湖基多模 Text-to-SQL 系统。
  This skill should be used when working on any part of the LUCID codebase,
  including backend (Go), frontend (Vue), deployment, or paper writing.
  It provides architecture overview, module relationships, design constraints,
  pipeline details, and iteration history.
---

# LUCID 项目上下文

## 定位

VLDB 2025/2026 Demo Track 投稿。基于 MariaDB 12 原生 VECTOR + HNSW 的湖基多模 Text-to-SQL 系统。

## 四大创新点（论文核心叙事）

1. **湖基多模统一存储** — rc_* 表统一存储 Schema + Rich Context + 向量，无 Redis/Milvus/ES
2. **Agent 自维持机制** — DDL 变更检测 → Context 过期标记 → LLM 增量刷新 → 闭环验证
3. **两阶段 Adaptive Schema Linking** — Stage 1 HNSW 向量粗筛 → Stage 2 LLM 精排
4. **Rich Context 生命周期** — Onboarding 生成 → 查询使用 → Evolution 演进 → 完整闭环

## 关键原则

- **Rich Context 是核心差异化**，任何优化不能削弱 context 在管线中的作用
- 所有存储在 MariaDB 内完成（rc_* 表 + HNSW 向量索引），这是"端到端无外挂"的基础
- Embedding/LLM 调用依赖外部 API，论文措辞需说清"存储和检索端到端"
- 向量维度 1536（Doubao embedding）

## 技术栈

- 后端: Go 1.24 + Gin + langchaingo
- 前端: Vue 3 + Vite + UnoCSS + Naive UI
- 数据库: MariaDB 12 (VECTOR + HNSW)
- LLM: OpenAI 兼容 API (deepseek-v3, qwen-max 等)

## 目录结构

```
backend/
  cmd/lucid-server/main.go  — 入口，初始化所有 service
  internal/
    lakebase/    — rc_* 表存储层 (ConnectionPool, MySQLRepository, MySQLVectorRepository)
    agent/       — 自维持 (AgentService, DDLDetector, ContextMaintainer, EvolutionService, ChangeLogger)
    grounding/   — Adaptive Pipeline: SmallScale(≤30表)/LargeScale(>30表)
    inference/   — ReAct Pipeline: SchemaLinker + SQLGenerator(OneShot/ReAct) + VerifySQLTool
    embedding/   — OpenAI 兼容 Embedding Provider
    adapter/     — MySQL Adapter (ExecuteQuery, DryRunSQL)
    llm/         — LLM Config 加载
  server/
    handlers/    — API handlers (text2sql, evolution, onboarding, lakebase)
    services/    — InferenceService, InferenceEngine, LakebaseService
deploy/
  docker-compose.yml — MariaDB + backend + frontend 三容器
  init/mariadb/      — 01_init_lakebase.sql + 02-04 demo 数据库
  scripts/           — create_evolution_db.sql（未被 docker 自动执行）
frontend/
  src/
    features/landing/    — 数据库连接 (AddDatabaseDialog)
    features/workspace/  — 查询界面 (QueryChat, SchemaExplorer)
    components/demo/     — SelfMaintainDemo.vue (自维持演示)
```

## 三条管线

### 1. Onboarding 管线
接入数据库 → Schema 同步到 rc_* → LLM 生成 Rich Context → Embedding 入库

- 入口: `POST /api/v1/onboarding/start`
- handler: `onboarding_handlers.go`
- 核心: `lakebase/` 包的 SyncSchema + GenerateRichContext

### 2. 推理管线
Grounding(向量+LLM) → Field Alignment(可选) → SQL Generation(ReAct) → Execution

- 入口: `POST /api/v1/text2sql/stream` (SSE)
- Phase 1 Grounding: `grounding/` 包
  - SmallScale (≤30表): 全量 schema → LinkingAgent LLM 精选
  - LargeScale (>30表): 4路并行向量检索粗筛 → LinkingAgent LLM 精选
  - 降级链: LargeScale → SmallScale → 粗筛结果
- Phase 2 Inference: `inference/` 包
  - ReAct Agent 循环 + execute_sql/verify_sql 工具
  - "虚假预算"策略: 告诉 LLM 5 次迭代实际给 15 次

### 3. 自维持管线
DDL 检测 → Context 过期标记 → LLM 刷新 → Embedding 更新 → ChangeLog 记录

- 入口: `POST /api/v1/evolution/*` 系列 API
- handler: `evolution_handlers.go`
- 核心: `agent/` 包
  - `EvolutionService` — 5 阶段 DDL 演进脚本管理
  - `DDLDetector` — 对比 rc_tables/rc_columns 与实际 INFORMATION_SCHEMA
  - `ContextMaintainer` — LLM 刷新过期 context + Embedding 重建
  - `ChangeLogger` — 变更记录 (rc_change_logs)
  - `AgentService` — 编排以上组件的顶层 service

## Demo 数据库

| 库 | 表数 | 用途 | 状态 |
|---|---|---|---|
| spider_tvshow | 3 | 推理管线主演示 | ✅ 就绪 |
| spider_flight | 3 | 航班查询 | ✅ 就绪 |
| spider_wta | 8 | 网球数据 | ✅ 就绪 |
| lucid_evolution | 2(初始) | 自维持演示(5阶段DDL演进) | ⚠️ SQL 未被 docker 自动执行 |
| TPC-H | 30+ | 两阶段 Linking 效率演示 | ❌ 尚未创建 |

## 演示策略 — 一库一创新点

- Spider 库 → 推理管线 + Rich Context 对 SQL 生成的指导价值
- lucid_evolution → Agent 自维持（5 阶段 DDL 演进闭环）
- TPC-H 30+ 表 → 两阶段 Linking 效率（context 在大规模库表的意义）

## 迭代历史摘要

参阅 `references/iteration-history.md` 了解 Iter1-Iter6 的完整迭代历程。

## Git 提交规范

`feat(module): description` / `fix(module): description` / `data: description` / `ops: description` / `ui: description`
