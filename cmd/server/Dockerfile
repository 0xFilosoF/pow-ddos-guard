FROM golang:1.24.2-alpine AS builder
WORKDIR /app
RUN apk add --no-cache make git

COPY go.mod go.sum ./
RUN --mount=type=cache,id=go-pow-ddos-guard-mod,target=/root/go/pkg/mod \
  go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd/server/main.go

FROM alpine:3.17 AS prod
WORKDIR /app
RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /app/server /app/

RUN addgroup -S appuser && \
    adduser -S -G appuser -u 10001 appuser
USER appuser

COPY --from=builder /app/configs /app/configs

EXPOSE 8443
CMD ["/app/server"]
