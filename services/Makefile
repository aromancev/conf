IMG_PROTO_GEN=confa/proto-gen
IMG_RTC_GEN=confa/rtc-gen

.PHONY: test
test:
	go test ./... -race -timeout 5m

.PHONY: lint
lint:
	docker run \
		--rm \
		-it \
		-w /app \
		-v `pwd`:/app:delegated \
		golangci/golangci-lint:v1.51-alpine golangci-lint run

.PHONY: build
build:
	go build -o bin/ ./cmd/iam/... \
		&& go build -o bin/ ./cmd/confa/... \
		&& go build -o bin/ ./cmd/rtc/... \
		&& go build -o bin/ ./cmd/gateway/... \
		&& go build -o bin/ ./cmd/sfu/... \
		&& go build -o bin/ ./cmd/turn/... \
		&& go build -o bin/ ./cmd/tracker/... \
		&& go build -o bin/ ./cmd/avp/... \
		&& go build -o bin/ ./cmd/sender/...

.PHONY: gen
gen:
	docker build -t ${IMG_PROTO_GEN} --build-arg ARC=`uname -p` -f ./internal/proto/Dockerfile .
	docker run \
		--rm \
		-w /app \
		-v `pwd`/internal/proto:/app \
		${IMG_PROTO_GEN} ./gen.sh
	docker build -t ${IMG_RTC_GEN} -f ./cmd/rtc/web/Dockerfile .
	docker run \
		--rm \
		-w /app \
		-v `pwd`/cmd/rtc/web:/app \
		${IMG_RTC_GEN} ./gen.sh
