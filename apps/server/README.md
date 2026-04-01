# MantrixFlow Server — Docker Deployment

Deploy the Go API + Python ETL on a Contabo VPS using Dokploy.

## Folder Structure

```
server/
├── docker-compose.prod.yml      ← orchestrates both services
├── .env.production.example      ← env var template
├── .env.production              ← your actual env vars (git-ignored)
├── README.md                    ← this file
├── main-server/                 ← mantrixflow-api repo (cloned)
│   ├── Dockerfile               ← Go multi-stage build
│   ├── .dockerignore
│   ├── cmd/server/main.go
│   ├── internal/
│   └── ...
└── etl-server/                  ← mantrixflow-etl repo (cloned)
    ├── Dockerfile               ← Python multi-stage build
    ├── .dockerignore
    ├── api/main.py
    ├── core/
    └── ...
```

> **Both repos are separate GitHub repositories.** You clone them into `main-server/` and `etl-server/` on the deployment server.

---

## Step 1 — Buy Contabo VPS

1. Go to [contabo.com/en/vps](https://contabo.com/en/vps/)
2. Pick **Cloud VPS M**: 4 vCPU / 8 GB RAM / 200 GB NVMe — ~$4.95–$7.49/mo
3. **Region:** closest to your Supabase project
4. **OS:** Ubuntu 24.04
5. **Add-on:** Daily auto-backup (10 versions, ~$0.93/mo)

---

## Step 2 — Secure the Server

```bash
# SSH in
ssh root@YOUR_SERVER_IP

# Update
apt update && apt upgrade -y

# Create deploy user
adduser deploy
usermod -aG sudo deploy

# From your LOCAL machine — copy SSH key
ssh-copy-id deploy@YOUR_SERVER_IP

# Back on the SERVER — disable password login
sudo sed -i 's/^PermitRootLogin yes/PermitRootLogin no/' /etc/ssh/sshd_config
sudo sed -i 's/^#PasswordAuthentication yes/PasswordAuthentication no/' /etc/ssh/sshd_config
sudo systemctl restart sshd

# Firewall
sudo ufw allow OpenSSH
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw enable
```

---

## Step 3 — Install Dokploy

```bash
curl -sSL https://dokploy.com/install.sh | sh
```

This installs Docker, Docker Compose, Traefik, and Dokploy.

Access the dashboard:

```
http://YOUR_SERVER_IP:3000
```

Create your admin account and set a strong password.

---

## Step 4 — DNS Setup

At your domain registrar, add A records:

| Type | Name     | Value           |
|------|----------|-----------------|
| A    | `api`    | YOUR_SERVER_IP  |
| A    | `deploy` | YOUR_SERVER_IP  |

Verify:

```bash
dig api.yourdomain.com +short
# → YOUR_SERVER_IP
```

---

## Step 5 — Clone Repos & Configure

SSH into the server as the deploy user:

```bash
ssh deploy@YOUR_SERVER_IP
```

### 5.1 Create the server directory

```bash
mkdir -p ~/server
cd ~/server
```

### 5.2 Clone both repos

```bash
# Replace with your actual GitHub repo URLs
git clone git@github.com:yourorg/mantrixflow-api.git main-server
git clone git@github.com:yourorg/mantrixflow-etl.git etl-server
```

### 5.3 Copy Docker files into repos

If the Dockerfiles aren't already in the repos, copy them:

```bash
# If Dockerfile isn't already committed in mantrixflow-api repo:
cp /path/to/server/main-server/Dockerfile ~/server/main-server/Dockerfile
cp /path/to/server/main-server/.dockerignore ~/server/main-server/.dockerignore

# If Dockerfile isn't already committed in mantrixflow-etl repo:
cp /path/to/server/etl-server/Dockerfile ~/server/etl-server/Dockerfile
cp /path/to/server/etl-server/.dockerignore ~/server/etl-server/.dockerignore
```

> **Best practice:** Commit the Dockerfile and .dockerignore into each repo so they stay version-controlled.

### 5.4 Copy docker-compose and env file

```bash
# These should already be in ~/server/ from this server/ folder
# If doing manual setup:
cp docker-compose.prod.yml ~/server/
cp .env.production.example ~/server/.env.production
```

### 5.5 Fill environment variables

```bash
nano ~/server/.env.production
```

Fill in all values. Critical ones:

| Variable | Where to get it |
|----------|----------------|
| `DATABASE_URL` | Supabase → Settings → Database → Connection string (with pooler) |
| `DATABASE_DIRECT_URL` | Supabase → Settings → Database → Direct connection |
| `SUPABASE_JWT_SECRET` | Supabase → Settings → API → JWT Secret |
| `SUPABASE_URL` | Supabase → Settings → API → Project URL |
| `SUPABASE_ANON_KEY` | Supabase → Settings → API → anon public key |
| `SUPABASE_SERVICE_ROLE_KEY` | Supabase → Settings → API → service_role key |
| `ENCRYPTION_MASTER_KEY` | Generate: `python3 -c "from cryptography.fernet import Fernet; print(Fernet.generate_key().decode())"` |
| `ENCRYPTION_KEY` | Same as `ENCRYPTION_MASTER_KEY` |
| `ETL_INTERNAL_TOKEN` | Generate: `openssl rand -hex 32` |
| `CALLBACK_TOKEN` | Generate: `openssl rand -hex 32` |
| `API_PUBLIC_URL` | `https://api.yourdomain.com` |
| `CORS_ALLOWED_ORIGINS` | Your frontend domain(s) |

---

## Step 6 — Deploy with Docker Compose

```bash
cd ~/server

# Build and start both services
docker compose -f docker-compose.prod.yml up -d --build
```

### Verify

```bash
# Check containers are running
docker ps

# Expected output:
# mantrixflow-api   ... Up (healthy)
# mantrixflow-etl   ... Up (healthy)

# Test API health
curl http://localhost:5000/health
# → {"status":"ok"}

# Test ETL health (from inside network)
docker exec mantrixflow-api wget -qO- http://etl:8000/health
```

---

## Step 7 — Configure Dokploy Domain & SSL

### Option A: Dokploy UI (easiest)

1. Open Dokploy dashboard at `http://YOUR_SERVER_IP:3000`
2. Go to **Projects** → your compose project
3. Assign domain `api.yourdomain.com` to the `api` container, port `5000`
4. Dokploy auto-provisions Let's Encrypt SSL

### Option B: Traefik labels (manual)

Add these labels to the `api` service in `docker-compose.prod.yml`:

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
      - dokploy-network

networks:
  dokploy-network:
    external: true
```

Then redeploy:

```bash
docker compose -f docker-compose.prod.yml up -d
```

### Verify HTTPS

```bash
curl https://api.yourdomain.com/health
# → {"status":"ok"}
```

---

## Step 8 — Auto-Deploy on Git Push

### Option A: Dokploy webhook (recommended)

1. In Dokploy → your project → **Build & Deploy** → copy the webhook URL
2. In each GitHub repo → **Settings → Webhooks → Add webhook**
   - URL: the Dokploy webhook URL
   - Events: `push`
   - Branch: `main`

Now every `git push` to `main` triggers a rebuild.

### Option B: Manual deploy script

Create `~/server/deploy.sh`:

```bash
#!/bin/bash
set -euo pipefail

cd ~/server

echo "→ Pulling main-server..."
cd main-server && git pull origin main && cd ..

echo "→ Pulling etl-server..."
cd etl-server && git pull origin main && cd ..

echo "→ Rebuilding & restarting..."
docker compose -f docker-compose.prod.yml up -d --build

echo "→ Cleaning old images..."
docker image prune -f

echo "✓ Deploy complete"
docker ps
```

```bash
chmod +x ~/server/deploy.sh
```

Run anytime: `~/server/deploy.sh`

---

## Common Operations

| Action | Command |
|--------|---------|
| Start all | `docker compose -f docker-compose.prod.yml up -d` |
| Stop all | `docker compose -f docker-compose.prod.yml down` |
| Rebuild & restart | `docker compose -f docker-compose.prod.yml up -d --build` |
| View all logs | `docker compose -f docker-compose.prod.yml logs -f` |
| API logs only | `docker logs mantrixflow-api -f --tail 100` |
| ETL logs only | `docker logs mantrixflow-etl -f --tail 100` |
| Resource usage | `docker stats` |
| Enter API shell | `docker exec -it mantrixflow-api sh` |
| Enter ETL shell | `docker exec -it mantrixflow-etl bash` |
| Update API only | `cd main-server && git pull && cd .. && docker compose -f docker-compose.prod.yml up -d --build api` |
| Update ETL only | `cd etl-server && git pull && cd .. && docker compose -f docker-compose.prod.yml up -d --build etl` |
| Clean disk | `docker system prune -f && docker image prune -a -f` |

---

## Troubleshooting

### Container won't start

```bash
docker logs mantrixflow-api 2>&1 | tail -50
docker logs mantrixflow-etl 2>&1 | tail -50
```

### Health check failing

```bash
docker exec mantrixflow-api wget -qO- http://localhost:5000/health
docker exec mantrixflow-etl curl -f http://localhost:8000/health
```

### Out of memory

```bash
docker stats --no-stream
# Reduce concurrent ETL runs:
# Edit .env.production → MAX_CONCURRENT_RUNS=2
# Then: docker compose -f docker-compose.prod.yml up -d
```

### API can't reach ETL

```bash
# Both must be on the same Docker network
docker network inspect mantrixflow-internal
# Should show both containers
```

### Rebuild from scratch

```bash
docker compose -f docker-compose.prod.yml down
docker compose -f docker-compose.prod.yml build --no-cache
docker compose -f docker-compose.prod.yml up -d
```

---

## Architecture

```
Internet
   │
   ▼
[Traefik / Dokploy]  ← SSL termination, reverse proxy
   │
   ▼ port 5000
┌──────────────┐      ┌──────────────┐
│  Go API      │─────▶│  Python ETL  │
│  (Fiber)     │ :8000│  (FastAPI)   │
│              │◀─────│  (dlt)       │
│  mantrixflow │callback             │
│  -api        │      │  mantrixflow │
│              │      │  -etl        │
└──────────────┘      └──────────────┘
   │                       │
   ▼                       ▼
[Supabase Postgres]   [Destination DBs]
```

- **API is public** — exposed via Traefik with SSL
- **ETL is internal only** — only reachable by the API container via Docker network
- Both share `ENCRYPTION_KEY` for credential encryption/decryption

---

## Cost

| Item | Monthly |
|------|---------|
| Contabo VPS (4 vCPU / 8 GB) | $4.95–$7.49 |
| Auto Backup (daily, 10 versions) | $0.93 |
| Dokploy | **Free** |
| SSL (Let's Encrypt) | **Free** |
| **Total** | **~$5.88–$8.42/mo** |
