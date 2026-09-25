include .env
export

COMPOSE ?= docker-compose
DB_SERVICE := tamagochi-postgres
FORWARDER_SERVICE := port-forwarder
MIGRATION_SERVICE := tamagochi-postgres-migrate

export PROJECT_ROOT := $(CURDIR)

.PHONY: env-up env-down env-dropdown migrate-create migrate-up migrate-down migrate-action

env-up:
	$(COMPOSE) up -d --wait $(DB_SERVICE) $(FORWARDER_SERVICE)

env-down:
	$(COMPOSE) down $(DB_SERVICE) $(FORWARDER_SERVICE)

env-dropdown:
	@read -p "Drop all database data? WARNING! [y/N]: " ans; \
	if [ "$$ans" = "y" ]; then \
		$(COMPOSE) down --remove-orphans; \
		rm -rf out/pgdata; \
		echo "Database data removed."; \
	else \
		echo "Action cancelled. Database data was kept."; \
	fi

migrate-create:
	@if [ -z "$(seq)" ]; then \
		echo "Usage: make migrate-create seq=<name>"; \
		exit 1; \
	fi
	$(COMPOSE) run --rm $(MIGRATION_SERVICE) create -ext sql -dir /migrations -seq "$(seq)"

migrate-up: env-up
	$(MAKE) migrate-action action=up

migrate-down: env-up
	$(MAKE) migrate-action action=down

migrate-action:
	@if [ -z "$(action)" ]; then \
		echo "Usage: make migrate-action action=up|down"; \
		exit 1; \
	fi
	@read -p "Run migration $(action)? [y/N]: " ans; \
	if [ "$$ans" = "y" ]; then \
		$(COMPOSE) run --rm --no-deps $(MIGRATION_SERVICE) \
			-path /migrations \
			-database "postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(DB_SERVICE):5432/$(POSTGRES_DB)?sslmode=disable" \
			"$(action)"; \
	else \
		echo "Migration cancelled."; \
	fi
