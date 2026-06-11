# ATLAS 迭代历史

## Iter1: 基础架构搭建
- 建立 Go + Gin 后端骨架
- MariaDB 连接池 + lakebase rc_* 表设计
- 前端 Vue 3 项目初始化

## Iter2: Onboarding 管线
- Schema 同步（INFORMATION_SCHEMA → rc_tables/rc_columns）
- LLM 生成 Rich Context（表描述、列语义注解、术语映射）
- Embedding 入库（Doubao 1536 维）

## Iter3: 推理管线 v1
- 单阶段 Schema Linking
- OneShot SQL 生成
- 基础 SSE 流式输出

## Iter4: ReAct Agent + 两阶段 Grounding
- ReAct 循环 + execute_sql/verify_sql 工具
- 向量粗筛（CoarseRetriever 4路并行）+ LLM 精选（LinkingAgent）
- SmallScale/LargeScale 自适应策略

## Iter5: Lakebase 存储重构 + Embedding
- rc_* 表体系完善（rc_terms, rc_relations, rc_embeddings）
- HNSW 向量索引
- ConnectionPool 重构

## Iter6: 系统完善 + 自维持管线
- Agent 模块（DDLDetector, ContextMaintainer, EvolutionService, ChangeLogger）
- Evolution 5阶段 DDL 演进脚本设计
- 前端 SelfMaintainDemo.vue（660行）
- Field Alignment 修复
- 指标页面规划

## 当前阶段 (Iter0 / 新周期)
- 推理管线现状分析完成
- 自维持管线现状分析完成，识别 7 个阻塞问题
- 行动计划制定：5 步修复自维持管线 → TPC-H 大规模库 → 论文

### 自维持管线阻塞问题
1. atlas_evolution 库未被 docker init 自动创建
2. Evolution DB 未注册为 rc_datasources
3. Reset 后无初始 Rich Context
4. getBusinessDB() 硬编码错误密码
5. Agent 后台循环未 Start()
6. Reset 后不自动触发 Onboarding
7. 前端 datasourceId 硬编码为 1
