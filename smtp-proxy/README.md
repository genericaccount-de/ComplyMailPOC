# ComplyMail SMTP Proxy

Lightweight SMTP gateway that sits in front of the customer's mail server.

For each outbound email it:
1. Parses the message (subject, body, recipients).
2. Calls the backend API (`POST /scan-outbound-email`) for rule checks and LLM classification.
3. Based on the result, either relays the email upstream or redirects it to a review mailbox.

## Run locally

```bash
go run ./cmd/proxy -config config.yaml
```

## Configuration

Configuration is loaded from a YAML file (default `config.yaml`, override with
`-config`). Copy `config.example.yaml` to `config.yaml` and adjust as needed.
