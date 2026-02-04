# LUCID 项目上下文

> 本文档记录项目背景和开发上下文，供 Agent 快速理解项目状态

## 项目目标

**VLDB 2025/2026 Demo Track 投稿**

### 四大核心创新点

| 创新点 | 学术表述 | 技术实现 |
|--------|----------|----------|
| 1. 湖基多模统一存储 | Lake-Base Multi-Modal Native Storage | rc_* 系列表 + 向量索引 |
| 2. Agent 自维持机制 | Agent-Driven Self-Maintaining | DDL 监听 + 过期检测 + 闭环更新 |
| 3. 库内向量检索Linking | In-Database Vector Retrieval Linking | MariaDB HNSW + 两阶段检索 |
| 4. 端到端湖基融合 | End-to-End Integration | 全流程无外挂组件 |

### Demo 演示场景

1. **Onboarding**: 数据源接入 → 自动生成 Rich Context
2. **向量检索 Schema Linking**: 两阶段召回（粗筛 + 精排）
3. **Agent 自维持**: DDL 变更 → 自动更新 Context
4. **Text-to-SQL**: ReAct 推理 + 执行

## 技术选型

- **数据库**: MariaDB 11.4+（原生 VECTOR + HNSW 索引）
- **后端**: Go 1.24 + Gin
- **前端**: Vue 3 + Vite + UnoCSS + Naive UI
- **LLM**: OpenAI / 混元
- **论文**: LaTeX (已安装 texlive)

## 数据库连接

```bash
Host: 127.0.0.1
Port: 19010
User: lucid
Password: lucid2024
Database: lucid

# 快速连接
make db-login
```

## 当前项目状态

### 架构 ✅ 已完成

2026-02-04 完成全容器化重构：
- 统一使用 MariaDB 11.4+ (VECTOR + HNSW)
- 端口规划：19000 (前端), 19001 (后端), 19010 (数据库)
- 单一 docker-compose.yml 一键部署
- 双语文档 (EN/ZH)

### Go 后端 ✅ 已修复

补齐了缺失的包：
- `backend/config` - 配置加载和结构定义
- `backend/interfaces` - 核心接口定义 (DBAdapter, InferenceEngine 等)
- `backend/bridge` - 桥接内部实现与接口

后端已可正常编译 (`go build ./...`)。

### 前端 ✅ 正常

- Vue3 + Vite + UnoCSS + Naive UI
- 依赖已安装，类型检查通过

### 待完善功能

1. **推理引擎桥接** - `bridge/inference_bridge.go` 目前是占位实现，需要完善：
   - 完整的 ReAct 推理循环
   - LLM 客户端初始化（基于 llm_config.json）
   - 流式推理支持

2. **Grounding Service** - 需要完整的 LLM 模型才能初始化

3. **LaTeX 编译** - 缺少 `subcaption` 包：
   ```bash
   yum install texlive-subcaption
   ```

## 关键参考文档

VLDB Demo Track 论文在 `paper/` 目录：
- `paper/sections/` - 各章节 LaTeX 源文件
- `paper/main.tex` - 主文档
