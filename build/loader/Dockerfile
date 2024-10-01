FROM --platform=$BUILDPLATFORM golang:1.22 AS builder
ARG TARGETARCH

WORKDIR /workspace
COPY . .

ENV GOCACHE=/root/gocache
RUN \
    --mount=type=cache,target=${GOCACHE} \
    --mount=type=cache,target=/go/pkg/mod \
    go mod download

RUN --mount=type=cache,target=${GOCACHE} \
    --mount=type=cache,id=model-manager,sharing=locked,target=/go/pkg/mod \
    CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} GO111MODULE=on make build-loader

FROM --platform=$BUILDPLATFORM python:3.12-slim
ARG TARGETARCH

WORKDIR /run

COPY --from=builder /workspace/bin/loader .

RUN pip install -U "huggingface_hub[cli]"

ENTRYPOINT ["./loader"]
