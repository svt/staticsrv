DOCKER_IMAGE="sverigestelevision/staticsrv"
.PHONY: all
build:
	@ docker build -t ${DOCKER_IMAGE} -f ./docker/build.dockerfile ./
clean:
	@ docker rmi ${DOCKER_IMAGE}
start:
	go run ./... -addr=127.0.0.1:8080 -metrics-addr=127.0.0.1:9090 -enable-access-log -enable-metrics -config-variables=USER ./example
