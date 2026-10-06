# CampusEats

CampusEats is a campus food-ordering backend. A student places an order, the kitchen accepts it, and a notifier records a status update. Payments and courier GPS are out of scope. The order row in Postgres is the source of truth.

This repository is the whole course. Each session extends it. Comments that start with `Session N:` are the work for that session.

Session 1 is written up in [docs/session-01.md](docs/session-01.md): why a process loads config from the environment, logs JSON, answers `/healthz`, and shuts down on SIGTERM. Later sessions are checklists in [docs/tasks.md](docs/tasks.md) until their write-ups are added.

## Layout

| Path | What it becomes |
|---|---|
| `cmd/gateway` | Public HTTP API |
| `cmd/order` | gRPC API, database, outbox publisher |
| `cmd/kitchen` | Kitchen commands and Kafka consumer |
| `cmd/notifier` | Kafka consumer that logs a push |
| `internal/order` | Status rules and order persistence |
| `internal/platform` | Config and logging shared by every process |
| `api/proto` | gRPC contracts |
| `api/openapi/gateway.yaml` | HTTP contract |
| `migrations` | Postgres schema |
| `deploy` | Docker, Compose, Kubernetes, AWS policy |
| `docs/tasks.md` | What to finish each session |
| `docs/architecture.md` | Design worksheet |

## Build

Go 1.23 or newer.

```bash
go build ./...
go test ./...
```

The processes exit immediately until session 1 wires them. Do not add a second module or a second repository.

## Run

Session 8 fills this in: Compose command, the curl that places an order, and where logs go.

## Status vocabulary

`placed`, `accepted`, `preparing`, `ready`, `picked_up`, `delivered`, `rejected`, `cancelled`

These strings live in `internal/order/status.go`. If you rename one, change the database constraint and `docs/architecture.md` in the same commit.
