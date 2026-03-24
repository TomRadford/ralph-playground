IMAGE_NAME ?= office-foosball-stats
IMAGE_TAG ?= latest
PORT ?= 8080
DATA_DIR ?= $(CURDIR)/data
CONTAINER_NAME ?= office-foosball-stats

.PHONY: generate-api run docker-build docker-run docker-stop

generate-api:
	go generate ./internal/api

run:
	go run ./cmd/server

docker-build:
	docker build -t $(IMAGE_NAME):$(IMAGE_TAG) .

docker-run:
	mkdir -p "$(DATA_DIR)"
	docker run --rm \
		--name $(CONTAINER_NAME) \
		-p $(PORT):8080 \
		-v "$(DATA_DIR):/data" \
		$(IMAGE_NAME):$(IMAGE_TAG)

docker-stop:
	docker stop $(CONTAINER_NAME)
