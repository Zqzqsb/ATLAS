# ATLAS

**A**daptive **T**ext-to-SQL with **L**ifecycle-**A**ware **S**elf-maintaining Context

> VLDB 2026 Demo Track

ATLAS is a self-contained Text-to-SQL system that co-locates schema metadata, semantic annotations, and vector embeddings entirely within a single RDBMS. Three Docker containers, one command, no external engines.

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?logo=docker&logoColor=white)](deploy/docker-compose.yml)
[![BIRD EX](https://img.shields.io/badge/BIRD_dev-75.55%25_EX-brightgreen)](#evaluation)

[English](README.md) | [简体中文](README.zh-CN.md)

<p align="center">
  <img src="paper/figures/demo_ui.png" alt="ATLAS Demo Interface" width="100%"/>
</p>
<p align="center"><em>(a) Forest-chunked onboarding on 517 tables &nbsp; (b) Two-stage adaptive query &nbsp; (c) Autonomous schema evolution</em></p>

## Four Innovations

### 1. Unified In-Database Storage

Schema, Rich Context, relationship graphs, vector embeddings (HNSW), and change audit logs all live in dedicated `rc_*` tables within a single MariaDB 12 instance. One SQL query combines vector similarity with relational filters — no external vector store, no consistency issues, full ACID guarantees.

### 2. Two-Stage Adaptive Schema Linking

- **Small schema** (≤30 tables): full schema goes directly to the LLM for one-shot linking.
- **Large schema** (>30 tables): vector retrieval narrows 500+ tables to ~20 candidates in sub-second time; LLM then refines to the truly relevant tables. The two stages run concurrently via an atomic slot, hiding retrieval latency entirely.

### 3. Rich Context Lifecycle

Rich Context is not static annotation — it flows through three phases:

| Phase | What happens |
|-------|-------------|
| **Onboarding** | ReAct agent samples data, generates descriptions/synonyms/business rules per column, embeds into HNSW |
| **Inference** | Vector retrieval injects relevant context into LLM prompt for disambiguation |
| **Evolution** | DDL changes detected → stale context marked → LLM regenerates → vectors re-embedded |

For large schemas (>30 tables), a **forest-based chunked** strategy decomposes the FK graph into connected subtrees for parallel agent processing.

### 4. Agent-Driven Self-Maintenance

A coordinator–executor architecture keeps context synchronized with live schema:

1. **DDL Detector** diffs `information_schema` against context tables
2. **Coordinator** marks affected entries as stale and plans tasks
3. **Executor** invokes LLM to regenerate descriptions and re-embed
4. **Change Logger** records all modifications with before/after diffs

## Evaluation

**BIRD dev set** (1,534 questions, 11 databases):

| Configuration | EX (%) | Avg Iters |
|---|---|---|
| **Full ATLAS pipeline** | **75.55** | 3.37 |
| − ReAct Loop (one-shot + RC) | 68.71 | 1.00 |
| − Business rules & value mappings | 72.04 | 3.62 |
| − Sample values & synonyms | 70.86 | 3.91 |
| Schema only (no Rich Context) | 65.45 | 4.49 |
| Baseline (direct generation) | 58.93 | 1.00 |

**System-level ablation** on TPC-H Enterprise (500+ tables, 30 cross-domain queries):

| Configuration | Recall@20 | EX (%) | Latency (s) |
|---|---|---|---|
| Full ATLAS pipeline | **93.3** | **70.0** | 4.8 |
| − Adaptive Linking | — (overflow) | — | timeout |
| − Vector Search | 66.7 | 50.0 | 5.6 |
| − ReAct Loop | 93.3 | 56.7 | 2.3 |
| − Rich Context | 80.0 | 53.3 | 4.9 |

> Detailed ablation results: [AtlasCore](https://github.com/Zqzqsb/AtlasCore)

## Architecture

<p align="center">
  <img src="paper/figures/architecture.png" alt="ATLAS Architecture" width="720"/>
</p>
<p align="center"><em>Three pipelines — Onboarding, Inference, Self-Maintenance — share unified in-database storage (rc_* tables).</em></p>

## Quick Start

```bash
git clone https://github.com/zqzqsb/atlas.git
cd atlas
docker compose -f deploy/docker-compose.yml up -d
```

Access the UI at **http://localhost:19000**

## Tech Stack

| Component | Technology |
|-----------|-----------|
| Database | MariaDB 12 (native VECTOR + HNSW) |
| Backend | Go 1.24 + Gin |
| Frontend | Vue 3 + Vite + UnoCSS + Naive UI |
| LLM | Any OpenAI-compatible API |
| Embedding | Any OpenAI-compatible embedding API |
| Deployment | Docker Compose (3 containers) |

## Usage

```bash
# Start all services
make up

# View logs
make logs

# Stop
make down

# Local development
make backend-dev    # Go backend
make frontend-dev   # Vue3 frontend
make db-up          # Database only
```

## Project Structure

```
atlas/
├── backend/              # Go backend
│   ├── internal/
│   │   ├── lakebase/         # Unified storage layer (rc_* tables)
│   │   ├── agent/            # Self-maintenance agent
│   │   ├── grounding/        # Schema linking
│   │   ├── inference/        # ReAct SQL generation
│   │   ├── embedding/        # Vector embedding
│   │   ├── context/          # Rich Context management
│   │   ├── adapter/          # Database adapters
│   │   └── llm/              # LLM client
│   └── server/               # HTTP API + SSE
├── frontend/             # Vue3 + Vite + UnoCSS + Naive UI
├── AtlasCore/            # Experiment framework (submodule)
├── paper/                # VLDB Demo paper (LaTeX)
├── deploy/               # Docker Compose configs
└── scripts/              # Demo video scripts
```

## Citation

```bibtex
@inproceedings{atlas2026vldb,
  title     = {ATLAS: Adaptive Text-to-SQL with Lifecycle-Aware Self-maintaining Context},
  author    = {Anonymous},
  booktitle = {Proceedings of the VLDB Endowment, Demo Track},
  year      = {2026}
}
```

## License

MIT License — see [LICENSE](LICENSE) for details.

## Acknowledgments

- MariaDB Foundation for native VECTOR support
- [BIRD](https://bird-bench.github.io/) and [Spider](https://yale-lily.github.io/spider) benchmarks
