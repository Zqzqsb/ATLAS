# LUCID - Lakebase-Unified Context-aware Intelligence for Data
# VLDB 2025/2026 Demo Track
#
# Ports: 19000 (frontend), 19001 (backend), 19010 (mariadb)

.PHONY: rebuild clean-build dev backend frontend paper clean help demo-up demo-down demo-reset

# ============== Primary Commands ==============
# Default target: idempotent build (first run or rebuild, preserves data)
.DEFAULT_GOAL := rebuild

rebuild:
	@echo "🔄 Building LUCID (preserving data)..."
	@# Stop and remove containers + images, keep volumes (data)
	-docker compose -f deploy/docker-compose.yml down --rmi local
	@# Clean local build artifacts
	rm -rf bin/ frontend/dist/ frontend/node_modules/.tmp/ frontend/.vite-cache/
	@# Rebuild from scratch with no cache
	docker compose -f deploy/docker-compose.yml build --no-cache
	@# Start services
	docker compose -f deploy/docker-compose.yml up -d
	@echo ""
	@echo "✅ LUCID is ready!"
	@echo "  Frontend: http://localhost:19000"
	@echo "  Backend:  http://localhost:19001"
	@echo "  MariaDB:  localhost:19010"
	@echo ""

clean-build:
	@echo "🧹 Clean build (removing ALL data and cache)..."
	@# Stop and remove everything including volumes
	-docker compose -f deploy/docker-compose.yml down -v --rmi local
	@# Clean local build artifacts
	rm -rf bin/ frontend/dist/ frontend/node_modules/.tmp/ frontend/.vite-cache/
	@echo "✅ Cleaned. Starting fresh build..."
	@# Rebuild from scratch
	docker compose -f deploy/docker-compose.yml build --no-cache
	@# Start MariaDB first (needs init time)
	docker compose -f deploy/docker-compose.yml up -d mariadb
	@echo "⏳ Waiting for MariaDB to initialize (30s)..."
	@sleep 30
	@# Start all services
	docker compose -f deploy/docker-compose.yml up -d
	@echo ""
	@echo "✅ LUCID is ready (fresh install)!"
	@echo "  Frontend: http://localhost:19000"
	@echo "  Backend:  http://localhost:19001"
	@echo "  MariaDB:  localhost:19010"
	@echo ""

# ============== Demo (One-Command Start) ==============
demo-up:
	@echo "🚀 Starting LUCID Demo System..."
	@echo ""
	docker compose -f deploy/docker-compose.yml up -d mariadb
	@echo "⏳ Waiting for MariaDB to initialize (30s)..."
	@sleep 30
	@echo "✅ MariaDB ready with:"
	@echo "   - Lake-Base storage (rc_* tables)"
	@echo "   - Spider tvshow database"
	@echo "   - Pre-generated Rich Context"
	@echo ""
	docker compose -f deploy/docker-compose.yml up -d
	@echo ""
	@echo "🎉 LUCID Demo System is ready!"
	@echo ""
	@echo "  Frontend:  http://localhost:19000"
	@echo "  Backend:   http://localhost:19001"
	@echo "  Database:  localhost:19010 (user: lucid, pass: lucid2024)"
	@echo ""
	@echo "Demo Database: spider_tvshow"
	@echo "  - TV_Channel (10 rows)"
	@echo "  - TV_series (10 rows)"
	@echo "  - Cartoon (10 rows)"
	@echo ""
	@echo "Try: make db-login-tvshow"
	@echo ""

demo-down:
	docker compose -f deploy/docker-compose.yml down -v
	@echo "✅ Demo system stopped and cleaned"

demo-reset:
	docker compose -f deploy/docker-compose.yml down -v
	docker compose -f deploy/docker-compose.yml up -d mariadb
	@echo "⏳ Waiting for MariaDB to reinitialize..."
	@sleep 30
	@echo "✅ Demo data reset complete"

demo-logs:
	docker compose -f deploy/docker-compose.yml logs -f

# ============== Quick Start ==============
up:
	docker compose -f deploy/docker-compose.yml up -d
	@echo ""
	@echo "LUCID is starting..."
	@echo "  Frontend: http://localhost:19000"
	@echo "  Backend:  http://localhost:19001"
	@echo ""

down:
	docker compose -f deploy/docker-compose.yml down

logs:
	docker compose -f deploy/docker-compose.yml logs -f

# ============== Development ==============
dev:
	docker compose -f deploy/docker-compose.yml up

dev-build:
	docker compose -f deploy/docker-compose.yml up --build --no-cache

backend-dev:
	cd backend && go run ./cmd/lucid-server -config configs/system.yaml

frontend-dev:
	cd frontend && pnpm dev --port 19000

# ============== Database ==============
db-up:
	docker compose -f deploy/docker-compose.yml up mariadb -d

db-down:
	docker compose -f deploy/docker-compose.yml down mariadb

db-login:
	mycli -h 127.0.0.1 -P 19010 -u lucid -plucid2024 lucid

db-login-tvshow:
	mycli -h 127.0.0.1 -P 19010 -u lucid -plucid2024 spider_tvshow

db-check:
	@echo "=== Lake-Base Tables ==="
	@mysql -h 127.0.0.1 -P 19010 -u lucid -plucid2024 -e "SELECT table_name, table_rows FROM information_schema.tables WHERE table_schema='lucid' AND table_name LIKE 'rc_%';"
	@echo ""
	@echo "=== Demo Datasources ==="
	@mysql -h 127.0.0.1 -P 19010 -u lucid -plucid2024 -e "SELECT id, name, db_type, db_name, status FROM lucid.rc_datasources;"
	@echo ""
	@echo "=== TVShow Tables ==="
	@mysql -h 127.0.0.1 -P 19010 -u lucid -plucid2024 -e "SHOW TABLES FROM spider_tvshow;"

# ============== Paper ==============
paper:
	cd paper && make

paper-watch:
	cd paper && make watch

paper-draft:
	cd paper && make draft

paper-clean:
	cd paper && make clean

# ============== Build ==============
build-backend:
	cd backend && go build -o ../bin/lucid-server ./cmd/lucid-server

build-frontend:
	cd frontend && pnpm build

build: build-backend build-frontend

# ============== Docker ==============
docker-build:
	docker compose -f deploy/docker-compose.yml build --no-cache

docker-up: up

docker-down: down

docker-logs: logs

docker-clean: clean-build

# ============== Test ==============
test-backend:
	cd backend && go test ./...

test-frontend:
	cd frontend && pnpm vue-tsc --noEmit

test: test-backend test-frontend

# ============== Clean ==============
clean:
	cd paper && make clean 2>/dev/null || true
	rm -rf bin/
	rm -rf frontend/dist/

# ============== Help ==============
help:
	@echo "LUCID - Lakebase-Unified Context-aware Intelligence for Data"
	@echo ""
	@echo "Primary Commands:"
	@echo "  make               - Build/rebuild (idempotent, preserves data)"
	@echo "  make rebuild       - Same as above"
	@echo "  make clean-build   - Clean ALL (data + cache) and rebuild fresh"
	@echo ""
	@echo "Quick Start:"
	@echo "  make up            - Start all services"
	@echo "  make down          - Stop all services"
	@echo "  make logs          - View service logs"
	@echo ""
	@echo "Development:"
	@echo "  make dev           - Start with Docker (foreground)"
	@echo "  make backend-dev   - Run Go backend locally"
	@echo "  make frontend-dev  - Run Vue frontend locally"
	@echo ""
	@echo "Database:"
	@echo "  make db-up         - Start database container"
	@echo "  make db-login      - Connect to Lake-Base (lucid)"
	@echo "  make db-login-tvshow - Connect to demo database"
	@echo "  make db-check      - Show database status"
	@echo ""
	@echo "Build:"
	@echo "  make build         - Build backend and frontend locally"
	@echo "  make docker-build  - Build Docker images"
	@echo ""
	@echo "Paper:"
	@echo "  make paper         - Build PDF"
	@echo "  make paper-watch   - Auto-rebuild on changes"
	@echo ""
	@echo "Ports:"
	@echo "  19000 - Frontend (Web UI)"
	@echo "  19001 - Backend (REST API)"
	@echo "  19010 - MariaDB (Lake-Base + Demo)"
