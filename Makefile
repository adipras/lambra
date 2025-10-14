.PHONY: help up down logs restart clean build migrate-up migrate-down backend-sh frontend-sh mysql-sh test

help:
	@echo "Lambra Platform - Available Commands:"
	@echo ""
	@echo "  make up           - Start all services"
	@echo "  make down         - Stop all services"
	@echo "  make logs         - View all logs"
	@echo "  make restart      - Restart all services"
	@echo "  make clean        - Stop and remove all containers and volumes"
	@echo "  make build        - Build all Docker images"
	@echo ""
	@echo "  make migrate-up   - Apply database migrations"
	@echo "  make migrate-down - Rollback database migrations"
	@echo ""
	@echo "  make backend-sh   - Open shell in backend container"
	@echo "  make frontend-sh  - Open shell in frontend container"
	@echo "  make mysql-sh     - Open MySQL shell"
	@echo ""
	@echo "  make test         - Run backend tests"

up:
	docker-compose up -d

down:
	docker-compose down

logs:
	docker-compose logs -f

restart:
	docker-compose restart

clean:
	docker-compose down -v
	docker system prune -f

build:
	docker-compose build

migrate-up:
	@echo "Applying all database migrations..."
	@for file in backend/migrations/*.up.sql; do \
		echo "Applying $$file..."; \
		docker cp $$file lambra-mysql:/tmp/migration.sql; \
		docker-compose exec mysql sh -c "mysql -ulambra -plambra_secret lambra_db < /tmp/migration.sql" || exit 1; \
	done
	@echo "All migrations applied successfully!"

migrate-down:
	@echo "Rolling back all database migrations..."
	@for file in $$(ls -r backend/migrations/*.down.sql); do \
		echo "Rolling back $$file..."; \
		docker cp $$file lambra-mysql:/tmp/migration.sql; \
		docker-compose exec mysql sh -c "mysql -ulambra -plambra_secret lambra_db < /tmp/migration.sql" || exit 1; \
	done
	@echo "All migrations rolled back successfully!"

backend-sh:
	docker-compose exec backend sh

frontend-sh:
	docker-compose exec frontend sh

mysql-sh:
	docker-compose exec mysql mysql -ulambra -plambra_secret lambra_db

test:
	docker-compose exec backend go test ./... -v

# Production commands
prod-up:
	docker-compose -f docker-compose.prod.yml up -d

prod-down:
	docker-compose -f docker-compose.prod.yml down

prod-logs:
	docker-compose -f docker-compose.prod.yml logs -f

prod-build:
	docker-compose -f docker-compose.prod.yml build
