FROM --platform=$BUILDPLATFORM golang:1.24.13-alpine3.22 as builder

ARG VERSION
ARG GIT_COMMIT
ARG GIT_DATE

RUN apk add --no-cache make gcc musl-dev linux-headers git jq bash just

# build op-proposer with the shared go.mod & go.sum files
COPY . /app

WORKDIR /app/op-proposer

RUN go mod download

ARG TARGETOS TARGETARCH

RUN GITCOMMIT="${GIT_COMMIT:-$(git rev-parse HEAD 2>/dev/null || true)}" && \
    GITDATE="${GIT_DATE:-$(git show -s --format='%ct' 2>/dev/null || true)}" && \
    if [ -n "${VERSION:-}" ]; then \
      just GITCOMMIT="$GITCOMMIT" GITDATE="$GITDATE" VERSION="$VERSION" GOOS=$TARGETOS GOARCH=$TARGETARCH op-proposer; \
    else \
      just GITCOMMIT="$GITCOMMIT" GITDATE="$GITDATE" GOOS=$TARGETOS GOARCH=$TARGETARCH op-proposer; \
    fi

FROM alpine:3.22

COPY --from=builder /app/op-proposer/bin/op-proposer /usr/local/bin

CMD ["op-proposer"]
