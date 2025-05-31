FROM golang:1.24-alpine3.21 AS builder

ARG MODE
ARG GIT_COMMIT_ID

ENV APP_REVISION=${GIT_COMMIT_ID:-"unknown"}

ENV HOME=/usr/src/app
ENV GO111MODULE=on
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV FRAMEWORK_PATH=./framework/snowdrop
ENV COMMON_MODULE_PATH=./common
ENV CONSOLE_MODULE_PATH=./console

WORKDIR ${HOME}
COPY ${FRAMEWORK_PATH}/go.mod ${FRAMEWORK_PATH}/go.sum ${FRAMEWORK_PATH}/
COPY ${COMMON_MODULE_PATH}/go.mod ${COMMON_MODULE_PATH}/go.sum ${COMMON_MODULE_PATH}/
COPY ${CONSOLE_MODULE_PATH}/go.mod ${CONSOLE_MODULE_PATH}/go.sum ${CONSOLE_MODULE_PATH}/

WORKDIR ${CONSOLE_MODULE_PATH}
RUN go mod download && go mod verify

WORKDIR ${HOME}
COPY ${FRAMEWORK_PATH} ${FRAMEWORK_PATH}/
COPY ${COMMON_MODULE_PATH} ${COMMON_MODULE_PATH}/
COPY ${CONSOLE_MODULE_PATH} ${CONSOLE_MODULE_PATH}/

WORKDIR ${CONSOLE_MODULE_PATH}
RUN CGO_ENABLED=0 go build \
    -ldflags="-s -w \
    -X 'main.AppVersion=$(cat VERSION)' \
    -X 'main.AppRevision=${APP_REVISION}'" \
    -o "/usr/bin/console" .

FROM alpine:3.22

RUN apk add --no-cache curl=8.14.1-r0

COPY --from=builder "/usr/bin/console" /usr/bin/console

ENTRYPOINT ["/usr/bin/console"]
