.PHONY: build
build:
	docker-compose build
	docker run \
		--rm \
		-w /app \
		-v `pwd`/web:/app \
		node:15.7.0 /bin/sh -c "npm install -g pnpm; pnpm install"

.PHONY: start-api
start-api:
	docker-compose up nginx api media beanstalkd postgres email

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
