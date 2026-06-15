# NovaAPI

<div align="center">

[![Go](https://img.shields.io/badge/Go-1.25+-00ADD8.svg)](https://golang.org/)
[![Vue](https://img.shields.io/badge/Vue-3.4+-4FC08D.svg)](https://vuejs.org/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15+-336791.svg)](https://www.postgresql.org/)
[![Redis](https://img.shields.io/badge/Redis-7+-DC382D.svg)](https://redis.io/)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED.svg)](https://www.docker.com/)

**AI API Gateway Platform for Subscription Quota Distribution**

English | [中文](README_ZH.md)

</div>

## ⚠️ Important Notice

Please read the following carefully before using this project:

- **🚨 Terms of Service Risk**: Using this project may violate the terms of service of upstream providers. Review the relevant providers' user agreements before use; all risks arising from such use are borne solely by the user.
- **⚖️ Compliant Use**: Use this project only in compliance with the laws and regulations of your country or region. Any unlawful use is strictly prohibited.
- **📖 Disclaimer**: This project is provided for technical learning and research purposes only. The authors assume no liability for account bans, service interruptions, data loss, or any other direct or indirect damages resulting from the use of this project.

## Overview

NovaAPI is an AI API gateway platform designed to distribute and manage API quotas from AI product subscriptions. Users access upstream AI services through platform-generated API Keys, while the platform handles authentication, billing, load balancing, and request forwarding.

## Features

- **Multi-Account Management** — Support multiple upstream account types (OAuth, API Key)
- **API Key Distribution** — Generate and manage API Keys for users
- **Precise Billing** — Token-level usage tracking and cost calculation
- **Smart Scheduling** — Intelligent account selection with sticky sessions
- **Concurrency Control** — Per-user and per-account concurrency limits
- **Rate Limiting** — Configurable request and token rate limits
- **Built-in Payment System** — Supports EasyPay, Alipay, WeChat Pay, and Stripe for user self-service top-up ([Configuration Guide](docs/PAYMENT.md))
- **Social Login** — Built-in OAuth sign-in with GitHub, Google, and Linux.do
- **Admin Dashboard** — Web interface for monitoring and management
- **External System Integration** — Embed external systems (e.g. ticketing) via iframe to extend the admin dashboard

## Tech Stack

| Component | Technology |
|-----------|------------|
| Backend | Go 1.25+, Gin, Ent |
| Frontend | Vue 3.4+, Vite 5+, TailwindCSS |
| Database | PostgreSQL 15+ |
| Cache/Queue | Redis 7+ |

---

## Deployment

> Replace `your-org/novaapi` and `https://your-domain.com` below with your own repository and domain.

### Method 1: Docker Compose (Recommended)

#### Prerequisites

- Docker 20.10+
- Docker Compose v2+

#### Manual Setup

```bash
# 1. Clone the repository
git clone https://github.com/your-org/novaapi.git
cd novaapi/deploy

# 2. Copy environment configuration
cp .env.example .env

# 3. Edit configuration (generate secure secrets)
nano .env
```

**Required configuration in `.env`:**

```bash
# PostgreSQL password (REQUIRED)
POSTGRES_PASSWORD=your_secure_password_here

# JWT Secret (RECOMMENDED — keeps users logged in after restart)
JWT_SECRET=your_jwt_secret_here

# TOTP Encryption Key (RECOMMENDED — preserves 2FA after restart)
TOTP_ENCRYPTION_KEY=your_totp_key_here

# Optional: Admin account
ADMIN_EMAIL=admin@example.com
ADMIN_PASSWORD=your_admin_password

# Optional: Custom port
SERVER_PORT=8080
```

**Generate secure secrets:**

```bash
openssl rand -hex 32   # JWT_SECRET
openssl rand -hex 32   # TOTP_ENCRYPTION_KEY
openssl rand -hex 32   # POSTGRES_PASSWORD
```

```bash
# 4. Create data directories
mkdir -p data postgres_data redis_data

# 5. Start all services
docker compose -f docker-compose.local.yml up -d

# 6. Check status / logs
docker compose -f docker-compose.local.yml ps
docker compose -f docker-compose.local.yml logs -f novaapi
```

#### Access

Open `http://YOUR_SERVER_IP:8080` in your browser. If the admin password was auto-generated, find it in the logs:

```bash
docker compose -f docker-compose.local.yml logs novaapi | grep "admin password"
```

#### Upgrade

```bash
docker compose -f docker-compose.local.yml pull
docker compose -f docker-compose.local.yml up -d
```

---

### Method 2: Build from Source

#### Prerequisites

- Go 1.21+
- Node.js 18+

#### Build

```bash
# Frontend
cd frontend
pnpm install
pnpm run build

# Backend
cd ../backend
go build -o novaapi ./cmd/server

# Run
./novaapi
```

#### Development Mode

```bash
# Backend (hot reload)
cd backend
go run ./cmd/server

# Frontend (hot reload)
cd frontend
pnpm run dev
```

#### Code Generation

When editing `backend/ent/schema`, regenerate Ent + Wire:

```bash
cd backend
go generate ./ent
go generate ./cmd/server
```

---

## Social Login (OAuth)

NovaAPI ships with built-in OAuth sign-in for **GitHub**, **Google**, and **Linux.do**. Configure each provider in **Admin Dashboard → Settings → OAuth/Login** — no code changes needed.

| Provider | Apply at | Backend Callback URL |
|----------|----------|----------------------|
| GitHub | github.com/settings/developers → OAuth Apps | `https://your-domain.com/api/v1/auth/oauth/github/callback` |
| Google | console.cloud.google.com/apis/credentials | `https://your-domain.com/api/v1/auth/oauth/google/callback` |
| Linux.do | connect.linux.do | `https://your-domain.com/api/v1/auth/oauth/linuxdo/callback` |

- Frontend callback for all three: `/auth/oauth/callback`
- Required scopes are sent automatically (GitHub: `read:user user:email`; Google: `openid email profile`)
- **Requires a real domain over HTTPS** — the callback URL must be publicly reachable. The callback registered with the provider, the one entered in the admin panel, and the actual access domain must match exactly, or you'll get `redirect_uri_mismatch`.
- Settings are stored in the database and take effect on save (no restart needed).

---

## Nginx Reverse Proxy Note

When using Nginx as a reverse proxy with Codex CLI, add the following to the `http` block:

```nginx
underscores_in_headers on;
```

Nginx drops headers containing underscores by default (e.g. `session_id`), which breaks sticky session routing in multi-account setups.

---

## Simple Mode

Simple Mode is designed for individual developers or internal teams who want quick access without full SaaS features.

- **Enable**: Set environment variable `RUN_MODE=simple`
- **Difference**: Hides SaaS-related features and skips the billing process
- **Security**: In production, you must also set `SIMPLE_MODE_CONFIRM=true` to allow startup

---

## Project Structure

```
novaapi/
├── backend/                  # Go backend service
│   ├── cmd/server/           # Application entry
│   ├── internal/             # Internal modules
│   │   ├── config/           # Configuration
│   │   ├── model/            # Data models
│   │   ├── service/          # Business logic
│   │   ├── handler/          # HTTP handlers
│   │   └── gateway/          # API gateway core
│   └── resources/            # Static resources
│
├── frontend/                 # Vue 3 frontend
│   └── src/
│       ├── api/              # API calls
│       ├── stores/           # State management
│       ├── views/            # Page components
│       └── components/       # Reusable components
│
└── deploy/                   # Deployment files
    ├── docker-compose.yml    # Docker Compose configuration
    ├── .env.example          # Environment variables
    ├── config.example.yaml   # Full config for binary deployment
    └── install.sh            # One-click installation script
```

---

## Disclaimer

> **Please read carefully before using this project:**
>
> :rotating_light: **Terms of Service Risk**: Using this project may violate upstream providers' Terms of Service. Read the relevant user agreements carefully before use. All risks arising from the use of this project are borne solely by the user.
>
> :book: **Disclaimer**: This project is for technical learning and research purposes only. The authors assume no responsibility for account suspension, service interruption, or any other losses caused by the use of this project.

---

## License

This project is licensed under the [GNU Lesser General Public License v3.0](LICENSE) (or later).

---

<div align="center">

**If you find NovaAPI useful, please give it a star!**

</div>
