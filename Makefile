# Make sure install migrate cli tool: go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Postgres environment variables
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_DB=codemaster
POSTGRES_USER=codemaster
POSTGRES_PASSWORD=codemaster


# Migrate commands

MIGRATE = migrate -path internal/migrations -database "postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable"

# (not real files)
.PHONY: migrate-up migrate-down migrate-version migrate-new

migrate-up:
	$(MIGRATE) up 1

migrate-down:
	$(MIGRATE) down 1

migrate-down-force:
	$(MIGRATE) force 1
	$(MIGRATE) down 1

migrate-version:
	$(MIGRATE) version

migrate-new:
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir internal/modules/core/migrations -seq $${name}


up-db:
	docker compose up -d postgres

down-db:
	docker compose down postgres
