# ✈️ 飞机订票系统 — 分布式数据库课程作业

基于 **TiDB** 分布式数据库的飞机订票系统，支持**普通用户**和**管理员**两种角色，演示分布式数据库架构在真实应用场景中的使用。

## 🏗️ 系统架构

```
┌─────────────────────────────────────────────────┐
│                   浏览器                         │
│            http://localhost:8080                 │
│     普通用户: /          管理员: /admin          │
└────────────────────┬────────────────────────────┘
                     │
┌────────────────────▼────────────────────────────┐
│            Go 后端 (Gin + REST API)              │
│    ┌──────────────────────────────────────┐     │
│    │  公开: /api/flights, /api/register    │     │
│    │  用户: /api/bookings, /api/profile    │     │
│    │  管理: /api/admin/flights|users|...   │     │
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
| 前端 | 原生 HTML/CSS/JS | 现代化 UI，响应式布局 |
| 后端 | Go + Gin | REST API，JWT 认证，角色权限 |
| 数据库 | TiDB (TiKV ×2 + PD) | 分布式 SQL，兼容 MySQL |
| 部署 | Docker Compose | 一键编排所有服务 |

## 👥 用户角色

### 普通用户
- 🔍 搜索航班（按出发地、目的地、日期）
- ✈️ 在线预订机票（事务 + 行锁防超卖）
- 📋 查看和管理个人订单
- 💳 模拟支付
- ❌ 取消未支付的订单

### 管理员
- ✈️ 航班的增删改查
- 👥 用户管理（查看、修改角色、删除）
- 📋 查看所有订单、强制取消
- 💰 查看支付记录
- 📊 后台数据统计面板

### 默认账号
| 角色 | 用户名 | 密码 |
|------|--------|------|
| 管理员 | `admin` | `admin123` |
| 普通用户 | 自行注册 | — |

## 🚀 快速开始

### 前置要求

- **Docker** ≥ 20.10
- **Docker Compose** ≥ 2.0
- 内存 ≥ 4GB（TiDB 集群需要约 2.5GB）

### 分享给组员（推荐方式）

```bash
# 1. 组员克隆仓库
git clone <仓库地址>
cd Database

# 2. 一键启动（组员只需这一条命令！）
./start.sh

# 3. 浏览器打开即可使用
# 普通用户: http://localhost:8080
# 管理后台: http://localhost:8080/admin  (admin / admin123)
```

首次启动会拉取 Docker 镜像并编译 Go 后端，需要几分钟。之后再次启动只需要几秒钟。

### 本地开发

要求额外安装 **Go** ≥ 1.26。

```bash
# 启动 TiDB 集群（Docker）+ 本地运行后端
./start.sh dev

# 或手动：
docker-compose up -d pd tikv1 tikv2 tidb
cd backend && go run .
```

## 📖 功能说明

### 前端页面

| 路径 | 功能 | 权限 |
|------|------|------|
| `/` | 主页：搜索航班 + 预订机票 + 我的订单 | 公开（预订需登录） |
| `/login` | 用户登录 | 公开 |
| `/register` | 用户注册 | 公开 |
| `/admin` | 管理后台：航班/用户/订单管理 | 仅管理员 |

### API 端点

#### 认证（公开）

| 方法 | 路径 | 说明 |
|------|------|------|
| `POST` | `/api/register` | 注册普通用户 |
| `POST` | `/api/login` | 登录，返回 JWT token + role |

#### 航班（公开）

| 方法 | 路径 | 说明 |
|------|------|------|
| `GET` | `/api/flights` | 查询航班（`?origin=&destination=&date=`） |
| `GET` | `/api/flights/:id` | 航班详情 |

#### 预订与支付（需认证）

| 方法 | 路径 | 说明 |
|------|------|------|
| `POST` | `/api/bookings` | 创建预订 |
| `GET` | `/api/bookings` | 我的订单列表 |
| `GET` | `/api/bookings/:booking_no` | 查看订单 |
| `DELETE` | `/api/bookings/:booking_no` | 取消订单 |
| `POST` | `/api/bookings/:booking_no/pay` | 模拟支付 |
| `GET` | `/api/profile` | 当前用户信息（含角色） |

#### 管理员 API（需认证 + admin 角色）

| 方法 | 路径 | 说明 |
|------|------|------|
| `GET` | `/api/admin/flights` | 所有航班（含已售罄） |
| `POST` | `/api/admin/flights` | 新增航班 |
| `PUT` | `/api/admin/flights/:id` | 修改航班 |
| `DELETE` | `/api/admin/flights/:id` | 删除航班 |
| `GET` | `/api/admin/users` | 用户列表 |
| `PUT` | `/api/admin/users/:id` | 修改用户角色 |
| `DELETE` | `/api/admin/users/:id` | 删除用户 |
| `GET` | `/api/admin/bookings` | 所有订单 |
| `GET` | `/api/admin/bookings/:booking_no` | 订单详情（含支付） |
| `DELETE` | `/api/admin/bookings/:booking_no` | 强制取消订单 |
| `GET` | `/api/admin/payments` | 支付记录 |

### 数据库表

| 表 | 说明 | 关键字段 |
|----|------|---------|
| `users` | 用户 | username, password_hash, role |
| `flights` | 航班 | flight_no, available_seats, price |
| `bookings` | 预订 | booking_no, user_id, flight_id, status |
| `payments` | 支付 | booking_id, user_id, amount, status |

### 预订流程

1. 用户注册 → 登录 → 获得 JWT token（24 小时有效）
2. 搜索航班 → 点击"预订" → 填写乘客信息
3. 系统使用 `SELECT ... FOR UPDATE` + 事务保证座位不超卖
4. 在"我的订单"中可查看订单、支付或取消

## 🧪 课程实验建议

### 实验一：环境搭建与基本 CRUD
- 使用 `./start.sh` 一键启动集群
- 通过 API 或 MySQL 客户端连接 TiDB
- 注册用户、查询航班、创建预订

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

### 实验四：角色权限验证
- 普通用户尝试访问 `/api/admin/users` → 返回 403
- 管理员登录 → 访问管理后台 → 可以增删改查
- 普通用户只能操作自己的订单，管理员可以查看/取消所有订单

### 实验五：性能对比
- 对比 TiDB（2 TiKV）vs 单机 MySQL 的查询性能
- 使用 `sysbench` 或自定义脚本压测

## 🛠️ 常用命令

```bash
./start.sh              # 一键启动（Docker 全容器化）
./start.sh dev          # 本地开发模式
./start.sh stop         # 停止所有服务
./start.sh restart      # 重启
./start.sh status       # 查看状态
./start.sh logs         # 查看所有容器日志
./start.sh logs backend # 仅查看后端日志
./start.sh reset        # 重置数据（危险操作）
```

## 📁 项目结构

```
Database/
├── docker-compose.yml       # 服务编排（TiDB 集群 + Go 后端）
├── start.sh                 # 一键启动脚本
├── README.md                # 本文件
├── .gitignore
├── .dockerignore
├── todo.txt                 # 需求文档
├── backend/
│   ├── Dockerfile           # Go 后端容器化
│   ├── go.mod / go.sum
│   ├── main.go              # 入口 + 路由
│   ├── config/config.go     # 配置（DSN、JWT密钥）
│   ├── database/
│   │   ├── db.go            # 数据库连接、建表、迁移
│   │   └── schema.sql       # DDL（4 张表）
│   ├── models/
│   │   ├── flight.go        # 航班查询 + 管理
│   │   ├── booking.go       # 预订（事务 + 行锁 + 所有权）
│   │   ├── user.go          # 用户注册/登录/JWT/角色
│   │   ├── payment.go       # 模拟支付
│   │   └── errors.go        # 业务错误定义
│   ├── handlers/
│   │   ├── flight.go        # 航班 API + 管理 API
│   │   ├── booking.go       # 预订 API + 支付 API + 管理 API
│   │   └── auth.go          # 认证 API + 用户管理 API
│   ├── middleware/
│   │   └── auth.go          # JWT 认证 + 管理员权限中间件
│   └── seed/seed.go         # 示例航班 + 默认管理员
└── frontend/
    ├── index.html           # 主页（航班搜索 + 预订 + 我的订单）
    ├── login.html           # 登录页
    ├── register.html        # 注册页
    ├── admin.html           # 管理后台（航班/用户/订单）
    └── style.css            # 全局现代化样式
```

## ⚠️ 注意事项

1. **内存不足**：TiKV 每个节点默认分配 1GB，确保 Docker 可用内存 ≥ 4GB
2. **端口冲突**：确保 4000（TiDB）、8080（后端）、2379（PD）端口未被占用
3. **首次启动慢**：Docker 拉取 TiDB 镜像可能需要几分钟，之后启动很快
4. **JWT 密钥**：生产环境请修改 `JWT_SECRET` 环境变量
5. **数据持久化**：TiDB 数据存储在 `data/` 目录，已在 `.gitignore` 中忽略
6. **管理员账号**：首次启动自动创建 `admin / admin123`，请尽快修改密码
7. **数据库迁移**：后端启动时自动检测并添加缺失的数据库列，兼容旧数据
