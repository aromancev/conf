.PHONY: start
start:
	docker-compose --profile email up -V

.PHONY: migrate
migrate:
	./mongo/init.sh
	./mongo/migrate.sh -source file://migrations/iam/ -database "mongodb://iam:iam@mongo:27017/iam?replicaSet=rs" up
	./mongo/migrate.sh -source file://migrations/rtc/ -database "mongodb://rtc:rtc@mongo:27017/rtc?replicaSet=rs" up
	./mongo/migrate.sh -source file://migrations/confa/ -database "mongodb://confa:confa@mongo:27017/confa?replicaSet=rs" up

.PHONY: mongosh
mongosh:
	docker run \
		--rm \
		-ti \
		--network="confa" \
		-v `pwd`/.artifacts/mongosh:/home/mongodb \
		mongo:4.2 mongo mongodb://mongo:mongo@mongo:27017/admin

.PHONY: test
test:
	cd api && $(MAKE) test

.PHONY: lint
lint:
	cd api && $(MAKE) lint
	cd web && $(MAKE) lint

.PHONY: gen
gen:
	cd api && $(MAKE) gen
	cd web && $(MAKE) gen

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

.PHONY: build
build:
	cd api \
	    && go build -o bin/ ./cmd/api/... \
	    && go build -o bin/ ./cmd/media/... \
	    && go build -o bin/ ./cmd/sfu/... \
		&& go build -o bin/ ./cmd/turn/...

.PHONY: check
check:
	make test
	cd api && go fmt ./...
	make lint
	make test
	make build
	echo DONE!

.PHONY: server-api
server-api:
	docker-compose -f deploy/api.docker-compose.yml down
	docker-compose -f deploy/api.docker-compose.yml build
	docker-compose -f deploy/api.docker-compose.yml --env-file deploy/.env up -d

.PHONY: server-sfu
server-sfu:
	docker-compose -f deploy/sfu.docker-compose.yml down
	docker-compose -f deploy/sfu.docker-compose.yml build
	docker-compose -f deploy/sfu.docker-compose.yml --env-file deploy/.env up -d
