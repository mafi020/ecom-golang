include .env

MIGRATION_PATH = ./migrations
PSQL_URL=postgres://$(PG_USER):$(PG_PASSWORD)@$(PG_HOST):$(PG_PORT)/$(PG_DB_NAME)?sslmode=disable
PROTO_DIR = ./api/proto
PROTO_GEN_OUT = ./internal/delivery/grpc/pb

.PHONY: start-server install-golang-migrate db-create migration-version migration migrate-up migrate-down migrate-reset migrate-force proto-clean proto-gen

start-server:
	air -c .air.toml


install-golang-migrate:
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

db-create:
	psql -U postgres -h localhost -c "CREATE DATABASE ecommerce_db;"

migration-version:
	migrate -database "$(PSQL_URL)" -path $(MIGRATION_PATH) version

migration:
# 	Expamle: make migration name=create_order_items_table
	migrate create -ext sql -dir $(MIGRATION_PATH) -seq $(name)

migrate-up:
	migrate -database "$(PSQL_URL)" -path $(MIGRATION_PATH) up

migrate-down:
	migrate -database "$(PSQL_URL)" -path $(MIGRATION_PATH) down

# Resets the database by running all down migrations, then all up migrations
migrate-reset:
	migrate -database "$(PSQL_URL)" -path $(MIGRATION_PATH) down -all
	migrate -database "$(PSQL_URL)" -path $(MIGRATION_PATH) up

# Use this if your migration fails and gets stuck (e.g., make migrate-force version=1)
migrate-force:
	migrate -database "$(PSQL_URL)" -path $(MIGRATION_PATH) force $(version)

# Removes previously auto-generated pb files to avoid ghost compilation artifacts
proto-clean:
	@echo "Cleaning old generated gRPC code..."
	@if [ -d "$(PROTO_GEN_OUT)" ]; then rm -rf $(PROTO_GEN_OUT)/*; fi

# Automatically creates output paths and compiles all .proto files in the folder
proto-gen: proto-clean
	@echo "Generating Go gRPC source files..."
	@mkdir -p $(PROTO_GEN_OUT)
	protoc --proto_path=$(PROTO_DIR) \
		-I=$(PROTO_DIR) \
		--go_out=. --go_opt=module=github.com/mafi020/ecom-golang \
		--go-grpc_out=. --go-grpc_opt=module=github.com/mafi020/ecom-golang \
		$(PROTO_DIR)/*.proto
	@echo "Successfully compiled Protobuf contracts!"

	

