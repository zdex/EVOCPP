# OCPP 1.6J Gateway (Go) — Starter Kit

This repo is a runnable OCPP 1.6J gateway skeleton:
- Accepts charger WebSocket connections at: `ws://localhost:8080/ocpp16/{chargePointId}` (use WSS in prod)
- Parses OCPP JSON frames
- Routes inbound actions to handlers (Boot/Heartbeat/Status/StartTx/MeterValues/StopTx)
- Emits normalized events to CPMS over HTTP (MVP)
- Accepts CPMS commands over HTTP at `POST /v1/gateway/commands`
- Sends outbound OCPP calls (RemoteStartTransaction / RemoteStopTransaction / ChangeConfiguration / Reset)
- Tracks pending call results (simple request/response correlation)

## Quick start

1) Run CPMS mock (optional) in another terminal:
```bash
go run ./cmd/cpms-mock
```

2) Run gateway:
```bash
export CPMS_BASE_URL=http://localhost:8081
export CPMS_API_KEY=dev
go run ./cmd/gateway
```

3) Connect a charger simulator (or any WS client) to:
`ws://localhost:8080/ocpp16/CP-123?secret=devsecret`

Gateway will accept auth if CPMS-mock returns allowed=true (or you can disable auth via env).

## Env vars

Gateway:
- LISTEN_ADDR=:8080
- CPMS_BASE_URL=http://localhost:8081
- CPMS_API_KEY=dev
- REQUIRE_CPMS_AUTH=true|false (default true)
- DEFAULT_HEARTBEAT_INTERVAL=300

CPMS Mock:
- CPMS_MOCK_ADDR=:8081
- CPMS_MOCK_ALLOW_SECRET=devsecret

## Notes
- For real deployments, replace `CheckOrigin: true`, enforce TLS/WSS, add rate limiting, and add Redis registry.
- This is intentionally minimal but “production-shaped”.
