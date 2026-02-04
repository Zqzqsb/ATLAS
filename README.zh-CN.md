# LUCID

**L**akebase-**U**nified **C**ontext-aware **I**ntelligence for **D**ata（湖基多模统一上下文感知数据智能系统）

一个自包含的 Text-to-SQL 系统，具备原生向量检索能力，面向 VLDB 2025/2026 Demo Track 投稿。

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?logo=docker&logoColor=white)](deploy/docker-compose.yml)

[English](README.md) | [简体中文](README.zh-CN.md)

## 快速开始

一行命令部署整个系统（需要 Docker）：

```bash
git clone https://github.com/your-repo/lucid.git
cd lucid
docker compose -f deploy/docker-compose.yml up -d
```

访问 **http://localhost:19000** 使用系统

## 核心特性

🗄️ **湖基原生存储**
- Schema、上下文、向量嵌入存储在单一 MariaDB 实例
- 无需外部向量数据库（Milvus、Elasticsearch 等）

🤖 **Agent 自维持**
- 自动 DDL 变更检测
- 闭环上下文更新
- 零维护 Rich Context

🔍 **库内向量检索**
- MariaDB 12 原生 VECTOR + HNSW 索引
- 毫秒级 Schema Linking
- 两阶段检索（粗筛 + 精排）

🔗 **端到端集成**
- 单一 Docker Compose 部署
- 无外部依赖
- 生产就绪架构

## 系统架构

```
┌─────────────────────────────────────────────────────┐
│                 LUCID 系统                          │
├─────────────────────────────────────────────────────┤
│                                                     │
│  前端 (Vue3)  ──→  后端 (Go)  ──→  MariaDB         │
│    :19000           :19001          :19010          │
│                                                     │
│  ┌──────────────────────────────────────────────┐  │
│  │ MariaDB 12 - 统一存储 (:19010)            │  │
│  │                                              │  │
│  │  湖基存储 (rc_* 表):                         │  │
│  │  ├─ rc_datasources  (元数据)                │  │
│  │  ├─ rc_tables       (Rich Context)          │  │
│  │  ├─ rc_columns      (Rich Context)          │  │
│  │  ├─ rc_embeddings   (VECTOR + HNSW)         │  │
│  │  └─ rc_change_log   (审计日志)              │  │
│  │                                              │  │
│  │  演示数据库:                                  │  │
│  │  ├─ demo_ecommerce  (电商)                   │  │
│  │  └─ demo_tpch       (TPC-H 基准测试)        │  │
│  └──────────────────────────────────────────────┘  │
│                                                     │
└─────────────────────────────────────────────────────┘
```

## 技术栈

| 组件 | 技术 |
|------|------|
| 数据库 | MariaDB 12 (VECTOR + HNSW) |
| 后端 | Go 1.24 + Gin |
| 前端 | Vue 3 + Vite + UnoCSS + Naive UI |
| LLM | OpenAI / 混元 |
| 部署 | Docker + Docker Compose |

## 端口分配

LUCID 使用 **19xxx** 端口段以避免冲突：

| 服务 | 端口 | 说明 |
|------|------|------|
| 前端 | 19000 | Web UI |
| 后端 | 19001 | REST API + SSE |
| MariaDB | 19010 | 湖基存储 + 演示数据库 |

## 使用方法

### 启动所有服务

```bash
make up
# 或
docker compose -f deploy/docker-compose.yml up -d
```

### 查看日志

```bash
make logs
# 或
docker compose -f deploy/docker-compose.yml logs -f
```

### 停止服务

```bash
make down
# 或
docker compose -f deploy/docker-compose.yml down
```

### 本地开发

```bash
# 后端 (Go)
make backend-dev

# 前端 (Vue3)
make frontend-dev

# 仅启动数据库
make db-up
```

## 配置

创建 `.env` 文件自定义端口和 API 密钥：

```bash
# 可选端口覆盖
LUCID_FRONTEND_PORT=19000
LUCID_BACKEND_PORT=19001
LUCID_MARIADB_PORT=19010

# LLM API Key
LLM_API_KEY=sk-your-api-key-here
```

## 数据库连接

连接到湖基存储：

```bash
# 使用 mycli
make db-login

# 或手动连接
mycli -h 127.0.0.1 -P 19010 -u lucid -plucid2024 lucid
```

## 文档

- [部署指南](docs/SETUP.md) - 详细部署说明
- [系统架构](docs/ARCHITECTURE.md) - 系统设计和组件
- [开发指南](CLAUDE.md) - 贡献者文档

## 研究与引用

LUCID 面向 **VLDB 2025/2026 Demo Track** 投稿。如果您在研究中使用本系统，请引用：

```bibtex
@inproceedings{lucid2026,
  title={LUCID: Lake-Base Unified Context-Aware Intelligence for Data},
  author={Your Name},
  booktitle={Proceedings of the VLDB Endowment},
  year={2026}
}
```

## 核心创新

1. **湖基多模存储** - Schema、上下文、向量统一存储
2. **Agent 自维持** - DDL 变更时自动更新上下文
3. **库内向量检索** - 原生 HNSW 索引实现 Schema Linking
4. **端到端集成** - 零外部依赖

## 贡献

欢迎贡献！详见 [CONTRIBUTING.md](CONTRIBUTING.md)。

## 许可证

MIT License - 详见 [LICENSE](LICENSE)

## 致谢

- MariaDB Foundation 提供 VECTOR 支持
- Spider 数据集用于基准测试
- VLDB 社区的启发

---

**注意**：这是面向 VLDB Demo Track 的研究原型。生产环境使用前，请检查安全配置并添加适当的身份验证。
