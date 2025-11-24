# Build Stage
FROM golang:1.25.4-alpine AS builder
WORKDIR /app
COPY . .
# 静态编译，不依赖系统库
RUN CGO_ENABLED=0 GOOS=linux go build -o gopher-cron cmd/main.go

# Run Stage (Distroless 更安全，或者用 alpine)
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/gopher-cron .
# 安装 timezone (调度系统必须要时间准确)
RUN apk add --no-cache tzdata
ENV TZ=Asia/Shanghai

CMD ["./gopher-cron"]
