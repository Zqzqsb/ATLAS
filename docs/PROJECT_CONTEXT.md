# LUCID 项目上下文

> 本文档记录项目背景和开发上下文，供 Agent 快速理解项目状态

## 项目来源

LUCID 从 `ReActSQL` 硕士论文项目迁移而来（2024-02-04）。

原仓库路径：`/root/workspace/ReActSql/`
- `system/` 目录迁移至 `backend/` + `frontend/`
- `paper/conference/` 迁移至 `paper/`

原仓库保持不动，作为硕士论文归档。

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

- **数据库**: MariaDB 12（原生 VECTOR + HNSW 索引）
- **后端**: Go 1.24 + Gin
- **前端**: Vue 3 + Vite + UnoCSS + Naive UI
- **LLM**: OpenAI / 混元
- **论文**: LaTeX (已安装 texlive)

## 数据库连接

```bash
Host: 127.0.0.1
Port: 3310
User: root
Password: your_strong_password

# 快速连接
make db-login
```

## 当前待修复问题

### Go 编译问题

有几个 import 路径需要补充迁移：
- `lucid/bridge` - 需要从原仓库迁移或删除引用
- `lucid/config` - 需要从原仓库迁移或删除引用
- `lucid/interfaces` - 需要从原仓库迁移或删除引用

### LaTeX 编译问题

缺少 `subcaption` 包，需要安装：
```bash
yum install texlive-subcaption
```

## 关键参考文档

原仓库中的规划文档已部分整合到 CLAUDE.md，更多细节参考：
- 原 `plans/LUCID_overview.md` - 系统概览
- 原 `plans/final_reference/DEVELOPMENT_GUIDE.md` - 完整开发指南
- 原 `plans/final_reference/前端新架构.md` - 前端设计

## 迁移日期

2024-02-04
