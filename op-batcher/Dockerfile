FROM --platform=$BUILDPLATFORM golang:1.24.10-alpine3.21 as builder

ARG VERSION
ARG GIT_COMMIT
ARG GIT_DATE

RUN apk add --no-cache make gcc musl-dev linux-headers git jq bash just

# build op-batcher with the shared go.mod & go.sum files
COPY . /app

WORKDIR /app/op-batcher

RUN go mod download

ARG TARGETOS TARGETARCH

RUN GITCOMMIT="${GIT_COMMIT:-$(git rev-parse HEAD 2>/dev/null || true)}" && \
    GITDATE="${GIT_DATE:-$(git show -s --format='%ct' 2>/dev/null || true)}" && \
    if [ -n "${VERSION:-}" ]; then \
      just GITCOMMIT="$GITCOMMIT" GITDATE="$GITDATE" VERSION="$VERSION" GOOS=$TARGETOS GOARCH=$TARGETARCH op-batcher; \
    else \
      just GITCOMMIT="$GITCOMMIT" GITDATE="$GITDATE" GOOS=$TARGETOS GOARCH=$TARGETARCH op-batcher; \
    fi

FROM alpine:3.21

COPY --from=builder /app/op-batcher/bin/op-batcher /usr/local/bin

CMD ["op-batcher"]
