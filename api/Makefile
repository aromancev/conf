TOOLS_IMG=confa/tools
MONGO_IMG=mongo:4.2

.PHONY: gen
gen:
	docker build -f Dockerfile.dev -t ${TOOLS_IMG} .
	find . -name "*.pb.go" -type f -delete
	find . -name "*.twirp.go" -type f -delete
	find . -name "*.gql.go" -type f -delete
	docker run \
		--rm \
		-w /app \
		-v `pwd`:/app \
		${TOOLS_IMG} go generate ./...

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
		golangci/golangci-lint:v1.41-alpine golangci-lint run

