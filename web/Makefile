TOOLS_IMG=confa/web-dev

.PHONY: gen-local
gen-local:
	docker build -f Dockerfile.dev -t ${TOOLS_IMG} .
	docker run \
		--rm \
		-w /app \
		--network=confa \
		-v `pwd`/src:/app/src \
		${TOOLS_IMG} pnpm run gen:local

.PHONY: modules
modules:
	docker run \
		--rm \
		-w /app \
		-v `pwd`:/app \
		node:16.15 bash -c "npm install -g pnpm@7.1.0 && pnpm install"

.PHONY: lint
lint:
	docker build -f Dockerfile.dev -t ${TOOLS_IMG} .
	docker run \
		--rm \
		-t \
		-w /app \
		-v `pwd`/src:/app/src \
		-v `pwd`/.proto:/app/.proto \
		${TOOLS_IMG} pnpm run lint

.PHONY: build
build:
	docker build -f Dockerfile.dev -t ${TOOLS_IMG} .
	docker run \
		--rm \
		-t \
		-w /app \
		-v `pwd`/src:/app/src \
		-v `pwd`/dist:/app/dist \
		${TOOLS_IMG} pnpm run build
