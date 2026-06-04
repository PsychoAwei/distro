# ===== 构建阶段 =====
FROM golang:alpine AS builder

WORKDIR /app

# 先复制依赖文件，利用 Docker 缓存
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# 复制后端源码并编译
COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o server .

# ===== 运行阶段 =====
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /app/server /server
COPY frontend/ /frontend/

ENV FRONTEND_DIR=/frontend

EXPOSE 8080

CMD ["/server"]
