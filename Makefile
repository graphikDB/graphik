version := "1.0.0"

.DEFAULT_GOAL := help

.PHONY: help
help:
	@echo "Makefile Commands:"
	@echo "----------------------------------------------------------------"
	@fgrep -h "##" $(MAKEFILE_LIST) | fgrep -v fgrep | sed -e 's/\\$$//' | sed -e 's/##//'
	@echo "----------------------------------------------------------------"

run:
	@go run cmd/graphik/main.go --open-id https://accounts.google.com/.well-known/openid-configuration

run-follower:
	@go run cmd/graphik/main.go \
	--open-id https://accounts.google.com/.well-known/openid-configuration \
	--listen-port 8080 \
	--join-raft localhost:7820 \
	--storage /tmp/graphik2

gen: proto gql

patch: ## bump sem version by 1 patch
	bumpversion patch --allow-dirty

minor: ## bump sem version by 1 minor
	bumpversion minor --allow-dirty

tag: ## tag the repo (remember to commit changes beforehand)
	git tag v$(version)

push:
	git push origin v$(version)

docker-build:
	@docker build -t graphikdb/graphik:v$(version) .

docker-push:
	@docker push graphikdb/graphik:v$(version)

.PHONY: proto
proto: ## regenerate gRPC code
	@echo "generating protobuf code..."
	@docker run -v `pwd`:/defs namely/prototool:latest generate
	@go fmt ./...

.PHONY: gql
gql: ## regenerate graphql code
	@gqlgen generate
	@graphdoc -s ./schema.graphql -o ./gen/gql/docs --force

release: ## build release binaries to ./bin
	@mkdir -p bin
	@gox -osarch="linux/amd64" -output="./bin/linux/{{.Dir}}"
	@gox -osarch="darwin/amd64" -output="./bin/darwin/{{.Dir}}"
	@gox -osarch="windows/amd64" -output="./bin/windows/{{.Dir}}"

.PHONY: up
up: ## start local containers
	@docker-compose -f docker-compose.yml pull
	@docker-compose -f docker-compose.yml up -d

.PHONY: down
down: ## shuts down local docker containers
	@docker-compose -f docker-compose.yml down --remove-orphans

build: ## build the server to ./bin
	@mkdir -p bin
	@cd cmd/graphik; gox -osarch="linux/amd64" -output="../../bin/linux/{{.Dir}}_linux_amd64"
	@cd cmd/graphik; gox -osarch="darwin/amd64" -output="../../bin/darwin/{{.Dir}}_darwin_amd64"
	@cd cmd/graphik; gox -osarch="windows/amd64" -output="../../bin/windows/{{.Dir}}_windows_amd64"
