# ATLAS - Adaptive Text-to-SQL with Lifecycle-Aware Self-maintaining Context
# VLDB 2026 Demo Track
#
# Quick start for a newcomer:
#   make            # build + start everything (preserves data), then self-checks
#   make doctor     # diagnose config / containers / datasources
#   make clean-build # fresh cold start (DESTROYS data, asks for confirmation)
#
# Ports: 19000 (frontend), 19001 (backend), 19010 (mariadb)

.PHONY: rebuild clean-build check-config check-secrets wait-health verify-data \
        status doctor down logs clean-logs help

COMPOSE := docker compose -f deploy/docker-compose.yml
BACKEND_URL := http://localhost:19001

# ============== Colors ==============
RED    := \033[1;31m
YELLOW := \033[1;33m
GREEN  := \033[1;32m
CYAN   := \033[1;36m
NC     := \033[0m

.DEFAULT_GOAL := rebuild

# ============== Config Bootstrap ==============
# Ensure required config files exist (copy from .example if missing)
check-config:
	@if [ ! -f backend/server/configs/system.yaml ]; then \
		printf "$(CYAN)📋 Creating backend/server/configs/system.yaml from example...$(NC)\n"; \
		cp backend/server/configs/system.yaml.example backend/server/configs/system.yaml; \
	fi
	@if [ ! -f backend/server/configs/lakebase.yaml ]; then \
		printf "$(CYAN)📋 Creating backend/server/configs/lakebase.yaml from example...$(NC)\n"; \
		cp backend/server/configs/lakebase.yaml.example backend/server/configs/lakebase.yaml; \
	fi
	@if [ ! -f llm_config.json ]; then \
		printf "$(CYAN)📋 Creating llm_config.json from example...$(NC)\n"; \
		cp llm_config.json.example llm_config.json; \
	fi
	@if [ ! -f .env ]; then \
		printf "$(CYAN)📋 Creating .env from .env.example...$(NC)\n"; \
		cp .env.example .env; \
	fi

# ============== Secret / Placeholder Detection ==============
# Non-fatal: warn loudly (red) if keys are still placeholders or passwords are default.
check-secrets:
	@warn=0; \
	if grep -qE 'your-deepseek-api-key|your-qwen-api-key|YOUR_TOKEN_HERE|your-embedding-api-key' llm_config.json 2>/dev/null; then \
		printf "$(RED)⚠️  llm_config.json has placeholder keys — set real tokens for LLM + embedding.$(NC)\n"; \
		printf "$(YELLOW)    → Edit llm_config.json: fill 'token' fields and '_embedding.api_key'.$(NC)\n"; \
		warn=1; \
	fi; \
	if [ $$warn -eq 0 ]; then printf "$(GREEN)✅ API keys configured (LLM + embedding in llm_config.json).$(NC)\n"; fi; \
	printf "$(YELLOW)ℹ️  Demo uses default DB passwords (atlas2024). Change MARIADB_PASSWORD in .env for any non-local deployment.$(NC)\n"

# ============== Primary Commands ==============
# Default target: idempotent build (first run or rebuild). Preserves the data volume.
# On a brand-new machine (no volume) this performs a clean cold start automatically.
rebuild: check-config check-secrets
	@printf "$(CYAN)🔄 Building ATLAS (preserving data volume)...$(NC)\n"
	-$(COMPOSE) down --rmi local
	rm -rf bin/ frontend/dist/ frontend/node_modules/.tmp/ frontend/.vite-cache/
	$(COMPOSE) build
	$(COMPOSE) up -d
	@$(MAKE) --no-print-directory wait-health
	@$(MAKE) --no-print-directory verify-data
	@$(MAKE) --no-print-directory status

# Fresh cold start: wipes the MariaDB volume and re-initializes all demo databases.
# DESTRUCTIVE — asks for confirmation unless FORCE=1.
clean-build: check-config check-secrets
	@if [ "$(FORCE)" != "1" ]; then \
		printf "$(RED)⚠️  clean-build DELETES the MariaDB volume.$(NC)\n"; \
		printf "$(YELLOW)   The demo (5 datasources + Rich Context + embeddings) is RE-SEEDED from$(NC)\n"; \
		printf "$(YELLOW)   deploy/init/mariadb/01_atlas_demo.sql.gz, but any context you generated$(NC)\n"; \
		printf "$(YELLOW)   yourself since first start will be lost.$(NC)\n"; \
		printf "$(YELLOW)   Type 'yes' to continue (or re-run with FORCE=1): $(NC)"; \
		read ans; \
		if [ "$$ans" != "yes" ]; then printf "$(CYAN)Aborted. Use 'make rebuild' to preserve data.$(NC)\n"; exit 1; fi; \
	fi
	@printf "$(CYAN)🧹 Clean cold start (re-seeding demo from 01_atlas_demo.sql.gz)...$(NC)\n"
	-$(COMPOSE) down -v --rmi local
	rm -rf bin/ frontend/dist/ frontend/node_modules/.tmp/ frontend/.vite-cache/
	$(COMPOSE) build --no-cache
	$(COMPOSE) up -d mariadb
	@printf "$(CYAN)⏳ Seeding demo databases from dump + rebuilding vector index (~60s)...$(NC)\n"
	@sleep 45
	$(COMPOSE) up -d
	@$(MAKE) --no-print-directory wait-health
	@$(MAKE) --no-print-directory verify-data
	@$(MAKE) --no-print-directory status

# ============== Health / Verification ==============
# Poll backend /health until healthy or timeout (~120s).
wait-health:
	@printf "$(CYAN)⏳ Waiting for backend to become healthy...$(NC)\n"
	@for i in $$(seq 1 60); do \
		if curl -sf $(BACKEND_URL)/health >/dev/null 2>&1; then \
			printf "$(GREEN)✅ Backend healthy.$(NC)\n"; exit 0; \
		fi; \
		sleep 2; \
	done; \
	printf "$(RED)⚠️  Backend did not become healthy in time. Inspect: docker logs atlas-backend$(NC)\n"

# Check that demo datasources are actually connected (this is what the UI lists).
verify-data:
	@printf "$(CYAN)🔎 Checking demo datasources...$(NC)\n"
	@resp=$$(curl -s --max-time 10 $(BACKEND_URL)/api/v1/lakebase/datasources 2>/dev/null); \
	count=$$(printf '%s' "$$resp" | grep -o '"count":[0-9]*' | head -1 | grep -o '[0-9]*'); \
	if [ -n "$$count" ] && [ "$$count" -gt 0 ]; then \
		printf "$(GREEN)✅ %s demo databases connected and synced.$(NC)\n" "$$count"; \
	else \
		printf "$(RED)⚠️  No datasources available — Lake-Base is not connected.$(NC)\n"; \
		if docker logs atlas-backend 2>&1 | grep -q "Access denied"; then \
			printf "$(YELLOW)   Cause: DB password mismatch with an EXISTING volume (passwords only apply on first init).$(NC)\n"; \
			printf "$(YELLOW)   Fix: align MARIADB_PASSWORD in .env with the volume, OR 'make clean-build' (DESTROYS data).$(NC)\n"; \
		else \
			printf "$(YELLOW)   Check: docker logs atlas-backend  (often a missing embedding/LLM key).$(NC)\n"; \
		fi; \
	fi

# Print access URLs + container status.
status:
	@printf "\n$(GREEN)✅ ATLAS is up.$(NC)\n"
	@printf "  Frontend: $(CYAN)http://localhost:19000$(NC)\n"
	@printf "  Backend:  $(CYAN)http://localhost:19001$(NC)\n"
	@printf "  MariaDB:  $(CYAN)localhost:19010$(NC)\n\n"
	@$(COMPOSE) ps 2>/dev/null || true

# One-shot diagnostic for newcomers: config files, placeholder keys, containers, datasources.
doctor:
	@printf "$(CYAN)🩺 ATLAS doctor$(NC)\n"
	@printf "$(CYAN)── Config files ──$(NC)\n"
	@for f in backend/server/configs/system.yaml backend/server/configs/lakebase.yaml llm_config.json .env; do \
		if [ -f $$f ]; then printf "  $(GREEN)✓$(NC) %s\n" "$$f"; else printf "  $(RED)✗ missing$(NC) %s  (run 'make check-config')\n" "$$f"; fi; \
	done
	@printf "$(CYAN)── Secrets ──$(NC)\n"
	@$(MAKE) --no-print-directory check-secrets
	@printf "$(CYAN)── Containers ──$(NC)\n"
	@$(COMPOSE) ps 2>/dev/null || printf "  $(YELLOW)compose not running$(NC)\n"
	@printf "$(CYAN)── Datasources ──$(NC)\n"
	@$(MAKE) --no-print-directory verify-data

# ============== Lifecycle ==============
down:
	@$(COMPOSE) down

# ============== Log Collection ==============
logs:
	@bash scripts/collect-logs.sh

clean-logs:
	@printf "$(CYAN)🧹 Cleaning old log directories (keeping latest 3)...$(NC)\n"
	@cd logs && ls -dt */ 2>/dev/null | tail -n +4 | xargs rm -rf 2>/dev/null; \
	printf "$(GREEN)✅ Done. Remaining:$(NC)\n"; ls -dt */ 2>/dev/null | head -5 || echo "   (none)"

# ============== Help ==============
help:
	@printf "$(CYAN)ATLAS — make targets$(NC)\n"
	@printf "  $(GREEN)make$(NC) / $(GREEN)make rebuild$(NC)   Build + start (preserves data), then self-check\n"
	@printf "  $(GREEN)make clean-build$(NC)     Fresh cold start (DESTROYS data; FORCE=1 to skip prompt)\n"
	@printf "  $(GREEN)make doctor$(NC)          Diagnose config / containers / datasources\n"
	@printf "  $(GREEN)make status$(NC)          Show URLs + container status\n"
	@printf "  $(GREEN)make down$(NC)            Stop all containers\n"
	@printf "  $(GREEN)make logs$(NC)            Collect logs into logs/<timestamp>/\n"
