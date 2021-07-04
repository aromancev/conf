TOOLS_IMG=confa/api-dev

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
