.DEFAULT_GOAL := help

.PHONY: help
help:
	@echo "Makefile Commands:"
	@echo "----------------------------------------------------------------"
	@fgrep -h "##" $(MAKEFILE_LIST) | fgrep -v fgrep | sed -e 's/\\$$//' | sed -e 's/##//'
	@echo "----------------------------------------------------------------"

.PHONY: lint
lint:
	@go fmt ./...
	@go vet ./...

.PHONY: gen
gen: lint ## lint project
	@go generate ./...

test: gen ## run all tests
	@go test -v ./...

.PHONY: up
up: ## start containers(mongo)
	@docker-compose -f docker-compose.yml pull
	@docker-compose -f docker-compose.yml up -d

.PHONY: down
down: ## shuts docker containers(mongo)
	docker-compose -f docker-compose.yml down --remove-orphans