OPENAPI_FILE := openapi.yaml

.PHONY: generate-api run

generate-api:
	go generate ./internal/api

run:
	go run ./cmd/server
