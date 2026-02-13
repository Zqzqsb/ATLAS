# LUCID - Lakebase-Unified Context-aware Intelligence for Data
# VLDB 2025/2026 Demo Track
#
# Ports: 19000 (frontend), 19001 (backend), 19010 (mariadb)

.PHONY: rebuild clean-build dev backend frontend paper clean help demo-up demo-down demo-reset collect-logs logs clean-logs

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
	docker compose -f deploy/docker-compose.yml build
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

# ============== Log Collection ==============
# Collect logs into logs/<timestamp>/ directory
logs:
	@bash scripts/collect-logs.sh

# Remove old log directories, keeping the 3 most recent
clean-logs:
	@echo "🧹 Cleaning old log directories (keeping latest 3)..."
	@cd logs && ls -dt */ 2>/dev/null | tail -n +4 | xargs rm -rf 2>/dev/null; \
	echo "✅ Done. Remaining:"; ls -dt */ 2>/dev/null | head -5 || echo "   (none)"
