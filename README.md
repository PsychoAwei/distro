# ✈️ 飞机订票系统 — 分布式数据库课程作业

基于 **TiDB** 分布式数据库的飞机订票系统，演示分布式数据库架构在真实应用场景中的使用。

## 🏗️ 系统架构

```
┌─────────────────────────────────────────────────┐
│                   浏览器                         │
│            http://localhost:8080                 │
└────────────────────┬────────────────────────────┘
                     │
┌────────────────────▼────────────────────────────┐
│            Go 后端 (Gin + REST API)              │
│    ┌──────────────────────────────────────┐     │
│    │  /api/register   /api/login          │     │
│    │  /api/flights    /api/bookings       │     │
│    │  /static/*       (前端页面)           │     │
│    └──────────────────────────────────────┘     │
└────────────────────┬────────────────────────────┘
                     │ MySQL 协议 (4000)
┌────────────────────▼────────────────────────────┐
│              TiDB Server (SQL 层)                │
│         SQL 解析、优化、执行计划生成              │
└────────┬───────────────────────────┬────────────┘
         │                           │
┌────────▼──────┐           ┌───────▼───────┐
│  TiKV Node 1  │           │  TiKV Node 2  │
│  分布式 KV     │◄─────────►│  分布式 KV     │
│  (Raft 副本)   │  Raft 复制 │  (Raft 副本)   │
└────────┬──────┘           └───────┬───────┘
         │                          │
┌────────▼──────────────────────────▼──────────────┐
│            PD (Placement Driver)                  │
│           集群调度 · Region 分裂 · 负载均衡        │
└──────────────────────────────────────────────────┘
```

### 技术栈

| 层 | 技术 | 说明 |
|------|------|------|
| 前端 | 原生 HTML/CSS/JS | 登录、注册、航班搜索、预订 |
| 后端 | Go 1.26 + Gin | REST API，JWT 认证 |
| 数据库 | TiDB (TiKV ×2 + PD) | 分布式 SQL，兼容 MySQL |
| 部署 | Docker Compose | 一键编排所有服务 |

## 🚀 快速开始

### 前置要求

- **Docker** ≥ 20.10
- **Docker Compose** ≥ 2.0
- 内存 ≥ 4GB（TiDB 集群需要约 2.5GB）

### 方式一：Docker 全容器化（推荐）

```bash
# 1. 克隆项目
git clone <your-repo-url>
cd Database

# 2. 一键启动（或直接 docker-compose up -d）
chmod +x start.sh
./start.sh

# 3. 浏览器打开
# http://localhost:8080
```

首次启动会拉取镜像并编译 Go 后端，需要几分钟。

### 方式二：本地开发

要求额外安装 **Go** ≥ 1.26。

```bash
# 启动 TiDB 集群（Docker）
./start.sh dev

# 或手动：
docker-compose up -d pd tikv1 tikv2 tidb
cd backend && go run .
```

## 📖 功能说明

### 前端页面

| 路径 | 功能 |
|------|------|
| `/` | 主页：搜索航班 + 预订机票 |
| `/login` | 用户登录 |
| `/register` | 用户注册 |

### API 端点

#### 认证（公开）

| 方法 | 路径 | 说明 |
|------|------|------|
| `POST` | `/api/register` | 注册（`username` + `password`） |
| `POST` | `/api/login` | 登录，返回 JWT token |

#### 航班（公开）

| 方法 | 路径 | 说明 |
|------|------|------|
| `GET` | `/api/flights` | 查询航班（`?origin=&destination=&date=`） |
| `GET` | `/api/flights/:id` | 航班详情 |

#### 预订（需认证：`Authorization: Bearer <token>`）

| 方法 | 路径 | 说明 |
|------|------|------|
| `POST` | `/api/bookings` | 创建预订 |
| `GET` | `/api/bookings/:booking_no` | 查询预订 |
| `DELETE` | `/api/bookings/:booking_no` | 取消预订 |
| `GET` | `/api/profile` | 当前用户信息 |

### 预订流程

1. 用户注册 → 登录 → 获得 JWT token（24 小时有效）
2. 搜索航班 → 点击"预订" → 填写乘客信息
3. 系统使用 `SELECT ... FOR UPDATE` + 事务保证座位不超卖
4. 返回预订号，取消预订时自动释放座位

## 🧪 课程实验建议

### 实验一：环境搭建与基本 CRUD
- 使用 Docker Compose 启动集群
- 通过 API 或 MySQL 客户端连接 TiDB
- 执行查询航班、创建预订等基本操作

### 实验二：事务与并发控制
- 使用多个终端同时预订同一航班
- 观察 `FOR UPDATE` 行锁效果
- 验证座位不会超卖

### 实验三：分布式特性验证
```bash
# 1. 查看 Region 分布
curl http://localhost:2379/pd/api/v1/stores

# 2. 模拟节点故障
docker pause tidb-tikv1
# 观察系统是否仍可正常读写

# 3. 恢复节点
docker unpause tidb-tikv1

# 4. 查看 PD 调度日志
./start.sh logs pd
```

### 实验四：性能对比
- 对比 TiDB（2 TiKV）vs 单机 MySQL 的查询性能
- 使用 `sysbench` 或自定义脚本压测

## 🛠️ 常用命令

```bash
./start.sh              # 一键启动
./start.sh stop         # 停止所有服务
./start.sh restart      # 重启
./start.sh status       # 查看状态
./start.sh logs         # 查看所有容器日志
./start.sh logs backend # 仅查看后端日志
./start.sh reset        # 重置数据（危险操作）
./start.sh dev          # 本地开发模式
```

## 📁 项目结构

```
Database/
├── docker-compose.yml       # 服务编排（TiDB 集群 + Go 后端）
├── start.sh                 # 一键启动脚本
├── README.md                # 本文件
├── .gitignore
├── backend/
│   ├── Dockerfile           # Go 后端容器化
│   ├── go.mod / go.sum
│   ├── main.go              # 入口
│   ├── config/config.go     # 配置（DSN、JWT密钥）
│   ├── database/
│   │   ├── db.go            # 数据库连接与表初始化
│   │   └── schema.sql       # DDL（users、flights、bookings）
│   ├── models/
│   │   ├── flight.go        # 航班查询
│   │   ├── booking.go       # 预订（事务 + 行锁）
│   │   ├── user.go          # 用户注册/登录/JWT
│   │   └── errors.go        # 业务错误定义
│   ├── handlers/
│   │   ├── flight.go        # 航班 API
│   │   ├── booking.go       # 预订 API
│   │   └── auth.go          # 认证 API
│   ├── middleware/
│   │   └── auth.go          # JWT 认证中间件
│   └── seed/seed.go         # 示例航班数据
└── frontend/
    ├── index.html           # 主页（航班搜索 + 预订）
    ├── login.html           # 登录页
    ├── register.html        # 注册页
    └── style.css            # 全局样式
```

## ⚠️ 注意事项

1. **内存不足**：TiKV 每个节点默认分配 1GB，确保 Docker 可用内存 ≥ 4GB
2. **端口冲突**：确保 4000（TiDB）、8080（后端）、2379（PD）端口未被占用
3. **首次启动慢**：Docker 拉取 TiDB 镜像可能需要几分钟
4. **JWT 密钥**：生产环境请修改 `JWT_SECRET` 环境变量
5. **数据持久化**：TiDB 数据存储在 `data/` 目录，已在 `.gitignore` 中忽略
