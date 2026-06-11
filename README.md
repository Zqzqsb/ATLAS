# ATLAS

**A**daptive **T**ext-to-SQL with **L**ifecycle-**A**ware **S**elf-maintaining Context

> VLDB 2026 Demo Track

ATLAS co-locates schema metadata, semantic annotations, and vector embeddings entirely within a single RDBMS — no external vector store, no consistency issues, full ACID guarantees.

[![License: Apache 2.0](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?logo=docker&logoColor=white)](deploy/docker-compose.yml)
[![BIRD EX](https://img.shields.io/badge/BIRD_dev-76.40%25_EX-brightgreen)](#evaluation)

<p align="center">
  <img src="docs/images/demo_ui.png" alt="ATLAS Demo Interface" width="100%"/>
</p>
<p align="center"><em>(a) Forest-chunked onboarding on 517 tables &nbsp; (b) Two-stage adaptive query &nbsp; (c) Autonomous schema evolution</em></p>

## Innovations

### 1. Unified In-Database Storage

Schema, Rich Context, relationship graphs, vector embeddings (HNSW), and change audit logs all live in `rc_*` tables within a single MariaDB 12 instance. One SQL query combines vector similarity with relational filters.

### 2. Two-Stage Adaptive Schema Linking

- **Small schema** (≤30 tables): one-shot LLM linking.
- **Large schema** (>30 tables): vector retrieval narrows 500+ tables to ~20 candidates in sub-second time; LLM then refines to the truly relevant tables.

### 3. Rich Context Lifecycle

| Phase | Description |
|-------|-------------|
| **Onboarding** | ReAct agent samples data, generates descriptions/synonyms/business rules per column, embeds into HNSW |
| **Inference** | Vector retrieval injects relevant context into LLM prompt for disambiguation |
| **Evolution** | DDL changes detected → stale context marked → LLM regenerates → vectors re-embedded |

For large schemas, a **forest-based chunked** strategy decomposes the FK graph into connected subtrees for parallel processing.

### 4. Agent-Driven Self-Maintenance

Coordinator–executor architecture: DDL Detector diffs `information_schema` → Coordinator marks stale entries → Executor invokes LLM to regenerate → Change Logger records all modifications.

## Evaluation

**BIRD dev set** (1,534 questions, 11 databases):

| Configuration | EX (%) | Avg Iters |
|---|---|---|
| **Full ATLAS pipeline** | **76.40** | 3.37 |
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
  <img src="docs/images/architecture.png" alt="ATLAS Architecture" width="720"/>
</p>
<p align="center"><em>Three pipelines — Onboarding, Inference, Self-Maintenance — share unified in-database storage (rc_* tables).</em></p>

## Quick Start

```bash
git clone https://github.com/Zqzqsb/atlas.git
cd atlas

# 1. Configure environment
cp .env.example .env
# Edit .env: set LLM_API_KEY, EMBEDDING_API_KEY, EMBEDDING_BASE_URL, EMBEDDING_MODEL, etc.

# 2. Configure backend
cp backend/server/configs/system.yaml.example backend/server/configs/system.yaml
cp backend/server/configs/lakebase.yaml.example backend/server/configs/lakebase.yaml
# Edit yaml files or rely on environment variables

# 3. Configure LLM
cp llm_config.json.example llm_config.json
# Edit llm_config.json: set your model token and base_url

# 4. Launch
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
├── frontend/             # Vue 3 + Vite + UnoCSS + Naive UI
├── deploy/               # Docker Compose configs
├── docs/                 # Images and documentation
└── scripts/              # Utility scripts
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

[Apache License 2.0](LICENSE)

## Acknowledgments

- MariaDB Foundation for native VECTOR support
- [BIRD](https://bird-bench.github.io/) and [Spider](https://yale-lily.github.io/spider) benchmarks
