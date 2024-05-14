FROM golang:1.18.0-alpine3.15 as builder

RUN apk add --no-cache make musl-dev linux-headers gcc git jq bash

# build dl-disperser with local monorepo go modules
COPY ./dl-disperser /app/dl-disperser
COPY ./common /app/common
COPY ./lib /app/lib

WORKDIR /app/dl-disperser

RUN make dl-disperser


FROM alpine:3.15

COPY --from=builder /app/dl-disperser/bin/dl-disperser /usr/local/bin

ENTRYPOINT ["dl-disperser"]