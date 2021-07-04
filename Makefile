.PHONY: build
build:
	docker-compose -f docker-compose.yml build
	docker run \
		--rm \
		-w /app \
		-v `pwd`/web:/app \
		node:15.7.0 /bin/sh -c "npm install -g pnpm; pnpm install"

.PHONY: start
start:
	docker-compose -f docker-compose.yml up

.PHONY: start-api
start-api:
	docker-compose -f docker-compose.yml up nginx api beanstalkd postgres email

.PHONY: stop
stop:
	docker-compose down

.PHONY: migrate
migrate:
	cd api && \
	./migrate.sh migrate -m internal/confa/migrations -c internal/confa/migrations/tern.conf && \
	./migrate.sh migrate -m internal/user/migrations -c internal/user/migrations/tern.conf

.PHONY: test
test:
	cd api && \
    go test ./...

.PHONY: lint-api
lint-api:
	docker run \
	--rm \
	-w /app \
	-v `pwd`/api:/app \
	golangci/golangci-lint:v1.39-alpine golangci-lint run

.PHONY: lint-web
lint-web:
	docker run \
	--rm \
	-w /app \
	-v `pwd`/web:/app \
	node:15.7.0 npm run lint

.PHONY: cert-create
cert-create:
	docker run -it --rm -p 443:443 -p 80:80 --name certbot \
	  -v /etc/letsencrypt:/etc/letsencrypt          \
	  -v /var/log/letsencrypt:/var/log/letsencrypt  \
	  certbot/certbot certonly --standalone

.PHONY: cert-renew
cert-renew:
	docker run -it --rm -p 443:443 -p 80:80 --name certbot \
	  -v /etc/letsencrypt:/etc/letsencrypt          \
	  -v /var/log/letsencrypt:/var/log/letsencrypt  \
	  certbot/certbot renew

.PHONY: check
check:
	make test
	cd api && go fmt ./...
	make lint-api
	cd api \
	    && go build -o bin/ cmd/api/... \
	    && go build -o bin/ cmd/media/... \
	    && go build -o bin/ cmd/sfu/...
	echo DONE!
