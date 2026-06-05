#!/usr/bin/env bash
#
# 飞机订票系统 - 启动脚本
# 支持 Docker 全容器化模式和本地开发模式
#
# 用法:
#   ./start.sh              Docker 模式启动全部服务（推荐交给组员用这个）
#   ./start.sh dev          本地开发模式（仅 TiDB 用 Docker，后端用 go run）
#   ./start.sh stop         停止全部服务
#   ./start.sh restart      重启全部服务
#   ./start.sh reset        重置数据库数据
#   ./start.sh status       查看服务状态
#   ./start.sh logs         查看容器日志

set -e

PROJECT_DIR="$(cd "$(dirname "$0")" && pwd)"
BACKEND_DIR="$PROJECT_DIR/backend"

# 颜色
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info()  { echo -e "${GREEN}[INFO]${NC}  $1"; }
log_warn()  { echo -e "${YELLOW}[WARN]${NC}  $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }
log_step()  { echo -e "${BLUE}[STEP]${NC} $1"; }

# ============================================
# Docker 模式：全部容器化启动
# ============================================
start_docker() {
    echo ""
    echo -e "${BLUE}╔══════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║   ✈️  飞机订票系统 · Docker 模式     ║${NC}"
    echo -e "${BLUE}╚══════════════════════════════════════╝${NC}"
    echo ""

    cd "$PROJECT_DIR"

    log_step "构建并启动所有容器..."
    docker-compose up -d --build

    log_info "等待服务就绪..."
    for i in $(seq 1 90); do
        if curl -s http://localhost:8080/api/ping >/dev/null 2>&1; then
            echo ""
            log_info "全部服务就绪 ✓ (${i}秒)"
            echo ""
            echo -e "${GREEN}══════════════════════════════════════${NC}"
            echo -e "${GREEN}   启动成功！🎉${NC}"
            echo ""
            echo "   🌐 主页:      http://localhost:8080"
            echo "   🔑 登录:      http://localhost:8080/login"
            echo "   📝 注册:      http://localhost:8080/register"
            echo "   🔧 管理后台:  http://localhost:8080/admin"
            echo ""
            echo "   👤 默认管理员: admin / admin123"
            echo ""
            echo "   查看状态: ./start.sh status"
            echo "   查看日志: ./start.sh logs"
            echo -e "${GREEN}══════════════════════════════════════${NC}"
            return 0
        fi
        printf "\r  等待中... %d/90 秒" "$i"
        sleep 1
    done

    echo ""
    log_error "启动超时，请检查日志：./start.sh logs"
    exit 1
}

# ============================================
# 本地开发模式：TiDB 用 Docker，后端本地跑
# ============================================
start_dev() {
    echo ""
    echo -e "${BLUE}╔══════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║   ✈️  飞机订票系统 · 本地开发模式    ║${NC}"
    echo -e "${BLUE}╚══════════════════════════════════════╝${NC}"
    echo ""

    cd "$PROJECT_DIR"

    log_step "启动 TiDB 集群（不含后端容器）..."
    docker-compose up -d pd tikv1 tikv2 tidb

    log_info "等待 TiDB 端口开放..."
    for i in $(seq 1 60); do
        if timeout 1 bash -c 'echo > /dev/tcp/127.0.0.1/4000' 2>/dev/null; then
            log_info "TiDB 端口已开放 (${i}秒)，等待 MySQL 协议初始化..."
            break
        fi
        sleep 1
    done

    # TiDB 端口开放后 MySQL 协议还需要 10~15 秒初始化
    for i in $(seq 1 15); do
        sleep 1
        printf "\r  等待 MySQL 协议就绪... %d/15 秒" "$i"
    done
    echo ""
    log_info "准备启动后端"

    log_step "启动 Go 后端（本地）..."
    cd "$BACKEND_DIR"
    go run .
}

# ============================================
# 停止服务
# ============================================
stop_all() {
    log_step "停止所有服务..."
    cd "$PROJECT_DIR"
    docker-compose down
    # 清理本地后端进程
    if lsof -t -i:8080 >/dev/null 2>&1; then
        kill $(lsof -t -i:8080) 2>/dev/null || true
        log_info "已清理本地端口 8080"
    fi
    log_info "全部服务已停止"
}

# ============================================
# 查看状态
# ============================================
show_status() {
    echo ""
    echo -e "${BLUE}══════════════════════════════════════${NC}"
    echo -e "${BLUE}   ✈️  飞机订票系统 - 服务状态${NC}"
    echo -e "${BLUE}══════════════════════════════════════${NC}"

    echo ""
    echo "📦 容器状态："
    docker-compose -f "$PROJECT_DIR/docker-compose.yml" ps 2>/dev/null || echo "   无法读取容器状态"

    echo ""
    echo "🔧 后端服务 (8080)："
    if curl -s http://localhost:8080/api/ping >/dev/null 2>&1; then
        HEALTH=$(curl -s http://localhost:8080/api/ping)
        echo -e "   ${GREEN}运行中 ✓${NC}  $HEALTH"
    else
        echo -e "   ${RED}未运行 ✗${NC}"
    fi

    echo ""
    echo "🌐 前端页面："
    for page in "/" "/login" "/register" "/admin"; do
        STATUS=$(curl -s -o /dev/null -w "%{http_code}" "http://localhost:8080$page" 2>/dev/null || echo "---")
        if [ "$STATUS" = "200" ]; then
            echo -e "   ${GREEN}$STATUS${NC}  http://localhost:8080$page"
        else
            echo -e "   ${RED}$STATUS${NC}  http://localhost:8080$page"
        fi
    done

    echo ""
}

# ============================================
# 重置数据
# ============================================
reset_data() {
    log_warn "⚠  将删除所有数据（容器+数据卷+预订记录），确认？[y/N]"
    read -r answer
    if [ "$answer" != "y" ] && [ "$answer" != "Y" ]; then
        log_info "已取消"
        return
    fi

    cd "$PROJECT_DIR"

    log_step "停止并删除容器及数据卷..."
    docker-compose down -v

    log_step "清理本地数据目录..."
    if [ -d data/pd ]; then
        # 尝试直接删除，如果权限不足则用 docker 临时容器删除
        rm -rf data/pd data/tikv1 data/tikv2 2>/dev/null || \
            docker run --rm -v "$PROJECT_DIR/data:/data" alpine rm -rf /data/pd /data/tikv1 /data/tikv2 2>/dev/null || \
            { log_warn "无法自动清理 data/ 目录 (需 root 权限)，请手动: sudo rm -rf data/pd data/tikv1 data/tikv2"; }
    fi

    log_info "数据已重置"

    echo ""
    read -p "是否现在重新启动？[Y/n] " restart_answer
    if [ "$restart_answer" != "n" ] && [ "$restart_answer" != "N" ]; then
        start_docker
    fi
}

# ============================================
# 主入口
# ============================================
case "${1:-start}" in
    start)
        start_docker
        ;;

    dev)
        start_dev
        ;;

    stop)
        stop_all
        ;;

    restart)
        stop_all
        echo ""
        start_docker
        ;;

    reset)
        reset_data
        ;;

    status)
        show_status
        ;;

    logs)
        cd "$PROJECT_DIR"
        docker-compose logs -f "${2:-}"
        ;;

    *)
        echo "用法: ./start.sh [命令]"
        echo ""
        echo "命令:"
        echo "  start      Docker 全容器化启动（默认，推荐）"
        echo "  dev        本地开发模式（TiDB 容器 + 本地 go run）"
        echo "  stop       停止所有服务"
        echo "  restart    重启所有服务"
        echo "  reset      重置数据库并删除数据卷"
        echo "  status     查看服务状态"
        echo "  logs       查看容器日志"
        echo ""
        echo "示例:"
        echo "  ./start.sh              一键启动"
        echo "  ./start.sh dev          本地开发"
        echo "  ./start.sh logs backend 查看后端日志"
        exit 1
        ;;
esac
