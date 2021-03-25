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

.PHONY: lint
lint:
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
