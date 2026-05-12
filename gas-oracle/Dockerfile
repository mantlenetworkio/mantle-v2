FROM --platform=$BUILDPLATFORM golang:1.24.13-alpine3.22 as builder

ARG VERSION
ARG GIT_COMMIT
ARG GIT_DATE

RUN apk add --no-cache make gcc musl-dev linux-headers git jq bash just

COPY . /app

WORKDIR /app/gas-oracle

RUN go mod download

ARG TARGETOS TARGETARCH

RUN GITCOMMIT="${GIT_COMMIT:-$(git rev-parse HEAD 2>/dev/null || true)}" && \
    GITDATE="${GIT_DATE:-$(git show -s --format='%ct' 2>/dev/null || true)}" && \
    if [ -n "${VERSION:-}" ]; then \
      make gas-oracle VERSION="$VERSION" GITCOMMIT="$GITCOMMIT" GITDATE="$GITDATE" GOOS=$TARGETOS GOARCH=$TARGETARCH; \
    else \
      make gas-oracle GITCOMMIT="$GITCOMMIT" GITDATE="$GITDATE" GOOS=$TARGETOS GOARCH=$TARGETARCH; \
    fi

FROM alpine:3.22

RUN apk add --no-cache ca-certificates jq curl
COPY --from=builder /app/gas-oracle/bin/gas-oracle /usr/local/bin/

WORKDIR /usr/local/bin/
ENTRYPOINT ["gas-oracle"]
