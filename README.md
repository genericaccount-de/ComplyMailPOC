# ComplyMail POC

AI-powered email style and compliance checker for Outlook. See [DESIGN.md](DESIGN.md) for the full design document.

## Repository Structure

| Directory | Description | Stack |
|---|---|---|
| `backend/` | Backend API — rule engine, LLM orchestrator, REST endpoints | Go |
| `smtp-proxy/` | Outbound SMTP gateway — scans emails via backend API | Go |
| `outlook-addin/` | Outlook Web Add-in — compose-time checks | TypeScript / HTML |
| `admin-console/` | Admin SPA — rule management & stats | React / TypeScript |
| `deploy/` | Kubernetes manifests, Terraform | — |
| `docs/` | OpenAPI spec, threat model, architecture | — |

## Quick Start

```bash
# Build Go services
make build-all

# Run backend API locally
make run-api

# Run SMTP proxy locally
make run-proxy

# Install JS dependencies
cd outlook-addin && npm install
cd admin-console && npm install

# Start everything via Docker Compose
make docker-up
```

## Configuration

Copy `.env.example` to `.env` and fill in values (Mistral API key, upstream SMTP host, etc.).