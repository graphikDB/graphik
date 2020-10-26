.DEFAULT_GOAL := help

.PHONY: help
help:
	@echo "Makefile Commands:"
	@echo "----------------------------------------------------------------"
	@fgrep -h "##" $(MAKEFILE_LIST) | fgrep -v fgrep | sed -e 's/\\$$//' | sed -e 's/##//'
	@echo "----------------------------------------------------------------"

run:
	@go run main.go