FROM golang:1.18.0-alpine3.15 as builder

RUN apk add --no-cache make musl-dev linux-headers gcc git jq bash

# build dl-node with local monorepo go modules
COPY ./dl-retriever /app/dl-retriever
COPY ./common /app/common
COPY ./lib /app/lib

WORKDIR /app/dl-retriever

RUN make dl-retriever


FROM alpine:3.15

COPY --from=builder /app/dl-retriever/bin/dl-retriever /usr/local/bin

ENTRYPOINT ["dl-retriever"]