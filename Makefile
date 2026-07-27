.PHONY: run build test vet fmt migrate-up migrate-down seed docker-up docker-down

run:
	go run ./cmd/api

build:
	go build -o ./bin/simpkl-api ./cmd/api

test:
	go test ./...

vet:
	go vet ./...

fmt:
	gofmt -w ./cmd ./internal

migrate-up:
	go run github.com/golang-migrate/migrate/v4/cmd/migrate -path migrations -database "$$DATABASE_URL" up

migrate-down:
	go run github.com/golang-migrate/migrate/v4/cmd/migrate -path migrations -database "$$DATABASE_URL" down 1

seed:
	go run ./cmd/seed

docker-up:
	docker compose up -d mysql

docker-down:
	docker compose down
