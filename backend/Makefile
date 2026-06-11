# Self-contained containerized deployment
# NO external dependencies - everything is within /system directory
#
# Usage: make up | make down | make restart | make logs

.PHONY: all build up down restart restart-frontend restart-backend \
        build-frontend build-backend logs logs-frontend logs-backend \
        clean dev dev-frontend dev-backend status help test \
        rebuild rebuild-frontend rebuild-backend stop-dev go-build \
        demo-db-up demo-db-down demo-db-status demo-db-logs demo-db-reset

# ============================================================
# Configuration
# ============================================================
COMPOSE_FILE := docker-compose.yml
COMPOSE_CMD := docker compose -f $(COMPOSE_FILE)
DEMO_COMPOSE_FILE := docker/docker-compose.demo.yml
DEMO_COMPOSE_CMD := docker compose -f $(DEMO_COMPOSE_FILE)

# Colors for output
GREEN := \033[0;32m
YELLOW := \033[0;33m
RED := \033[0;31m
BLUE := \033[0;34m
NC := \033[0m # No Color

# ============================================================
# Demo Database Commands
# ============================================================

# Start demo databases
demo-db-up:
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "$(GREEN)🗄️  Starting Demo Databases...$(NC)"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@docker network create atlas_demo_network 2>/dev/null || true
	$(DEMO_COMPOSE_CMD) up -d
	@echo ""
	@echo "$(GREEN)✅ Demo databases started!$(NC)"
	@echo ""
	@echo "$(BLUE)📊 Available Databases:$(NC)"
	@echo "  car_1:     MySQL :15001  PostgreSQL :15002"
	@echo "  flight_2:  MySQL :15011  PostgreSQL :15012"
	@echo "  college_2: MySQL :15021  PostgreSQL :15022"
	@echo "  employee:  MySQL :15031  PostgreSQL :15032"
	@echo ""
	@echo "$(YELLOW)💡 Credentials: atlas / atlas123$(NC)"
	@$(MAKE) demo-db-status

# Stop demo databases
demo-db-down:
	@echo "$(RED)🛑 Stopping demo databases...$(NC)"
	$(DEMO_COMPOSE_CMD) down
	@echo "$(GREEN)✅ Demo databases stopped.$(NC)"

# Show demo database status
demo-db-status:
	@echo ""
	@echo "$(BLUE)📊 Demo Database Status:$(NC)"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@$(DEMO_COMPOSE_CMD) ps

# View demo database logs
demo-db-logs:
	$(DEMO_COMPOSE_CMD) logs -f

# Reset demo databases (remove volumes and restart)
demo-db-reset:
	@echo "$(RED)🔄 Resetting demo databases (this will delete all data)...$(NC)"
	$(DEMO_COMPOSE_CMD) down -v
	@docker network create atlas_demo_network 2>/dev/null || true
	$(DEMO_COMPOSE_CMD) up -d
	@echo "$(GREEN)✅ Demo databases reset and restarted.$(NC)"

# ============================================================
# Main Commands (Containerized)
# ============================================================

# Default target
all: help

# Start all services (containerized)
up:
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "$(GREEN)🚀 Starting atlas Demo System (Containers)...$(NC)"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	$(COMPOSE_CMD) up -d
	@echo ""
	@echo "$(GREEN)✅ Services started!$(NC)"
	@echo "   Frontend: http://localhost:8081"
	@echo "   Backend:  http://localhost:8082"
	@$(MAKE) status

# Stop all services
down:
	@echo "$(RED)🛑 Stopping all services...$(NC)"
	$(COMPOSE_CMD) down
	@echo "$(GREEN)✅ All container services stopped.$(NC)"

# Build all container images
build: build-backend build-frontend
	@echo "$(GREEN)✅ All images built successfully!$(NC)"

# Build frontend image only
build-frontend:
	@echo "$(BLUE)🔨 Building frontend image...$(NC)"
	$(COMPOSE_CMD) build frontend

# Build frontend static files (for local deployment)
build-frontend-static:
	@echo "$(BLUE)🔨 Building frontend static files...$(NC)"
	cd web-new && pnpm install && pnpm build
	@echo "$(GREEN)✅ Frontend built to web-new/dist/$(NC)"

# Build backend image only
build-backend:
	@echo "$(BLUE)🔨 Building backend image...$(NC)"
	$(COMPOSE_CMD) build backend

# ============================================================
# Restart Commands (Containerized)
# ============================================================

# Restart all services
restart:
	@echo "$(YELLOW)🔄 Restarting all services...$(NC)"
	$(COMPOSE_CMD) restart
	@$(MAKE) status

# Restart frontend only
restart-frontend:
	@echo "$(YELLOW)🔄 Restarting frontend...$(NC)"
	$(COMPOSE_CMD) restart frontend
	@echo "$(GREEN)✅ Frontend restarted.$(NC)"

# Restart backend only
restart-backend:
	@echo "$(YELLOW)🔄 Restarting backend...$(NC)"
	$(COMPOSE_CMD) restart backend
	@echo "$(GREEN)✅ Backend restarted.$(NC)"

# ============================================================
# Rebuild & Restart (for code changes)
# ============================================================

# Rebuild and restart all
rebuild:
	@echo "$(BLUE)🔨 Rebuilding and restarting all services...$(NC)"
	$(COMPOSE_CMD) build --no-cache
	$(COMPOSE_CMD) up -d --force-recreate
	@echo "$(GREEN)✅ All services rebuilt and restarted.$(NC)"

# Rebuild and restart frontend
rebuild-frontend:
	@echo "$(BLUE)🔨 Rebuilding and restarting frontend...$(NC)"
	$(COMPOSE_CMD) build --no-cache frontend
	$(COMPOSE_CMD) up -d --force-recreate frontend
	@echo "$(GREEN)✅ Frontend rebuilt and restarted.$(NC)"

# Rebuild and restart backend
rebuild-backend:
	@echo "$(BLUE)🔨 Rebuilding and restarting backend...$(NC)"
	$(COMPOSE_CMD) build --no-cache backend
	$(COMPOSE_CMD) up -d --force-recreate backend
	@echo "$(GREEN)✅ Backend rebuilt and restarted.$(NC)"

# ============================================================
# Logs (Containerized)
# ============================================================

# View all logs
logs:
	$(COMPOSE_CMD) logs -f

# View frontend logs only
logs-frontend:
	$(COMPOSE_CMD) logs -f frontend

# View backend logs only
logs-backend:
	$(COMPOSE_CMD) logs -f backend

# ============================================================
# Development Mode (Non-containerized, within /system)
# ============================================================

# Start local development (both frontend and backend)
dev: stop-dev
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "$(GREEN)🔧 Starting Development Mode (Local)...$(NC)"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@$(MAKE) -j2 dev-backend-bg dev-frontend-bg
	@echo ""
	@echo "$(GREEN)✅ Development services started!$(NC)"
	@echo "   Frontend: http://localhost:8081"
	@echo "   Backend:  http://localhost:8090"
	@echo ""
	@echo "   Use 'make logs-dev' to view logs"
	@echo "   Use 'make stop-dev' to stop"

# Start frontend dev server (foreground)
dev-frontend:
	@echo "$(BLUE)🔧 Starting frontend dev server...$(NC)"
	cd web-new && pnpm install && pnpm dev --host 0.0.0.0 --port 8081

# Start frontend dev server (background)
dev-frontend-bg:
	@echo "$(BLUE)🔧 Starting frontend dev server (background)...$(NC)"
	@mkdir -p logs
	@cd web-new && nohup pnpm dev --host 0.0.0.0 --port 8081 > ../logs/frontend.log 2>&1 &
	@echo "$(GREEN)✅ Frontend started on port 8081$(NC)"

# Start backend dev server (foreground)
dev-backend:
	@echo "$(BLUE)🔧 Starting backend dev server...$(NC)"
	cd server && ./server -config configs/system.yaml

# Start backend dev server (background)
dev-backend-bg: go-build
	@echo "$(BLUE)🔧 Starting backend dev server (background)...$(NC)"
	@mkdir -p logs
	@cd server && nohup ./server -config configs/system.yaml > ../logs/backend.log 2>&1 &
	@echo "$(GREEN)✅ Backend started on port 8090$(NC)"

# Build Go binary locally
go-build:
	@echo "$(BLUE)🔨 Building Go backend...$(NC)"
	@go build -o server/server ./server/
	@echo "$(GREEN)✅ Backend built: server/server$(NC)"

# Stop all dev processes
stop-dev:
	@echo "$(RED)🛑 Stopping development processes...$(NC)"
	@-pkill -f "pnpm dev.*8081" 2>/dev/null || true
	@-pkill -f "vite.*8081" 2>/dev/null || true
	@-pkill -f "./server.*system.yaml" 2>/dev/null || true
	@-pkill -f "server/server" 2>/dev/null || true
	@echo "$(GREEN)✅ Development processes stopped.$(NC)"

# View dev logs
logs-dev:
	@echo "$(BLUE)📋 Development Logs$(NC)"
	@echo "━━━━━━━━━━ Frontend ━━━━━━━━━━"
	@tail -50 logs/frontend.log 2>/dev/null || echo "No frontend logs"
	@echo ""
	@echo "━━━━━━━━━━ Backend ━━━━━━━━━━━"
	@tail -50 logs/backend.log 2>/dev/null || echo "No backend logs"

# Follow dev logs
logs-dev-follow:
	@tail -f logs/frontend.log logs/backend.log

# ============================================================
# Status & Testing
# ============================================================

# Show service status
status:
	@echo ""
	@echo "$(BLUE)📊 Service Status:$(NC)"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "$(YELLOW)Containers:$(NC)"
	@$(COMPOSE_CMD) ps 2>/dev/null || echo "  No containers running"
	@echo ""
	@echo "$(YELLOW)Local Processes:$(NC)"
	@ps aux | grep -E "(vite|pnpm.*8081|server.*system.yaml)" | grep -v grep || echo "  No local dev processes"
	@echo ""

# Test API endpoints
test:
	@echo "$(BLUE)🧪 Testing services...$(NC)"
	@echo ""
	@echo "$(YELLOW)Backend health (8082 container / 8090 local):$(NC)"
	@curl -s http://localhost:8082/health 2>/dev/null && echo "" || \
	 curl -s http://localhost:8090/health 2>/dev/null && echo "" || \
	 echo "$(RED)❌ Backend not responding$(NC)"
	@echo ""
	@echo "$(YELLOW)Frontend (8081):$(NC)"
	@curl -s -o /dev/null -w "HTTP %{http_code}\n" http://localhost:8081/ || echo "$(RED)❌ Frontend not responding$(NC)"
	@echo ""
	@echo "$(YELLOW)API Spider databases:$(NC)"
	@curl -s http://localhost:8081/api/v1/spider/databases 2>/dev/null | head -c 200 || \
	 echo "$(RED)❌ API not responding$(NC)"
	@echo ""

# ============================================================
# Cleanup
# ============================================================

# Remove containers and images
clean:
	@echo "$(RED)🧹 Cleaning up...$(NC)"
	$(COMPOSE_CMD) down --rmi local --volumes --remove-orphans 2>/dev/null || true
	@rm -rf logs/
	@echo "$(GREEN)✅ Cleanup complete.$(NC)"

# Deep clean (including all unused images)
clean-all: stop-dev
	@echo "$(RED)🧹 Deep cleaning...$(NC)"
	$(COMPOSE_CMD) down --rmi all --volumes --remove-orphans 2>/dev/null || true
	@docker image prune -f
	@rm -rf logs/ server/server
	@echo "$(GREEN)✅ Deep cleanup complete.$(NC)"

# ============================================================
# Data Management
# ============================================================

# Check data directories
check-data:
	@echo "$(BLUE)📁 Checking data directories...$(NC)"
	@echo ""
	@echo "Required directories:"
	@ls -la data/ 2>/dev/null || echo "$(RED)  ❌ data/ directory missing$(NC)"
	@echo ""
	@ls data/contexts/*.json 2>/dev/null | wc -l | xargs -I {} echo "  contexts/: {} files"
	@ls data/dbs/*.db 2>/dev/null | wc -l | xargs -I {} echo "  dbs/: {} databases"
	@ls data/spider_data/*.json 2>/dev/null | wc -l | xargs -I {} echo "  spider_data/: {} files"

# ============================================================
# Help
# ============================================================
help:
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "$(GREEN)atlas System - Self-Contained Deployment$(NC)"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@echo "$(YELLOW)🗄️ Demo Database Management:$(NC)"
	@echo "  make demo-db-up      - Start demo databases"
	@echo "  make demo-db-down    - Stop demo databases"
	@echo "  make demo-db-status  - Show demo database status"
	@echo "  make demo-db-logs    - View demo database logs"
	@echo "  make demo-db-reset   - Reset demo databases (delete data)"
	@echo ""
	@echo "$(YELLOW)🐳 Container Mode:$(NC)"
	@echo "  make up              - Start all containers"
	@echo "  make down            - Stop all containers"
	@echo "  make restart         - Restart all containers"
	@echo "  make build           - Build all images"
	@echo "  make rebuild         - Rebuild and restart all"
	@echo "  make logs            - View container logs"
	@echo ""
	@echo "$(YELLOW)🔧 Development Mode (Local):$(NC)"
	@echo "  make dev             - Start local dev (frontend + backend)"
	@echo "  make dev-frontend    - Start frontend only"
	@echo "  make dev-backend     - Start backend only"
	@echo "  make stop-dev        - Stop all dev processes"
	@echo "  make logs-dev        - View dev logs"
	@echo "  make go-build        - Build Go binary"
	@echo ""
	@echo "$(YELLOW)📊 Status & Testing:$(NC)"
	@echo "  make status          - Show all service status"
	@echo "  make test            - Test API endpoints"
	@echo "  make check-data      - Check data directories"
	@echo ""
	@echo "$(YELLOW)🧹 Cleanup:$(NC)"
	@echo "  make clean           - Remove containers + local images"
	@echo "  make clean-all       - Deep clean (all images + volumes)"
	@echo ""
	@echo "$(BLUE)📊 Access Points:$(NC)"
	@echo "  Container mode:  Frontend http://localhost:8081"
	@echo "                   Backend  http://localhost:8082"
	@echo "  Dev mode:        Frontend http://localhost:8081"
	@echo "                   Backend  http://localhost:8090"
