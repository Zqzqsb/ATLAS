# LUCID

**L**akebase-**U**nified **C**ontext-aware **I**ntelligence for **D**ata

A self-contained Text-to-SQL system with native vector search capabilities, designed for VLDB 2025/2026 Demo Track.

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?logo=docker&logoColor=white)](deploy/docker-compose.yml)

[English](README.md) | [简体中文](README.zh-CN.md)

## Quick Start

Deploy the entire stack with one command (Docker required):

```bash
git clone https://github.com/your-repo/lucid.git
cd lucid
docker compose -f deploy/docker-compose.yml up -d
```

Access the UI at **http://localhost:19000**

## Key Features

🗄️ **Lake-Base Native Storage**
- Schema, context, and vector embeddings in a single MariaDB instance
- No external vector databases (Milvus, Elasticsearch, etc.) required

🤖 **Agent Self-Maintaining**
- Automatic DDL change detection
- Closed-loop context updates
- Zero-maintenance rich context

🔍 **In-Database Vector Retrieval**
- MariaDB 12 native VECTOR + HNSW index
- Millisecond-level schema linking
- Two-stage retrieval (coarse + fine)

🔗 **End-to-End Integration**
- Single Docker Compose deployment
- No external dependencies
- Production-ready architecture

## Architecture

```
┌─────────────────────────────────────────────────────┐
│                 LUCID System                        │
├─────────────────────────────────────────────────────┤
│                                                     │
│  Frontend (Vue3)  ──→  Backend (Go)  ──→  MariaDB  │
│    :19000              :19001            :19010     │
│                                                     │
│  ┌──────────────────────────────────────────────┐  │
│  │ MariaDB 11.4 - Unified Storage (:19010)     │  │
│  │                                              │  │
│  │  Lake-Base (rc_* tables):                   │  │
│  │  ├─ rc_datasources  (Metadata)              │  │
│  │  ├─ rc_tables       (Rich Context)          │  │
│  │  ├─ rc_columns      (Rich Context)          │  │
│  │  ├─ rc_embeddings   (VECTOR + HNSW)         │  │
│  │  └─ rc_change_log   (Audit Trail)           │  │
│  │                                              │  │
│  │  Demo Databases:                             │  │
│  │  ├─ demo_ecommerce  (E-commerce)            │  │
│  │  └─ demo_tpch       (TPC-H Benchmark)       │  │
│  └──────────────────────────────────────────────┘  │
│                                                     │
└─────────────────────────────────────────────────────┘
```

## Tech Stack

| Component | Technology |
|-----------|-----------|
| Database | MariaDB 12 (VECTOR + HNSW) |
| Backend | Go 1.24 + Gin |
| Frontend | Vue 3 + Vite + UnoCSS + Naive UI |
| LLM | OpenAI / Hunyuan |
| Deployment | Docker + Docker Compose |

## Port Allocation

LUCID uses the **19xxx** port range to avoid conflicts:

| Service | Port | Description |
|---------|------|-------------|
| Frontend | 19000 | Web UI |
| Backend | 19001 | REST API + SSE |
| MariaDB | 19010 | Lake-Base + Demo Databases |

## Usage

### Start All Services

```bash
make up
# or
docker compose -f deploy/docker-compose.yml up -d
```

### View Logs

```bash
make logs
# or
docker compose -f deploy/docker-compose.yml logs -f
```

### Stop Services

```bash
make down
# or
docker compose -f deploy/docker-compose.yml down
```

### Local Development

```bash
# Backend (Go)
make backend-dev

# Frontend (Vue3)
make frontend-dev

# Database only
make db-up
```

## Configuration

Create a `.env` file to customize ports and API keys:

```bash
# Optional port overrides
LUCID_FRONTEND_PORT=19000
LUCID_BACKEND_PORT=19001
LUCID_MARIADB_PORT=19010

# LLM API Key
LLM_API_KEY=sk-your-api-key-here
```

## Database Connection

Connect to the Lake-Base storage:

```bash
# Using mycli
make db-login

# Or manually
mycli -h 127.0.0.1 -P 19010 -u lucid -plucid2024 lucid
```

## Documentation

- [Setup Guide](docs/SETUP.md) - Detailed deployment instructions
- [Architecture](docs/ARCHITECTURE.md) - System design and components
- [Development Guide](CLAUDE.md) - For contributors

## Research & Citation

LUCID is designed for **VLDB 2025/2026 Demo Track** submission. If you use this system in your research, please cite:

```bibtex
@inproceedings{lucid2026,
  title={LUCID: Lake-Base Unified Context-Aware Intelligence for Data},
  author={Your Name},
  booktitle={Proceedings of the VLDB Endowment},
  year={2026}
}
```

## Core Innovations

1. **Lake-Base Multi-Modal Storage** - Unified storage for schema, context, and vectors
2. **Agent Self-Maintaining** - Automatic context updates on DDL changes
3. **In-Database Vector Retrieval** - Native HNSW index for schema linking
4. **End-to-End Integration** - Zero external dependencies

## Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) for details.

## License

MIT License - see [LICENSE](LICENSE) for details.

## Acknowledgments

- MariaDB Foundation for VECTOR support
- Spider dataset for benchmarking
- VLDB community for inspiration

---

**Note**: This is a research prototype for VLDB Demo Track. For production use, please review security configurations and add proper authentication.
