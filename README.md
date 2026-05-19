# MXCMDB

MXCMDB 是一个轻量化、Agentless 的企业级基础设施管理与运维控制平台。

核心目标：

- 统一管理云资源与 IDC 资产
- 提供安全、可审计的运维入口
- 简化 SSH、用户、密钥、任务等日常运维操作
- 构建企业统一运维控制平面

许可证：AGPL-3.0

---

## Features

### 资产管理（CMDB）

- **云资源**：阿里云 ECS、VPC、安全组、SLB、RDS、Redis、OSS
- **IDC 资产**：物理服务器、虚拟机、网络设备、机柜、交换机、防火墙
- **资产标签**：env、team、project、owner、region 等自定义标签
- **资产归属**：部门、项目、负责人、业务、环境（dev/test/prod）
- **项目管理**：业务项目与资产关联，资产按项目维度聚合统计

### 云账号托管

- 阿里云 AccessKey 托管（AES-GCM 加密存储）
- 多账号管理与资源自动同步

### 资产探针（Agentless）

- 通过 SSH 远程采集：Hostname、CPU、Memory、Disk、Network、OS、Kernel、Docker、Running Services、SSH Users
- 支持并发采集、采集结果入库

### SSH Key 管理

- 公钥批量下发 / 删除 / 轮换
- 离职清理、authorized_keys 文件操作

### 用户管理

- **平台用户**：JWT 认证、角色（admin/operator/viewer）、MFA 双因素
- **Linux 用户**：远程创建 / 删除 / sudo / 锁定

### 批量命令执行

- Shell 命令批量下发，多主机并发
- 实时输出（WebSocket），执行记录与历史追踪

### 文件分发

- SCP/SFTP 批量文件下发
- 配置同步、权限控制

### Web Terminal

- xterm.js + WebSocket，浏览器内 SSH 终端
- 会话录制与回放审计

### 运维剧本

- 剧本编排（多步骤脚本串联）
- 内置剧本（系统信息采集、安全加固、内核优化、时间同步）
- 资产初始化一键执行

### 审批管理

- 危险操作审批工单
- 提交 / 审核 / 取消流程

### 审计日志

- 登录、命令、文件分发、SSH 操作、用户操作
- 全量记录，支持回溯

### IPAM

- 网段（CIDR）管理
- IP 分配 / 回收 / 占用统计、资产关联

---

## Tech Stack

| 层 | 技术 |
|---|---|
| 后端 | Go 1.22+, Gin, GORM |
| 前端 | React, Ant Design Pro (TypeScript) |
| 数据库 | PostgreSQL 15+ |
| 缓存 | Redis 7+ |
| SSH | golang.org/x/crypto/ssh |
| Web Terminal | xterm.js + WebSocket |

---

## 项目结构

```
mxcmdb/
├── cmd/server/             # 主程序入口
├── configs/                # 配置文件
│   ├── config.yaml         # 本地开发配置（gitignore）
│   └── config.example.yaml # 配置示例
├── internal/               # 内部包
│   ├── config/             # 配置加载（支持环境变量覆盖）
│   ├── middleware/          # Gin 中间件（JWT, Audit, CORS, RBAC）
│   ├── model/              # GORM 模型
│   ├── handler/            # HTTP Handler
│   ├── service/            # 业务逻辑层
│   ├── repository/         # 数据访问层
│   ├── router/             # 路由注册
│   ├── dto/                # 请求/响应结构体
│   └── pkg/                # 内部工具包
├── web/                    # 前端项目（React + Ant Design Pro）
├── deploy/                 # 部署文件
│   ├── Dockerfile          # 后端生产镜像（多阶段构建）
│   ├── dev/                # 开发环境（Air 热重载）
│   └── prod/               # 生产环境（nginx + API）
├── Makefile                # 常用命令
├── go.mod
└── go.sum
```

---

## 快速开始

### 环境要求

- Go >= 1.22
- Node.js >= 18
- PostgreSQL >= 15
- Redis >= 7
- Docker & Docker Compose
- Make

### 开发环境

本地需先运行 PostgreSQL 和 Redis（通过 Docker 或直接安装）。

```bash
# 复制配置文件并修改数据库连接
cp configs/config.example.yaml configs/config.yaml
vim configs/config.yaml

# Docker 方式启动开发环境（Air 热重载 + Vite HMR）
make dev-docker-up-d

# 查看日志
make dev-docker-logs
```

开发环境默认端口：
- 后端 API：`10030`
- 前端页面：`3800`

### 生产部署

```bash
# 1. 配置环境变量（数据库密码、JWT 密钥、加密密钥等）
cp deploy/prod/.env.example deploy/prod/.env
vim deploy/prod/.env

# 2. 构建并启动（PostgreSQL → Redis → API → Web 按依赖顺序启动）
make prod-docker-up-d

# 3. 查看日志
make prod-docker-logs

# 4. 停止
make prod-docker-down

# 5. 重启（重新构建镜像）
make prod-docker-restart
```

生产环境默认端口：
- Web 入口：`3800`（nginx 提供静态服务 + API/WebSocket 反向代理）
- API 直连：`10030`

生产配置说明：
- `deploy/prod/.env` — 所有敏感配置（密码、密钥），**不提交到版本控制**
- `deploy/prod/config.yaml` — 非敏感基线配置（release 模式、JSON 日志）
- 环境变量（`MXCMDB_*`）优先级高于 config.yaml，用于覆盖敏感值

### 常用命令

```bash
# 构建 & 检查
make build                  # 编译后端
make test                   # 运行测试
make lint                   # 代码检查（golangci-lint）

# 前端
cd web && npm install       # 安装依赖
cd web && npm run dev       # 启动开发服务
cd web && npm run build     # 生产构建

# Docker 开发环境
make dev-docker-up          # 前台启动
make dev-docker-up-d        # 后台启动
make dev-docker-down        # 停止
make dev-docker-logs        # 查看日志
make dev-docker-restart     # 重启

# Docker 生产环境
make prod-docker-up-d       # 后台启动
make prod-docker-down       # 停止
make prod-docker-logs       # 查看日志
make prod-docker-restart    # 重启
```

---

## API 规范

### 路由格式

```
/api/v1/{module}/{resource}
```

### 统一响应

```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

### 分页

```
GET /api/v1/assets?page=1&page_size=20&keyword=web&env=prod
```

---

## 架构分层

```
Router → Middleware → Handler → Service → Repository → DB
                                  ↓
                              SSH Executor（探针 / 命令执行 / Key 管理）
```

- **Handler**：参数校验、绑定 DTO、调用 Service、返回响应
- **Service**：业务逻辑、事务编排
- **Repository**：数据库 CRUD

---

## Performance Benchmark

针对 1000 / 10000 台主机场景的压力测试。

### 数据库层

| 测试项 | 数据 |
|---|---|
| GetByIDs vs GetByID (1000台) | 3.9ms vs 370ms, **94x** |
| 批量插入 vs 逐条 (1000条) | 11ms vs 568ms, **50x** |
| 10 任务并发写 10000 条 | 61ms, **162,520 TPS** |
| 5000 资产分页查询 | 首页 3ms, 末页 1.9ms |

### 应用层

| 测试项 | 数据 |
|---|---|
| 1000 台内存占用 | **2.7 KB/台** |
| 10000 台极限 (200并发) | **3932 台/秒**, 堆增长 10.5MB |
| 混合负载 (1000探针+500任务+20终端) | **2.0s**, 残留 0 goroutine |

### API 层

| 接口 | 并发 | QPS | P50 | P99 |
|---|---|---|---|---|
| 资产列表 | 50 | **1279** | 38ms | 74ms |
| Dashboard 统计 | 100 | 215 | 81ms | 157ms |

> 以上在 Docker + Air 热重载模式下执行，生产环境直接二进制部署预计提升 3-5x。

---

## License

AGPL-3.0
