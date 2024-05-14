FROM golang:1.18.0-alpine3.15 as builder

RUN apk add --no-cache make musl-dev linux-headers gcc git jq bash

# build dl-node with local monorepo go modules
COPY ./dl-node /app/dl-node
COPY ./common /app/common
COPY ./lib /app/lib

WORKDIR /app/dl-node

RUN make dl-node


FROM alpine:3.15

COPY --from=builder /app/dl-node/bin/dl-node /usr/local/bin

ENTRYPOINT ["dl-node"]