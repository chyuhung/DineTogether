FROM docker.1ms.run/library/golang:1.24-alpine AS builder

ARG GOPROXY=https://goproxy.cn,https://goproxy.io,direct
ENV GOPROXY=${GOPROXY}

WORKDIR /app

# ── 依赖（只随 go.mod/go.sum 变化而失效）──
COPY go.mod go.sum ./
RUN go mod download

# ── Go 源码 → 编译（只随 .go 文件变化而失效）──
COPY main.go .
COPY handlers/ handlers/
COPY middleware/ middleware/
COPY models/ models/
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o dinetogether .

# ── 静态资源与配置（不影响编译缓存）──
COPY templates/ templates/
COPY static/ static/
COPY config.yaml .
COPY schema.sql .

# ─────────────────────────────────────
FROM docker.1ms.run/library/alpine:3.20

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && \
    apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/dinetogether .
COPY --from=builder /app/config.yaml .
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/static ./static
COPY --from=builder /app/schema.sql .

EXPOSE 8080

VOLUME ["/app/data"]

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD wget -qO- http://localhost:8080/api/health || exit 1

CMD ["./dinetogether"]
