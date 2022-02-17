IMG=confa/proto

.PHONY: gen
gen:
	docker build -t ${IMG} .
	docker run \
		--rm \
		-w /app/proto \
		-v `pwd`/../:/app \
		${IMG} ./gen.sh
