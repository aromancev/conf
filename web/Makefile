TOOLS_IMG=confa/web-dev

.PHONY: gen
gen:
	docker build -f Dockerfile.dev -t ${TOOLS_IMG} .
	docker run \
		--rm \
		-it \
		--network=confa \
		-w /app/web \
		-v `pwd`:/app/web \
		${TOOLS_IMG} bash -c "apollo codegen:generate --target=typescript --outputFlat src/api/schema.ts \
			&& curl http://rtc/room/schema > /tmp/schema && quicktype -s schema -o src/api/room/schema.ts -t Message /tmp/schema"

.PHONY: modules
modules:
	docker run \
		--rm \
		-it \
		-w /app \
		-v `pwd`:/app \
		node:16.15 bash -c "npm install -g pnpm@7.1.0 && pnpm install"

.PHONY: lint
lint:
	docker build -f Dockerfile.dev -t ${TOOLS_IMG} .
	docker run \
		--rm \
		-it \
		-w /app/web \
		-v `pwd`:/app/web \
		-v /app/web/node_modules \
      	-v /app/web/.pnpm-store \
      	-v /app/web/dist \
		${TOOLS_IMG} npm run lint

.PHONY: build
build:
	docker build -f Dockerfile.dev -t ${TOOLS_IMG} .
	docker run \
		--rm \
		-it \
		-w /app/web \
		-v `pwd`:/app/web \
		-v /app/web/node_modules \
      	-v /app/web/.pnpm-store \
      	-v /app/web/dist \
		${TOOLS_IMG} npm run build
