FROM golang:1.25 AS builder

COPY . /src
WORKDIR /src

RUN GOPROXY=https://goproxy.cn go build \
    -ldflags "-X main.Version=$(git describe --tags --always 2>/dev/null || echo dev)" \
    -o /bin/op-faucet \
    ./op-faucet/cmd

FROM debian:stable-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
        ca-certificates \
        netbase \
        gettext-base \
    && rm -rf /var/lib/apt/lists/ \
    && apt-get autoremove -y \
    && apt-get autoclean -y

COPY --from=builder /bin/op-faucet /usr/local/bin/op-faucet

WORKDIR /app

EXPOSE 9000
EXPOSE 7300

ENTRYPOINT ["op-faucet"]
CMD ["--config=/app/configs/config.yaml", "--rpc.addr=0.0.0.0", "--rpc.port=9000"]
