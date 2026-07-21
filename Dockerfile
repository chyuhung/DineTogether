FROM golang:1.24-alpine AS builder

ARG GOPROXY=https://goproxy.cn,direct
ENV GOPROXY=${GOPROXY}

RUN echo "https://mirrors.aliyun.com/alpine/v3.20/main" > /etc/apk/repositories && \
    echo "https://mirrors.aliyun.com/alpine/v3.20/community" >> /etc/apk/repositories && \
    apk add --no-cache gcc musl-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 go build -ldflags="-s -w" -o dinetogether .

FROM alpine:3.20

RUN echo "https://mirrors.aliyun.com/alpine/v3.20/main" > /etc/apk/repositories && \
    echo "https://mirrors.aliyun.com/alpine/v3.20/community" >> /etc/apk/repositories && \
    apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/dinetogether .
COPY --from=builder /app/config.yaml .
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/static ./static
COPY --from=builder /app/schema.sql .

EXPOSE 8080

VOLUME ["/app/db"]

CMD ["./dinetogether"]
