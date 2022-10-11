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
		golangci/golangci-lint:v1.50-alpine golangci-lint run

.PHONY: build
build:
	go build -o bin/ ./cmd/iam/... \
		&& go build -o bin/ ./cmd/confa/... \
		&& go build -o bin/ ./cmd/rtc/... \
		&& go build -o bin/ ./cmd/gateway/... \
		&& go build -o bin/ ./cmd/sfu/... \
		&& go build -o bin/ ./cmd/turn/... \
		&& go build -o bin/ ./cmd/tracker/... \
		&& go build -o bin/ ./cmd/avp/...
