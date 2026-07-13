# Stage 1: Builder
FROM golang:1.26.1-alpine AS builder

ARG SERVICE

WORKDIR /app


COPY internal/common ./internal/common
COPY internal/${SERVICE}/go.mod internal/${SERVICE}/go.sum ./internal/${SERVICE}/
RUN go mod download -C internal/${SERVICE}

COPY internal/${SERVICE} ./internal/${SERVICE}

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -C internal/${SERVICE} \
    -ldflags="-w -s" \
    -o /app/server \
    ./cmd

# Stage 2: Runtime
FROM gcr.io/distroless/static-debian13:nonroot-amd64

COPY --from=builder /app/server /server

EXPOSE 4000
ENTRYPOINT ["/server"]
