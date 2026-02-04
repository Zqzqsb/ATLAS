# LUCID - Lake-base Unified Context-aware Intelligence for Data
.PHONY: all dev backend frontend paper clean help

# ============== Development ==============
dev:
	docker-compose -f deploy/docker-compose.yml up

dev-build:
	docker-compose -f deploy/docker-compose.yml up --build

backend-dev:
	cd backend && go run ./server

frontend-dev:
	cd frontend && pnpm dev

# ============== Database ==============
db-up:
	docker-compose -f deploy/docker-compose.yml up mariadb -d

db-down:
	docker-compose -f deploy/docker-compose.yml down mariadb

db-login:
	mycli -h 127.0.0.1 -P 3310 -u root -pyour_strong_password

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
	cd backend && go build -o ../bin/lucid-server ./server

build-frontend:
	cd frontend && pnpm build

build: build-backend build-frontend

# ============== Docker ==============
docker-build:
	docker-compose -f deploy/docker-compose.yml build

docker-up:
	docker-compose -f deploy/docker-compose.yml up -d

docker-down:
	docker-compose -f deploy/docker-compose.yml down

docker-logs:
	docker-compose -f deploy/docker-compose.yml logs -f

# ============== Test ==============
test-backend:
	cd backend && go test ./...

# ============== Clean ==============
clean:
	cd paper && make clean
	rm -rf bin/
	rm -rf frontend/dist/

# ============== Help ==============
help:
	@echo "LUCID - Available Commands:"
	@echo ""
	@echo "Development:"
	@echo "  make dev            - Start full stack with Docker"
	@echo "  make backend-dev    - Run Go backend locally"
	@echo "  make frontend-dev   - Run Vue frontend locally"
	@echo ""
	@echo "Database:"
	@echo "  make db-up          - Start MariaDB container"
	@echo "  make db-login       - Connect to MariaDB"
	@echo ""
	@echo "Paper:"
	@echo "  make paper          - Build PDF"
	@echo "  make paper-watch    - Auto-rebuild on changes"
	@echo ""
	@echo "Build:"
	@echo "  make build          - Build all"
	@echo "  make docker-build   - Build Docker images"
