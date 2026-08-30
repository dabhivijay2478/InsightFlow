# Service Startup Report

Run date: 2026-07-26  
E2E run prefix: `e2e_1785038655750`

| Service | Command | URL | Health endpoint | Status | Startup failure | Log path |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| Next.js frontend | `bun run dev` | `http://localhost:3000` | `/auth/login` | Passed | None | Existing Codex terminal process |
| Go control plane | `go run ./cmd/server` | `http://localhost:5000` | `/api/v1/health` | Passed | None after restart with fixes | Codex terminal session `80697` |
| Python ELT | `.venv/bin/python -m uvicorn api.main:app --host 0.0.0.0 --port 8000` | `http://localhost:8000` | `/health` | Passed | None after restart with fixes | Codex terminal session `51647` |

All product actions followed Browser → Go API → Python ELT. No browser action called the ELT service directly.
