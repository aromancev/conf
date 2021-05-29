TOOLS_IMG=confa/tools

.PHONY: tools
tools:
	docker build -f Dockerfile.tools -t ${TOOLS_IMG} .

.PHONY: gen
gen:
	docker run \
      --rm \
      -w /app \
      -v `pwd`:/app \
      ${TOOLS_IMG} go generate ./...
