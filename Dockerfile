FROM --platform=linux/amd64 docker.m.daocloud.io/golang:1.24-alpine AS builder

ARG GOPROXY=https://goproxy.cn,https://goproxy.io,direct
ENV GOPROXY=${GOPROXY}

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o dinetogether .

FROM --platform=linux/amd64 docker.m.daocloud.io/alpine:3.20

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && \
    apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/dinetogether .
COPY --from=builder /app/config.yaml .
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/static ./static
COPY --from=builder /app/schema.sql .

EXPOSE 8080

VOLUME ["/app/db"]

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD wget -qO- http://localhost:8080/api/health || exit 1

CMD ["./dinetogether"]
