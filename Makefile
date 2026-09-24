include .env
export

export PROJECT_ROOT=${shell pwd}
env-up:
	docker-compose up -d tamagochi-postgres port-forwarder


env-down:
	docker-compose down tamagochi-postgres

env-dropdown:
	@read -p "Drop all volumes? WARNING!!! [y/N] : " ans; \
	if [ "$$ans" = "y" ]; then \
	docker-compose down tamagochi-postgres && \
	rm -rf out/pgdata && \
	echo "Files were complietly down";\
	else \
		echo "Action cancelled. Data is still"; \
	fi
	@docker-compose down port-forwarder \
	echo "Port forwarder closed";

migrate-create:
	@if [ -z "$(seq)" ]; then \
		echo "check the input/call make migrate-create seq='value'"; \
		exit 1; \
	fi; \
	docker-compose run --rm tamagochi-postgres-migrate \
		create \
		-ext sql \
		-dir /migrations \
		-seq "$(seq)"
migrate-up:
	make migrate-action action=up

migrate-down:
	make migrate-action action=down	


migrate-action:

	@if [ -z "$(action)" ]; then \
		echo "check the input/call make migrate-action 'action=value'"; \
		exit 1; \
	fi; \

	@read -p "Do you want to make migration up/down [y/N] : " ans; \
	docker-compose run --rm tamagochi-postgres-migrate \
		-path /migrations \
		-database postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@tamagochi-postgres:5432/${POSTGRES_DB}?sslmode=disable \
		"$(action)"
