# 阶段 1：编译构建二进制文件
FROM golang:1.22-alpine AS builder
# 设置容器内的工作目录
WORKDIR /app
# 只复制依赖声明文件
COPY go.mod go.sum ./
# 下载依赖（这一步会被缓存）
RUN go mod download
# 复制源码并编译
COPY . .
# CGO_ENABLED=0 强制使用纯 Go 实现，生成一个纯静态二进制文件 GOOS 指定目标操作系统
# go build 构建命令 -o main 编译生成的二进制文件命名为main ./main.go 指定源码入口
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./main.go

# 轻量化运行环境
FROM alpine:latest
WORKDIR /app
# --from=builder 从前面定义的阶段提取 . 表示当前工作目录/app
COPY --from=builder /app/main .
# 声明监听端口 仅提示作用无实际操作
EXPOSE 8080
# docker run 时默认要执行的入口程序
CMD [ "./main" ]