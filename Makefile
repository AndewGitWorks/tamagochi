include .env
export


env-up:
	docker compose up -d tamagochi-postres




env-down:
	docker compose down tamagochi-postres



env-dropdown:
	@read -p "Drop all volumes? WARNING!!! [y/N]: " ans; \
	if ["$$ans" = "y"]; then \
	docker compose down tamagochi-postres && \
	rm -rf out/pgdata && \
	echo "Files were complietly down";\
	else \
		echo "Action cancelled. Data is still"; \
	fi