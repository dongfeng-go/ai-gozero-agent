# 第一阶段:构建
FROM golang:1.26 AS builder

# 设置环境变量
ENV GO111MODULE=on
ENV GOPROXY=https://goproxy.cn,direct

# 设置工作目录
WORKDIR /build

# 复制项目文件
COPY . .

# 构建应用
WORKDIR /build/mcp
RUN CGO_ENABLED=0 GOOS=linux go build -o /output/mcp .


# 第二阶段:运行
FROM alpine:latest

# 创建配置目录
RUN mkdir -p /mcp/etc

# 复制配置文件和可执行文件
COPY --from=builder /build/mcp/etc/chat-docker.yml /app/mcp/chat.yml
COPY --from=builder /output/mcp /app/mcp

WORKDIR /mcp
CMD ["./mcp"]