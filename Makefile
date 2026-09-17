.PHONY: run fmt docker-up

run:
	go run ./cmd/server

fmt:
	gofmt -w cmd internal

docker-up:
	docker compose up --build