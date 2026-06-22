include .env
export

DATABASE_URL=postgres://$(PG_USER):$(PG_PASSWORD)@$(PG_HOST):$(PG_PORT)/$(PG_DATABASE)?sslmode=$(PG_SSL_MODE)

local-dev-run:
	@go run cmd/main.go

local-migrate-up:
	migrate -path migrations -database "$(DATABASE_URL)" up

local-migrate-down:
	migrate -path migrations -database "$(DATABASE_URL)" down

make-migration:
	@migrate create -ext sql -dir migrations -seq $(name)

test:
	go test ./... -v

deploy:
	docker compose up --build -d --remove-orphans

down:
	docker compose down

logs:
	docker compose logs -f

logs-web:
	docker compose logs -f web

restart:
	docker compose restart web
