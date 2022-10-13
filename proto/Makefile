IMG_GO=confa/proto-go
IMG_NODE=confa/proto-node

.PHONY: gen
gen:
	docker build -t ${IMG_GO} --build-arg ARC=`uname -p` -f Dockerfile .
	docker run \
		--rm \
		-w /app \
		-v ${PWD}:/app \
		${IMG_GO} ./go.sh
	# docker build -t ${IMG_NODE} -f Dockerfile.node .
	# docker run \
	# 	--rm \
	# 	-w /app/proto \
	# 	-v ${PWD}/../:/app \
	# 	${IMG_NODE} ./node.sh
