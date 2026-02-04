# LUCID - Lake-base Unified Context-aware Intelligence for Data

## 项目定位

**VLDB 2025/2026 Demo Track** 投稿项目

基于 **MariaDB 12** 原生 VECTOR + HNSW 能力的湖基多模 Text-to-SQL 系统。

## 核心创新

1. **湖基多模统一存储** - rc_* 表统一存储 Schema + Context + 向量
2. **Agent 自维持机制** - DDL 变更自动感知，Context 闭环更新
3. **库内向量检索 Linking** - MariaDB 原生 HNSW，两阶段 Schema Linking
4. **端到端无外挂** - 全流程在单一数据库内完成

## 目录结构

```
lucid/
├── backend/              # Go 后端
│   ├── config/           # 配置加载和结构定义
│   ├── interfaces/       # 核心接口定义
│   ├── bridge/           # 内部实现与接口桥接层
│   ├── internal/         # 核心业务模块
│   │   ├── lakebase/         # 湖基存储层 (rc_* 表)
│   │   ├── agent/            # 自维持 Agent
│   │   ├── grounding/        # Schema Linking
│   │   ├── inference/        # ReAct 推理引擎
│   │   ├── embedding/        # 向量嵌入
│   │   ├── context/          # Rich Context 管理
│   │   ├── adapter/          # 数据库适配器
│   │   └── llm/              # LLM 客户端
│   └── server/           # HTTP API + SSE
│       ├── handlers/         # API 处理器
│       ├── services/         # 业务服务层
│       └── configs/          # 服务配置文件
├── frontend/             # Vue3 + Vite + UnoCSS + Naive UI
├── paper/                # VLDB Demo 论文 LaTeX
├── deploy/               # Docker 部署配置
└── docs/                 # 开发文档
```

## 快速命令

```bash
# 开发
make dev              # Docker 全栈启动
make backend-dev      # 本地启动后端
make frontend-dev     # 本地启动前端

# 数据库
make db-up            # 启动 MariaDB
make db-login         # 连接数据库

# 论文
make paper            # 编译 PDF
make paper-watch      # 实时预览
```

## 数据库连接

```bash
Host: 127.0.0.1
Port: 3310
User: root
Password: your_strong_password
Database: lucid
```

## 开发约束

1. **所有修改限于本仓库**
2. 后端模块边界：`backend/internal/` 各子目录职责清晰
3. 新增模块需在此文档更新目录结构
4. 向量维度：1536 (OpenAI text-embedding-3-small)

## 核心表结构 (rc_* 系列)

| 表名 | 职责 |
|------|------|
| rc_databases | 已接入的数据库连接信息 |
| rc_tables | 表级 Rich Context |
| rc_columns | 列级 Rich Context |
| rc_relations | 表关系图谱 |
| rc_terms | 业务术语词典 |
| rc_embeddings | 向量嵌入 (HNSW 索引) |
| rc_change_log | Context 变更审计日志 |

## Git 提交规范

```bash
feat(lakebase): add rc_embeddings table schema
feat(agent): implement DDL change detection
feat(frontend): add context manager UI
fix(grounding): fix vector search threshold
docs(paper): update system architecture section
```
