.PHONY: test
test:
	go test ./... -race -timeout 2m

.PHONY: lint
lint:
	docker run \
		--rm \
		-t \
		-w /app \
		-v `pwd`:/app \
		golangci/golangci-lint:v1.45-alpine golangci-lint run
