# Deployment: Contabo + Dokploy (VPS)

This is the long-form Contabo + Dokploy guide, kept here as the repo-level reference.

It targets deploying:

- Go API: `apps/server/main-server`
- Python ELT: `apps/server/elt-server`
- (Optional) App: `apps/app` (usually Vercel, but can also be hosted elsewhere)

> Source: migrated from `CONTABO_DOKPLOY_DEPLOYMENT.md` (renamed for consistency).

## Full guide

The full step-by-step content lives in the original document text below (kept as-is, only wrapped by this header for naming consistency).

# MantrixFlow — Contabo + Dokploy Deployment Guide

Complete step-by-step guide to deploy MantrixFlow (Go API + Python ETL) on a Contabo VPS using Dokploy.

---

## Table of Contents

1. [Server Requirements](#1-server-requirements)
2. [Buy Contabo VPS](#2-buy-contabo-vps)
3. [Initial Server Setup](#3-initial-server-setup)
4. [Install Dokploy](#4-install-dokploy)
5. [Configure Domain & DNS](#5-configure-domain--dns)
6. [Deploy MantrixFlow](#6-deploy-mantrixflow)
7. [SSL Certificates](#7-ssl-certificates)
8. [Monitoring & Logs](#8-monitoring--logs)
9. [Backups](#9-backups)
10. [Scaling](#10-scaling)
11. [Troubleshooting](#11-troubleshooting)

---

## 1. Server Requirements

| Spec | Minimum | Recommended |
|------|---------|-------------|
| CPU | 2 vCPU | 4 vCPU |
| RAM | 4 GB | 8 GB |
| Disk | 50 GB SSD | 100 GB NVMe |
| OS | Ubuntu 22.04+ | Ubuntu 24.04 LTS |
| Bandwidth | 1 TB | Unlimited |

**Memory breakdown on 8GB server:**

| Component | RAM Usage |
|-----------|-----------|
| OS + Docker + Dokploy | ~800 MB |
| Go API | ~30–50 MB |
| ETL Python (idle) | ~100 MB |
| ETL Python (20 concurrent runs) | ~2–4 GB |
| **Total (20 concurrent runs)** | **~3–5 GB** |
| **Free headroom** | **~3–5 GB** |

---

## 2. Buy Contabo VPS

1. Go to [contabo.com/en/vps](https://contabo.com/en/vps/)
2. Select **Cloud VPS M** (4 vCPU / 8 GB / 200 GB NVMe) — **$7.49/mo** (or cheaper with 12-month commitment)
3. Choose:
   - **Region:** Closest to your Supabase project region
   - **OS:** Ubuntu 24.04
   - **Networking:** Enable IPv4
4. **Auto Backup:** Add daily backups (10 versions) — ~$0.93/mo
5. Complete purchase; note down the **IP address** and **root password**

---

## 3. Initial Server Setup

### 3.1 SSH into your server

```bash
ssh root@YOUR_SERVER_IP
```

### 3.2 System update

```bash
apt update && apt upgrade -y
```

### 3.3 Create a deploy user (don't run everything as root)

```bash
adduser deploy
usermod -aG sudo deploy
```

### 3.4 SSH key authentication (from your local machine)

```bash
# On your local machine
ssh-copy-id deploy@YOUR_SERVER_IP
```

### 3.5 Disable root password login

```bash
sudo sed -i 's/^PermitRootLogin yes/PermitRootLogin no/' /etc/ssh/sshd_config
sudo sed -i 's/^#PasswordAuthentication yes/PasswordAuthentication no/' /etc/ssh/sshd_config
sudo systemctl restart sshd
```

### 3.6 Basic firewall

```bash
sudo ufw allow OpenSSH
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw enable
```

> **Note:** Dokploy handles Docker and Traefik port mapping. Only 22, 80, 443 need to be open.

---

## 4. Install Dokploy

### 4.1 Install Dokploy (one command)

```bash
curl -sSL https://dokploy.com/install.sh | sh
```

This installs:
- Docker Engine
- Docker Compose
- Traefik (reverse proxy)
- Dokploy dashboard

### 4.2 Access Dokploy dashboard

Open in your browser:

```
http://YOUR_SERVER_IP:3000
```

1. Create your admin account
2. Set a strong password
3. You'll land on the Dokploy dashboard

### 4.3 Secure the Dokploy dashboard

In Dokploy dashboard → **Settings → Server**:
- Set a custom domain for the dashboard (e.g., `deploy.yourdomain.com`)
- Enable HTTPS (Dokploy auto-provisions Let's Encrypt)

---

## 5. Configure Domain & DNS

### 5.1 DNS records

Add these records at your domain registrar (Cloudflare, Namecheap, etc.):

| Type | Name | Value | Proxy |
|------|------|-------|-------|
| A | `api` | `YOUR_SERVER_IP` | Off (DNS only) |
| A | `deploy` | `YOUR_SERVER_IP` | Off (DNS only) |

> **Important:** If using Cloudflare, set to **DNS only** (grey cloud) initially. Dokploy/Traefik handles SSL.

### 5.2 Wait for DNS propagation

```bash
dig api.yourdomain.com +short
# Should return YOUR_SERVER_IP
```

---

## 6. Deploy MantrixFlow

### 6.1 Connect GitHub repo to Dokploy

1. In Dokploy dashboard → **Settings → Git**
2. Click **Connect GitHub**
3. Authorize Dokploy to access your `ai-bi` repository

### 6.2 Create a Compose project

1. Go to **Projects** → **Create Project**
2. Name: `mantrixflow`
3. Click the project → **Add Service** → **Compose**
4. Name: `mantrixflow-stack`

### 6.3 Configure the Compose service

**Source:** Select your GitHub repo and branch (`main`)

**Compose Path:** `docker-compose.prod.yml` (the file we created at the repo root)

### 6.4 Set environment variables

In the Compose service → **Environment** tab → add all variables from `env.production.example`:

```env
# Supabase
DATABASE_URL=postgresql://postgres.[ref]:[pass]@aws-0-[region].pooler.supabase.com:6543/postgres
DATABASE_DIRECT_URL=postgresql://postgres.[ref]:[pass]@aws-0-[region].pooler.supabase.com:5432/postgres
SUPABASE_URL=https://[ref].supabase.co
SUPABASE_ANON_KEY=eyJ...
SUPABASE_SERVICE_ROLE_KEY=eyJ...
SUPABASE_JWT_SECRET=your-jwt-secret

# Encryption (MUST be identical for Go and Python)
ENCRYPTION_MASTER_KEY=your-fernet-key-here
ENCRYPTION_KEY=your-fernet-key-here

# Inter-service auth
ETL_INTERNAL_TOKEN=generate-with-openssl-rand-hex-32
CALLBACK_TOKEN=generate-with-openssl-rand-hex-32

# Public URL (your domain)
API_PUBLIC_URL=https://api.yourdomain.com

# CORS
CORS_ALLOWED_ORIGINS=https://app.yourdomain.com

# ETL
MAX_CONCURRENT_RUNS=2
MAX_TAPS_PER_SOURCE=2
DEFAULT_SYNC_TIMEOUT_SECONDS=300
LOG_LEVEL=INFO
API_PORT=5000
```

### 6.5 Configure Domains (Traefik routing)

In Dokploy, the Compose service doesn't use Dokploy's domain routing directly (Traefik labels are in the compose file). Instead, you need to add Traefik labels.

**Option A: Use Dokploy's domain feature**

If Dokploy supports per-container domain routing for Compose services, set:
- Container `mantrixflow-api` → Domain: `api.yourdomain.com` → Port: `5000`

**Option B: Add Traefik labels to docker-compose.prod.yml**

If needed, add Traefik labels to the `api` service. See [Section 7](#7-ssl-certificates).

### 6.6 Deploy

Click **Deploy** in Dokploy. It will:
1. Pull your repo
2. Build both Docker images on the server
3. Start the containers
4. Run health checks

### 6.7 Verify

```bash
# From your local machine
curl https://api.yourdomain.com/health

# Expected response:
# {"status":"ok","timestamp":"..."}
```

---

## 7. SSL Certificates

### Option A: Dokploy handles it (easiest)

If you assigned a domain to the `api` container via Dokploy UI, SSL is automatic. Dokploy uses Traefik + Let's Encrypt.

### Option B: Traefik labels in Compose

If you need explicit Traefik labels, update the `api` service in `docker-compose.prod.yml`:

```yaml
services:
  api:
    # ... existing config ...
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.mantrixflow-api.rule=Host(`api.yourdomain.com`)"
      - "traefik.http.routers.mantrixflow-api.entrypoints=websecure"
      - "traefik.http.routers.mantrixflow-api.tls.certresolver=letsencrypt"
      - "traefik.http.services.mantrixflow-api.loadbalancer.server.port=5000"
    networks:
      - mantrixflow-internal
      - dokploy-network  # Dokploy's Traefik network

networks:
  mantrixflow-internal:
    driver: bridge
  dokploy-network:
    external: true
```

> **Note:** The network name `dokploy-network` may vary. Check with `docker network ls` on your server.

---

## 8. Monitoring & Logs

### 8.1 Dokploy dashboard

- **Logs:** Click on any service → Logs tab (live streaming)
- **Deployments:** See build history, roll back to any previous deploy
- **Resources:** CPU/RAM usage per container

### 8.2 CLI commands (SSH)

```bash
# All container statuses
docker ps

# Live logs (both services)
docker compose -f docker-compose.prod.yml logs -f

# Only API logs
docker logs mantrixflow-api -f --tail 100

# Only ETL logs
docker logs mantrixflow-etl -f --tail 100

# Resource usage
docker stats

# Check disk usage
docker system df
```

### 8.3 Health check endpoints

```bash
# Go API
curl http://localhost:5000/health

# ETL (only accessible from within the Docker network)
docker exec mantrixflow-api wget -qO- http://etl:8000/health
```

---

## 9. Backups

### 9.1 Contabo auto-backup

Enable during VPS purchase or in Contabo panel:
- Daily snapshots
- 10 versions retained
- ~$0.93/mo

### 9.2 Docker volume backup (optional)

```bash
# Backup ETL dlt state
docker run --rm \
  -v mantrixflow-etl-dlt-data:/data \
  -v $(pwd)/backups:/backup \
  alpine tar czf /backup/etl-data-$(date +%Y%m%d).tar.gz -C /data .
```

### 9.3 Database backup

Your database is on Supabase (managed). Supabase handles backups automatically on Pro plan. For Free plan, set up pg_dump:

```bash
# Manual backup (run from server or locally)
pg_dump "$DATABASE_DIRECT_URL" | gzip > backup-$(date +%Y%m%d).sql.gz
```

---

## 10. Scaling

### When to stay on 1 server

- < 20 concurrent ETL runs
- < 100 API requests/second
- RAM usage < 6 GB

### When to add a second server

- Consistently > 20 concurrent ETL runs
- RAM > 6 GB sustained
- Add second Contabo VPS for ETL only

### How to scale ETL to second server

1. Buy another Contabo VPS
2. Install Dokploy on it (or connect to existing Dokploy as a remote server)
3. Deploy only the ETL service on the second server
4. Update `ETL_PYTHON_SERVICE_URL` on the API to point to the second server's internal/private IP

### Vertical scaling

Contabo allows upgrading VPS plans without data loss:
- 4 vCPU / 8 GB → 6 vCPU / 16 GB → 8 vCPU / 30 GB

---

## 11. Troubleshooting

### Container won't start

```bash
# Check logs for errors
docker logs mantrixflow-api 2>&1 | tail -50
docker logs mantrixflow-etl 2>&1 | tail -50

# Rebuild from scratch
docker compose -f docker-compose.prod.yml build --no-cache
docker compose -f docker-compose.prod.yml up -d
```

### Health check failing

```bash
# Test inside container
docker exec mantrixflow-api wget -qO- http://localhost:5000/health
docker exec mantrixflow-etl curl -f http://localhost:8000/health
```

### Out of memory

```bash
# Check what's using memory
docker stats --no-stream

# If ETL is using too much, reduce concurrent runs
# In Dokploy env vars, set:
MAX_CONCURRENT_RUNS=10
```

### Can't reach API from internet

```bash
# Check firewall
sudo ufw status

# Check Traefik is running
docker ps | grep traefik

# Check DNS
dig api.yourdomain.com +short

# Check Dokploy port
sudo ss -tlnp | grep -E '80|443|3000|5000'
```

### Redeploy / rollback

In Dokploy dashboard:
1. Go to your Compose service
2. Click **Deployments** tab
3. Click **Rollback** on any previous successful deploy

Or via CLI:
```bash
cd /path/to/repo
git pull
docker compose -f docker-compose.prod.yml build
docker compose -f docker-compose.prod.yml up -d
```

### Disk full

```bash
# Clean unused Docker resources
docker system prune -f
docker image prune -a -f
docker volume prune -f
```

---

## Quick Reference

| Action | Command |
|--------|---------|
| Start services | `docker compose -f docker-compose.prod.yml up -d` |
| Stop services | `docker compose -f docker-compose.prod.yml down` |
| Rebuild & restart | `docker compose -f docker-compose.prod.yml up -d --build` |
| View logs | `docker compose -f docker-compose.prod.yml logs -f` |
| Check status | `docker ps` |
| Resource usage | `docker stats` |
| Enter API container | `docker exec -it mantrixflow-api sh` |
| Enter ETL container | `docker exec -it mantrixflow-etl bash` |

---

## Cost Summary

| Item | Monthly Cost |
|------|-------------|
| Contabo VPS (4 vCPU / 8 GB) | $4.95–$7.49 |
| Auto Backup (10 versions) | $0.93 |
| Dokploy | **Free** |
| SSL (Let's Encrypt) | **Free** |
| **Total** | **~$5.88–$8.42/mo** |

