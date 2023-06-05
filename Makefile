.PHONY: start
start:
	docker-compose --profile email up -V

.PHONY: migrate
migrate:
	./minio/mc.sh mb -p local/user-uploads local/user-public local/confa-tracks-internal local/confa-tracks-public
	./minio/mc.sh policy set download local/user-public
	./minio/mc.sh policy set download local/confa-tracks-public
	./mongo/init.sh
	./mongo/migrate.sh -source file://services/migrations/iam/ -database "mongodb://iam:iam@mongo:27017/iam?replicaSet=rs" up
	./mongo/migrate.sh -source file://services/migrations/rtc/ -database "mongodb://rtc:rtc@mongo:27017/rtc?replicaSet=rs" up
	./mongo/migrate.sh -source file://services/migrations/confa/ -database "mongodb://confa:confa@mongo:27017/confa?replicaSet=rs" up

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
	cd services && $(MAKE) test

.PHONY: lint
lint:
	cd services && $(MAKE) lint
	cd web && $(MAKE) lint

.PHONY: gen-services
gen-services:
	cd services && $(MAKE) gen

.PHONY: gen-web
gen-web:
	cd web && $(MAKE) gen

.PHONY: cert-create
cert-create:
	docker run -it --rm -p 443:443 -p 80:80 --name certbot \
	  -v /etc/letsencrypt:/etc/letsencrypt          \
	  -v /var/log/letsencrypt:/var/log/letsencrypt  \
	  certbot/certbot certonly --standalone

.PHONY: cert-renew
cert-renew:
	cd deploy && docker-compose -f api.docker-compose.yml down
	docker run -it --rm -p 443:443 -p 80:80 --name certbot \
	  -v /etc/letsencrypt:/etc/letsencrypt          \
	  -v /var/log/letsencrypt:/var/log/letsencrypt  \
	  certbot/certbot renew
	cd deploy && docker-compose -f api.docker-compose.yml up -d

.PHONY: build
build:
	cd services && $(MAKE) build
	cd web && $(MAKE) build

.PHONY: server-api
server-api:
	cd deploy-manual \
		&& docker-compose -f api.docker-compose.yml down \
		&& docker-compose -f api.docker-compose.yml build \
		&& docker-compose -f api.docker-compose.yml up -d

.PHONY: server-sfu
server-sfu:
	cd deploy-manual \
		&& docker-compose -f sfu.docker-compose.yml down \
		&& docker-compose -f sfu.docker-compose.yml build \
		&& docker-compose -f sfu.docker-compose.yml up -d
