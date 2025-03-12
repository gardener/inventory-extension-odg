FROM golang:1.24.1 AS builder
ARG TARGETOS
ARG TARGETARCH

WORKDIR /workspace
COPY go.mod go.mod
COPY go.sum go.sum

# Cache deps before building and copying source so that we don't need to
# re-download as much, and so that source changes don't invalidate our
# dependency layer.
RUN go mod download

# Build
COPY cmd/ ./cmd
COPY pkg/ ./pkg
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} go build -a -o inventory-extension-odg ./cmd/inventory-extension-odg

# TODO: switch to distroless
#FROM gcr.io/distroless/static:nonroot
FROM alpine:3.21

RUN addgroup -S nonroot && adduser -S nonroot -G nonroot
WORKDIR /app
COPY --from=builder /workspace/inventory-extension-odg .
USER nonroot:nonroot

ENTRYPOINT ["/app/inventory-extension-odg"]
