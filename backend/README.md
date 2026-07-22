# ComplyMail Backend API

HTTP service exposing:
- `POST /check-style-email` — called by the Outlook Add-in to get style & security suggestions.
- `POST /scan-outbound-email` — called by the SMTP proxy to classify outbound emails.
- `GET /healthz` — health check.

## Run locally

```bash
go run ./cmd/api -config config.yaml
```

## Configuration

Configuration is loaded from a YAML file (default `config.yaml`, override with
`-config`). Copy `config.example.yaml` to `config.yaml` and adjust as needed.

The LLM API key can be supplied via the config file (`llm.api_key`) or overridden
at runtime with the `LLM_API_KEY` environment variable to keep secrets out of the
file.
