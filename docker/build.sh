#!/bin/sh

go mod vendor

GO_IN=${GO_IN:-"./"}
GO_OUT=${GO_OUT:-"build/staticsrv"}

CURRENT_REVISION=$(git rev-parse --short --verify HEAD)
CURRENT_TAG=$(git describe --tags ${CURRENT_REVISION} 2>/dev/null)

go build \
    -o ${GO_OUT} \
    -mod=vendor \
    -ldflags "
        -X main.ref=${CURRENT_REVISION}
        -X main.tag=${CURRENT_TAG}
        -X github.com/svt/staticsrv.ref=${CURRENT_REVISION}
        -X github.com/svt/staticsrv.tag=${CURRENT_TAG}
    " \
    ${GO_IN}

upx -q ${GO_OUT}
