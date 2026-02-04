# LUCID

> **L**akebase-**U**nified **C**ontext-aware **I**ntelligence for **D**ata

基于湖基多模原生能力的企业级 Text-to-SQL 系统

## 核心特性

- 🗄️ **湖基统一存储** - Schema、Context、向量嵌入存储在同一 MariaDB 实例
- 🤖 **Agent 自维持** - DDL 变更自动感知，Context 闭环更新
- 🔍 **库内向量检索** - MariaDB 原生 HNSW 索引，毫秒级 Schema Linking
- 🔗 **端到端无外挂** - 无需 Milvus、Elasticsearch 等外部组件

## 快速开始

```bash
# 启动全栈
make dev

# 访问前端
open http://localhost:3000

# 连接数据库
make db-login
```

## 技术栈

| 组件 | 技术 |
|------|------|
| 数据库 | MariaDB 12 (VECTOR + HNSW) |
| 后端 | Go 1.24 + Gin |
| 前端 | Vue 3 + Vite + UnoCSS + Naive UI |
| LLM | OpenAI / 混元 |

## 项目目标

VLDB 2025/2026 Demo Track 投稿

## License

MIT
