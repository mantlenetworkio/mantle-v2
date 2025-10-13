FROM --platform=$BUILDPLATFORM golang:1.24.7-alpine3.22 as builder

ARG VERSION=v0.0.0

RUN apk add --no-cache make gcc musl-dev linux-headers git jq bash

COPY ./gas-oracle /app/gas-oracle
COPY ./op-service /app/op-service
COPY ./go.mod /app/go.mod
COPY ./go.sum /app/go.sum
COPY ./.git /app/.git

WORKDIR /app/gas-oracle

RUN go mod download

ARG TARGETOS TARGETARCH

RUN make gas-oracle VERSION="$VERSION" GOOS=$TARGETOS GOARCH=$TARGETARCH

FROM alpine:3.21

RUN apk add --no-cache ca-certificates jq curl
COPY --from=builder /app/gas-oracle/bin/gas-oracle /usr/local/bin/

WORKDIR /usr/local/bin/
ENTRYPOINT ["gas-oracle"]
