FROM golang:1.17-alpine
RUN apk update && \
    apk --no-cache add \
        ca-certificates \
        tzdata \
        git

ENV CGO_ENABLED "0"
ENV GO_OUT "/build/staticsrv"
ENV GO_IN "/repo"

COPY . ${GO_IN}
WORKDIR ${GO_IN}
RUN /repo/docker/build.sh

FROM alpine:latest
RUN apk update && \
    apk --no-cache add \
	ca-certificates \
	tzdata

WORKDIR /
COPY --from=0 /build/staticsrv /bin/staticsrv
RUN mkdir -p /srv/www

ONBUILD USER 1000:1000
ONBUILD WORKDIR /srv/www
ONBUILD EXPOSE 8080:8080
ONBUILD CMD ["/bin/staticsrv"]