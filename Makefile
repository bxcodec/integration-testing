.PHONY: vendor unit-test integration-test docker-up docker-down clear 

vendor:
	@go get -v ./...

tidy: vendor
	@go mod tidy

integration-test: docker-up tidy
	@go test -v ./...

unit-test: tidy
	@go test -v -short ./...

docker-up:
	@docker-compose up -d

docker-down:
	@docker-compose down

clear: docker-downa