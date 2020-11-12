version := "0.0.0"

.DEFAULT_GOAL := help

.PHONY: help
help:
	@echo "Makefile Commands:"
	@echo "----------------------------------------------------------------"
	@fgrep -h "##" $(MAKEFILE_LIST) | fgrep -v fgrep | sed -e 's/\\$$//' | sed -e 's/##//'
	@echo "----------------------------------------------------------------"

run:
	@go run graphik/main.go  --auth.jwks=https://www.googleapis.com/oauth2/v3/certs

gen:
	@go generate ./...

patch: ## bump version by 1 patch
	bumpversion patch --allow-dirty

tag: ## tag the repo (remember to commit changes beforehand)
	git tag v$(version)

push:
	git push origin v$(version)

docker-build:
	@docker build -t colemanword/graphik:v$(version) .

docker-push:
	@docker push colemanword/graphik:v$(version)

.PHONY: proto
proto: ## regenerate gRPC code
	@echo "generating protobuf code..."
	@rm -rf gen
	@docker run -v `pwd`:/tmp colemanword/prototool:latest prototool generate
	@go fmt ./...