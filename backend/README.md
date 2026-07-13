# ComplyMail Backend API

HTTP service exposing:
- `POST /check-compose-email` — called by the Outlook Add-in to get style & security suggestions.
- `POST /scan-outbound-email` — called by the SMTP proxy to classify outbound emails.
- `GET /healthz` — health check.

## Run locally

```bash
go run ./cmd/api
```

## Configuration

See `../.env.example` for available environment variables.
