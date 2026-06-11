# ATLAS

**A**daptive **T**ext-to-SQL with **L**ifecycle-**A**ware **S**elf-maintaining Context

> VLDB 2026 Demo Track

ATLAS 是一个自包含的 Text-to-SQL 系统，将 Schema 元数据、语义标注和向量嵌入全部存储在单一 RDBMS 内。三个 Docker 容器，一条命令部署，无需任何外部引擎。

[![License: Apache 2.0](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?logo=docker&logoColor=white)](deploy/docker-compose.yml)
[![BIRD EX](https://img.shields.io/badge/BIRD_dev-76.40%25_EX-brightgreen)](#评估结果)

[English](README.md) | [简体中文](README.zh-CN.md)

<p align="center">
  <img src="paper/figures/demo_ui.png" alt="ATLAS Demo 界面" width="100%"/>
</p>
<p align="center"><em>(a) 517 表 Forest-Chunked Onboarding &nbsp; (b) 两阶段自适应查询 &nbsp; (c) 自主 Schema 演进</em></p>

## 四项核心创新

### 1. 库内统一存储

Schema、Rich Context、关系图谱、向量嵌入（HNSW）和变更审计日志全部存放在 MariaDB 12 的 `rc_*` 系列表中。一条 SQL 即可同时做向量相似度搜索和关系型过滤——无外部向量库、无一致性问题、完整 ACID 保障。

### 2. 两阶段自适应 Schema Linking

- **小规模** (≤30 表)：全量 Schema 直接发给 LLM 做 one-shot linking。
- **大规模** (>30 表)：向量检索在亚秒级内将 500+ 张表缩减到 ~20 个候选；LLM 再做精确推理。两阶段通过原子槽并发执行，完全隐藏检索延迟。

### 3. Rich Context 生命周期

Rich Context 不是一次性静态标注，而是经历三个阶段：

| 阶段 | 内容 |
|------|------|
| **Onboarding** | ReAct Agent 采样数据，为每列生成描述/同义词/业务规则，嵌入 HNSW 索引 |
| **Inference** | 向量检索召回相关 Context 注入 LLM prompt，辅助语义消歧 |
| **Evolution** | DDL 变更检测 → 标记过时 Context → LLM 重新生成 → 向量重新嵌入 |

大规模 Schema (>30 表) 使用 **Forest-Based Chunked** 策略，将 FK 图分解为连通子树并行处理。

### 4. Agent 驱动的自维持

Coordinator–Executor 架构保持 Context 与活跃 Schema 同步：

1. **DDL 检测器** 对比 `information_schema` 与 Context 表的差异
2. **Coordinator** 标记受影响条目为过时，规划维护任务
3. **Executor** 调用 LLM 重新生成描述并重新嵌入向量
4. **Change Logger** 记录所有变更的 before/after 对比

## 评估结果

**BIRD 开发集** (1,534 问题, 11 数据库)：

| 配置 | EX (%) | 平均迭代次数 |
|------|--------|------------|
| **完整 ATLAS 管线** | **76.40** | 3.37 |
| − ReAct 循环 (one-shot + RC) | 68.71 | 1.00 |
| − 业务规则与值映射 | 72.04 | 3.62 |
| − 样例值与同义词 | 70.86 | 3.91 |
| 仅 Schema (无 Rich Context) | 65.45 | 4.49 |
| 基线 (直接生成) | 58.93 | 1.00 |

**系统级消融实验** — TPC-H Enterprise (500+ 表, 30 个跨域查询)：

| 配置 | Recall@20 | EX (%) | 延迟 (s) |
|------|-----------|--------|---------|
| 完整 ATLAS 管线 | **93.3** | **70.0** | 4.8 |
| − 自适应 Linking | — (溢出) | — | 超时 |
| − 向量检索 | 66.7 | 50.0 | 5.6 |
| − ReAct 循环 | 93.3 | 56.7 | 2.3 |
| − Rich Context | 80.0 | 53.3 | 4.9 |

> 详细消融结果: [AtlasCore](https://github.com/atlas-demo/AtlasCore)

## 系统架构

<p align="center">
  <img src="paper/figures/architecture.png" alt="ATLAS 架构" width="720"/>
</p>
<p align="center"><em>三条管线 — Onboarding、Inference、Self-Maintenance — 共享库内统一存储 (rc_* 表)。</em></p>

## 快速开始

```bash
git clone https://github.com/atlas-demo/atlas.git
cd atlas
docker compose -f deploy/docker-compose.yml up -d
```

访问 **http://localhost:19000** 使用系统

## 技术栈

| 组件 | 技术 |
|------|------|
| 数据库 | MariaDB 12 (原生 VECTOR + HNSW) |
| 后端 | Go 1.24 + Gin |
| 前端 | Vue 3 + Vite + UnoCSS + Naive UI |
| LLM | 任意 OpenAI 兼容 API |
| 嵌入模型 | 任意 OpenAI 兼容 Embedding API |
| 部署 | Docker Compose (3 容器) |

## 使用方法

```bash
# 启动所有服务
make up

# 查看日志
make logs

# 停止
make down

# 本地开发
make backend-dev    # Go 后端
make frontend-dev   # Vue3 前端
make db-up          # 仅启动数据库
```

## 项目结构

```
atlas/
├── backend/              # Go 后端
│   ├── internal/
│   │   ├── lakebase/         # 湖基存储层 (rc_* 表)
│   │   ├── agent/            # 自维持 Agent
│   │   ├── grounding/        # Schema Linking
│   │   ├── inference/        # ReAct 推理引擎
│   │   ├── embedding/        # 向量嵌入
│   │   ├── context/          # Rich Context 管理
│   │   ├── adapter/          # 数据库适配器
│   │   └── llm/              # LLM 客户端
│   └── server/               # HTTP API + SSE
├── frontend/             # Vue3 + Vite + UnoCSS + Naive UI
├── AtlasCore/            # 实验框架 (submodule)
├── paper/                # VLDB Demo 论文 (LaTeX)
├── deploy/               # Docker Compose 配置
└── scripts/              # Demo 视频脚本
```

## 引用

```bibtex
@inproceedings{atlas2026vldb,
  title     = {ATLAS: Adaptive Text-to-SQL with Lifecycle-Aware Self-maintaining Context},
  author    = {Anonymous},
  booktitle = {Proceedings of the VLDB Endowment, Demo Track},
  year      = {2026}
}
```

## 许可证

Apache License 2.0 — 详见 [LICENSE](LICENSE)

## 致谢

- MariaDB Foundation 提供原生 VECTOR 支持
- [BIRD](https://bird-bench.github.io/) 和 [Spider](https://yale-lily.github.io/spider) 基准测试
