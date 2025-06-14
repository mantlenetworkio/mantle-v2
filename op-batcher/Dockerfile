FROM --platform=$BUILDPLATFORM golang:1.19.9-alpine3.16 AS builder

ARG VERSION=v0.0.0

RUN apk add --no-cache make gcc musl-dev linux-headers git jq bash

# build op-batcher with the shared go.mod & go.sum files
COPY ./op-batcher /app/op-batcher
COPY ./op-bindings /app/op-bindings
COPY ./op-node /app/op-node
COPY ./op-service /app/op-service
COPY ./op-signer /app/op-signer
COPY ./datalayr /app/datalayr
COPY ./go.mod /app/go.mod
COPY ./go.sum /app/go.sum
COPY ./bss-core /app/bss-core

COPY ./op-batcher/docker.go.work /go/go.work

WORKDIR /app/op-batcher

RUN go mod tidy

RUN go mod download

ARG TARGETOS TARGETARCH

RUN make op-batcher VERSION="$VERSION" GOOS=$TARGETOS GOARCH=$TARGETARCH

FROM alpine:3.16

COPY --from=builder /app/op-batcher/bin/op-batcher /usr/local/bin

ENTRYPOINT ["op-batcher"]
