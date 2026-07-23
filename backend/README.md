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

Any config value may reference an environment variable using `${VAR}` syntax
(resolved at load time; startup fails if the variable is unset). Use this to
keep secrets out of the file, e.g. `api_key: "${LLM_API_KEY}"`.
