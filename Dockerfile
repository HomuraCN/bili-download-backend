# 使用最简 Linux 镜像
FROM alpine:latest

# 安装必要的基础库 (比如处理时区和HTTPS证书)
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# 1. 【核心优化】复制并赋予执行权限
# 这里的 --chmod=755 是 Docker 的原生特性，比 RUN chmod 更节省空间
# 即使 Action 环境偶尔丢失了 +x 权限，这一步也能强制补回来，确保 ./main 能运行
COPY --chmod=755 main .

# 2. 复制 Cookie 占位文件
# (Actions 里的 YAML 会先生成一个空的 cookie.json，防止 COPY 报错)
COPY cookie.json .

# 暴露端口 (对应你 Gin 的端口)
EXPOSE 9961

# 启动命令
CMD ["./main"]