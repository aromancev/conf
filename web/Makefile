TOOLS_IMG=confa/web-dev

.PHONY: gen
gen:
	docker build -f Dockerfile.dev -t ${TOOLS_IMG} .
	rm -r .proto || true
	cp -r ../api/proto/api/ .proto
	docker run \
		--rm \
		-w /app \
		-v `pwd`/src:/app/src \
		-v `pwd`/.proto:/app/.proto \
		${TOOLS_IMG} pnpm run gen

.PHONY: modules
modules:
	docker run \
		--rm \
		-w /app \
		-v `pwd`:/app \
		${TOOLS_IMG} pnpm install
