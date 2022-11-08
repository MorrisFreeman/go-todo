.PHONY: help build build-local up down logs ps test
.DEFAULT_GOAL := HELP

DOCKER_TAG := latest
build: ## Build docker image to deploy
	docker build -t MorrisFreeman/go-todo:${DOCKER_TAG} \
		-- target deploy ./

build-local: ## Build docker image to local development
	docker dompose build --no-cache

up: ## Do docker compose up with hot reload
	docker compose up -deploy

down: ## Do docker compose down
	docker compose down

logs: ## Tail docker compose logs
	docker compose logs -f

ps: ## Check container status
	docker compose ps

test: ## Execute tests
	go test -race -shuffle=on ./...

help: ## Show options
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'