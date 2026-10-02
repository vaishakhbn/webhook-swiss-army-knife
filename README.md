# Webhook Swiss Army Knife

A tiny, dependency-free Go service for inspecting and debugging webhooks. The
first tool accepts a POST request and writes its `Authorization` header to the
service's structured logs.

[![CI](https://github.com/vaishakhbn/webhook-swiss-army-knife/actions/workflows/ci.yml/badge.svg)](https://github.com/vaishakhbn/webhook-swiss-army-knife/actions/workflows/ci.yml)
[![Deploy to Render](https://render.com/images/deploy-to-render-button.svg)](https://render.com/deploy?repo=https://github.com/vaishakhbn/webhook-swiss-army-knife)

> [!CAUTION]
> Authorization headers commonly contain credentials. Use this only for
> controlled debugging, restrict access to the service, and remove sensitive
> logs when you are done.

## Run locally

```sh
go run ./cmd/server
```

In another terminal:

```sh
curl -X POST http://localhost:10000/authorization \
  -H 'Authorization: Bearer example-token'
```

Add any single path segment—such as a UUID—to correlate the log entry with a
specific sender or test:

```sh
curl -X POST \
  http://localhost:10000/authorization/550e8400-e29b-41d4-a716-446655440000 \
  -H 'Authorization: Bearer example-token'
```

The response is deliberately minimal and never echoes the credential:

```json
{"logged":true}
```

The server writes a JSON log entry to stdout. On Render, find it on the
service's **Logs** page. Each authorization log includes:

- `client_ip`: the original caller IP (the first address in Render's
  `X-Forwarded-For` header), falling back to the direct connection address
- `x_forwarded_for`: the complete proxy chain for debugging
- `cf_ray`: Cloudflare's request trace ID
- `remote_addr`: the direct connection address, which is usually Render's proxy

## Endpoints

| Method | Path | Purpose |
| --- | --- | --- |
| `POST` | `/authorization` | Log the `Authorization` header |
| `POST` | `/authorization/{request_id}` | Log the header and request identifier |
| `GET` | `/healthz` | Render health check |
| `GET` | `/` | List the available tools |

## Deploy to Render

Click **Deploy to Render** above, or create a Blueprint in Render and connect
this GitHub repository. Render reads `render.yaml` and creates the service.

Then send requests to
`https://<your-service>.onrender.com/authorization`.

No environment variables are required. Render supplies `PORT` automatically;
local runs default to port `10000`.

## Add another tool

Register another method-and-path handler in `internal/httpserver/server.go`, add
a focused test, and update the endpoint table above.

## Test

```sh
go test ./...
```
