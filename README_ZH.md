# NovaAPI

<div align="center">

[![Go](https://img.shields.io/badge/Go-1.25%2B-00ADD8.svg)](https://golang.org/)
[![Vue](https://img.shields.io/badge/Vue-3.4%2B-4FC08D.svg)](https://vuejs.org/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15%2B-336791.svg)](https://www.postgresql.org/)
[![Redis](https://img.shields.io/badge/Redis-7%2B-DC382D.svg)](https://redis.io/)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED.svg)](https://www.docker.com/)

**面向订阅额度分发的 AI API 网关平台**

[English](README.md) | 简体中文

</div>

## ⚠️ 重要声明

使用本项目前，请仔细阅读以下内容：

- **🚨 服务条款风险**：使用本项目可能违反 Anthropic 及其他上游服务商的服务条款。请在使用前查阅相关服务商的用户协议；由此产生的一切风险由使用者自行承担。
- **⚖️ 合规使用**：请仅在符合您所在国家或地区法律法规的前提下使用本项目，严禁任何非法用途。
- **📖 免责声明**：本项目仅供技术学习与研究使用。对于因使用本项目导致的账号封禁、服务中断、数据丢失或任何其他直接或间接损失，作者不承担任何责任。

## 项目简介

NovaAPI 是一个 AI API 网关平台，用于分发和管理来自 AI 产品订阅的 API 额度。用户通过平台生成的 API Key 访问上游 AI 服务，平台负责鉴权、计费、负载均衡和请求转发。

## 功能特性

- **多账号管理** —— 支持多种上游账号类型（OAuth、API Key）
- **API Key 分发** —— 为用户生成并管理 API Key
- **精准计费** —— Token 级别的用量追踪与成本核算
- **智能调度** —— 带粘性会话的智能账号选择
- **并发控制** —— 按用户、按账号的并发限制
- **限流** —— 可配置的请求与 Token 速率限制
- **内置支付系统** —— 支持易支付、支付宝、微信支付与 Stripe，用户可自助充值，无需单独部署支付服务（[配置指南](docs/PAYMENT.md)）
- **快捷登录** —— 支持 GitHub、Google、Linux.do 等 OAuth 第三方登录
- **管理后台** —— 用于监控与管理的 Web 界面
- **外部系统集成** —— 通过 iframe 嵌入外部系统（如工单系统）扩展管理后台

## 技术栈

| 组件 | 技术 |
|-----------|------------|
| 后端 | Go 1.25+、Gin、Ent |
| 前端 | Vue 3.4+、Vite 5+、TailwindCSS |
| 数据库 | PostgreSQL 15+ |
| 缓存/队列 | Redis 7+ |

---

## 快捷登录配置（OAuth）

NovaAPI 内置 GitHub、Google、Linux.do 第三方登录。在 **管理后台 → 系统设置 → OAuth/登录** 中开启并填写凭证即可，无需改动代码。

| 平台 | 申请地址 | 后端回调地址 |
|------|----------|--------------|
| **GitHub** | `github.com/settings/developers` → OAuth Apps → New | `https://你的域名/api/v1/auth/oauth/github/callback` |
| **Google** | `console.cloud.google.com/apis/credentials` → OAuth client ID（Web application） | `https://你的域名/api/v1/auth/oauth/google/callback` |
| **Linux.do** | `connect.linux.do` | `https://你的域名/api/v1/auth/oauth/linuxdo/callback` |

**通用要点：**
- 前端回跳地址统一填 `/auth/oauth/callback`
- scope 由后端自动携带（GitHub：`read:user user:email`；Google：`openid email profile`），无需手动填写
- 必须使用真实域名 + HTTPS，OAuth 回调地址需公网可达；`localhost` 无法完成真实跳转
- OAuth 应用登记的回调地址、管理后台填写的后端回调、实际访问域名三者必须完全一致，否则报 `redirect_uri_mismatch`
- 配置保存在数据库中，管理后台保存后即时生效，无需重启

---

## Nginx 反向代理注意事项

当使用 Nginx 作为 NovaAPI（或 CRS）配合 Codex CLI 的反向代理时，请在 Nginx 配置的 `http` 块中加入：

```nginx
underscores_in_headers on;
```

Nginx 默认会丢弃含下划线的请求头（如 `session_id`），这会破坏多账号场景下的粘性会话路由。

---

## 部署

> 以下命令中的 `your-org/novaapi` 为占位符，请替换为你自己的仓库地址。

### 方式一：脚本安装（推荐）

一键安装脚本，从 GitHub Releases 下载预编译二进制文件。

#### 前置要求

- Linux 服务器（amd64 或 arm64）
- PostgreSQL 15+（已安装并运行）
- Redis 7+（已安装并运行）
- Root 权限

#### 安装步骤

```bash
curl -sSL https://raw.githubusercontent.com/your-org/novaapi/main/deploy/install.sh | sudo bash
```

脚本将自动：
1. 检测系统架构
2. 下载最新版本
3. 安装二进制到 `/opt/novaapi`
4. 创建 systemd 服务
5. 配置系统用户与权限

#### 安装后操作

```bash
# 1. 启动服务
sudo systemctl start novaapi

# 2. 设置开机自启
sudo systemctl enable novaapi

# 3. 在浏览器中打开安装向导
# http://你的服务器IP:8080
```

安装向导将引导你完成：
- 数据库配置
- Redis 配置
- 管理员账号创建

#### 常用命令

```bash
# 查看状态
sudo systemctl status novaapi

# 查看日志
sudo journalctl -u novaapi -f

# 重启服务
sudo systemctl restart novaapi
```

---

### 方式二：Docker Compose（推荐）

使用 Docker Compose 部署，包含 PostgreSQL 与 Redis 容器。

#### 前置要求

- Docker 20.10+
- Docker Compose v2+

#### 手动部署

```bash
# 1. 克隆仓库
git clone https://github.com/your-org/novaapi.git
cd novaapi/deploy

# 2. 复制环境配置
cp .env.example .env

# 3. 编辑配置（生成安全密码）
nano .env
```

**`.env` 中的必填配置：**

```bash
# PostgreSQL 密码（必填）
POSTGRES_PASSWORD=your_secure_password_here

# JWT 密钥（推荐 —— 重启后保持用户登录态）
JWT_SECRET=your_jwt_secret_here

# TOTP 加密密钥（推荐 —— 重启后保留 2FA）
TOTP_ENCRYPTION_KEY=your_totp_key_here

# 可选：管理员账号
ADMIN_EMAIL=admin@example.com
ADMIN_PASSWORD=your_admin_password

# 可选：自定义端口
SERVER_PORT=8080
```

**生成安全密钥：**

```bash
openssl rand -hex 32   # 分别用于 JWT_SECRET / TOTP_ENCRYPTION_KEY / POSTGRES_PASSWORD
```

```bash
# 4. 创建数据目录（本地目录版）
mkdir -p data postgres_data redis_data

# 5. 启动所有服务
docker compose -f docker-compose.local.yml up -d

# 6. 查看状态
docker compose -f docker-compose.local.yml ps

# 7. 查看日志
docker compose -f docker-compose.local.yml logs -f novaapi
```

#### 部署版本对比

| 版本 | 数据存储 | 迁移 | 适用场景 |
|---------|-------------|-----------|----------|
| **docker-compose.local.yml** | 本地目录 | ✅ 简单（tar 打包整个目录） | 生产环境、频繁备份 |
| **docker-compose.yml** | 命名卷 | ⚠️ 需要 docker 命令 | 简单部署 |

**推荐**：使用 `docker-compose.local.yml`，数据管理更方便。

#### 访问

在浏览器中打开 `http://你的服务器IP:8080`。

若管理员密码为自动生成，可在日志中查找：

```bash
docker compose -f docker-compose.local.yml logs novaapi | grep "admin password"
```

#### 升级

```bash
docker compose -f docker-compose.local.yml pull
docker compose -f docker-compose.local.yml up -d
```

#### 便捷迁移（本地目录版）

使用 `docker-compose.local.yml` 时，可轻松迁移到新服务器：

```bash
# 源服务器
docker compose -f docker-compose.local.yml down
cd ..
tar czf novaapi-complete.tar.gz novaapi-deploy/

# 传输到新服务器
scp novaapi-complete.tar.gz user@new-server:/path/

# 新服务器
tar xzf novaapi-complete.tar.gz
cd novaapi-deploy/
docker compose -f docker-compose.local.yml up -d
```

---

### 方式三：从源码构建

用于开发或定制。

#### 前置要求

- Go 1.21+
- Node.js 18+

#### 开发模式

```bash
# 后端（热重载）
cd backend
go run ./cmd/server

# 前端（热重载）
cd frontend
pnpm install
pnpm run dev
```

#### 代码生成

修改 `backend/ent/schema` 后，需重新生成 Ent + Wire：

```bash
cd backend
go generate ./ent
go generate ./cmd/server
```

---

## Simple 模式

Simple 模式面向个人开发者或内部团队，无需完整 SaaS 功能即可快速使用。

- 启用：设置环境变量 `RUN_MODE=simple`
- 区别：隐藏 SaaS 相关功能，跳过计费流程
- 安全提示：生产环境中还需设置 `SIMPLE_MODE_CONFIRM=true` 才允许启动

---

## 项目结构

```
novaapi/
├── backend/                  # Go 后端服务
│   ├── cmd/server/           # 应用入口
│   ├── internal/             # 内部模块
│   │   ├── config/           # 配置
│   │   ├── model/            # 数据模型
│   │   ├── service/          # 业务逻辑
│   │   ├── handler/          # HTTP 处理器
│   │   └── gateway/          # API 网关核心
│   └── resources/            # 静态资源
│
├── frontend/                 # Vue 3 前端
│   └── src/
│       ├── api/              # API 调用
│       ├── stores/           # 状态管理
│       ├── views/            # 页面组件
│       └── components/       # 可复用组件
│
└── deploy/                   # 部署文件
    ├── docker-compose.yml    # Docker Compose 配置
    ├── .env.example          # Docker Compose 环境变量
    ├── config.example.yaml   # 二进制部署的完整配置文件
    └── install.sh            # 一键安装脚本
```

## 免责声明

> **使用本项目前请仔细阅读：**
>
> :rotating_light: **服务条款风险**：使用本项目可能违反 Anthropic 的服务条款。请在使用前仔细阅读 Anthropic 用户协议。由使用本项目产生的一切风险由使用者自行承担。
>
> :book: **免责声明**：本项目仅供技术学习与研究使用。对于因使用本项目导致的账号封禁、服务中断或任何其他损失，作者不承担任何责任。

---

## 许可证

本项目采用 [GNU Lesser General Public License v3.0](LICENSE)（或更高版本）授权。

---

<div align="center">

**如果觉得本项目有用，欢迎点个 Star！**

</div>
