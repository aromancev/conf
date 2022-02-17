TOOLS_IMG=confa/web-dev

.PHONY: gen
gen:
	docker build -f Dockerfile.dev -t ${TOOLS_IMG} .
	docker run \
		--rm \
		-w /app \
		--network=confa \
		-v `pwd`/src:/app/src \
		${TOOLS_IMG} pnpm run gen

.PHONY: modules
modules:
	docker build -f Dockerfile.dev -t ${TOOLS_IMG} .
	docker run \
		--rm \
		-w /app \
		-v `pwd`:/app \
		${TOOLS_IMG} pnpm install

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
