DOCKER_IMAGE="sverigestelevision/staticsrv"
.PHONY: all
build:
	@ docker build -t ${DOCKER_IMAGE} -f ./docker/build.dockerfile ./
clean:
	@ docker rmi ${DOCKER_IMAGE}