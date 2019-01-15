.PHONY: vendor unit-test integration-test docker-up docker-down clear 

vendor:
	@go mod tidy

integration-test: docker-up vendor
	@go test -v ./...

unit-test:
	@go test -v -short ./...

docker-up:
	@docker-compose up -d

docker-down:
	@docker-compose down

clear: docker-down