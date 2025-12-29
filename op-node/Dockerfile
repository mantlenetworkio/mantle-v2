FROM --platform=$BUILDPLATFORM golang:1.24.10-alpine3.21 as builder

ARG VERSION=v0.0.0

RUN apk add --no-cache make gcc musl-dev linux-headers git jq bash just

# build op-node with the shared go.mod & go.sum files
COPY . /app

WORKDIR /app/op-node

RUN go mod download

ARG TARGETOS TARGETARCH

RUN make op-node VERSION="$VERSION" GOOS=$TARGETOS GOARCH=$TARGETARCH

FROM alpine:3.21

COPY --from=builder /app/op-node/bin/op-node /usr/local/bin

CMD ["op-node"]
