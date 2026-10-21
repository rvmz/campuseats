# Session tasks

Check these off in order. Each item has a `Session N:` comment at the file you are supposed to change.

## 1 — Process baseline

Explanation: [session-01.md](session-01.md).

- [ ] `config.Load` reads `HTTP_ADDR` (default `:8080`)
- [ ] `logx.New` returns a JSON logger with a `service` attribute
- [ ] `gateway` serves `GET /healthz` with 200 and does not touch a database
- [ ] Ctrl-C stops `gateway` via `Shutdown` within 10 seconds
- [ ] `TestHealth` uses `httptest` and does not bind a port
- [ ] `order`, `kitchen`, and `notifier` start, log, and block until SIGINT/SIGTERM

## 2 — System design

Explanation: [session-02.md](session-02.md).

- [ ] Every heading in `docs/architecture.md` is filled in
- [ ] The place-order sequence names who talks HTTP, who talks gRPC, and who writes Postgres
- [ ] Two SLOs have numbers: `POST /v1/orders` latency, and how quickly the kitchen sees a committed order
- [ ] Three failures say what the student still has afterwards
- [ ] Proto and OpenAPI names match the API table in that doc

## 3 — Databases

Explanation: [session-03.md](session-03.md). Postgres user is `campuseats`, the same URL as `.env.example`.

- [ ] `migrations/001_init.sql` applies with goose up and reverses with goose down
- [ ] Money is integer cents. Status has a check constraint
- [ ] `orders.idempotency_key` and the `outbox` table exist
- [ ] Creating an order and its items is one transaction
- [ ] A failing item insert leaves no order row
- [ ] Price on the line is copied from `menu_items` at insert time
- [ ] `scripts/seed.sql` loads the fixed users, restaurant, and two items
- [ ] Generated sqlc code in `internal/gen/sqlc` is committed

## 4 — REST and gRPC

Explanation: [session-04.md](session-04.md). Stay on the `net/http` mux from session 1.

- [ ] `CreateOrder` and `GetOrder` are implemented on `order`
- [ ] `POST /v1/orders` and `GET /v1/orders/{id}` go through the gateway
- [ ] Empty items → 400. Unknown order → 404. Unknown gRPC failure → 500
- [ ] Gateway uses one gRPC client for the process lifetime and a 2s deadline
- [ ] `Idempotency-Key` is forwarded. It is not stored in the gateway
- [ ] `api/openapi/gateway.yaml` matches the routes you actually serve
- [ ] Postgres error text is logged, not returned to the client
- [ ] Generated protobuf code in `internal/gen` is committed

## 5 — Auth

Explanation: [session-05.md](session-05.md).

- [ ] Missing or bad bearer token → 401. Wrong role → 403
- [ ] `user_id` is taken from the token. The public JSON does not contain it
- [ ] A student cannot accept an order. A kitchen token reaches accept and gets 204 until session 6 changes the status
- [ ] A student cannot read someone else's order
- [ ] `/healthz` stays unauthenticated
- [ ] `JWT_SECRET` comes from the environment
- [ ] The JWT `alg` is checked

## 6 — Kafka and outbox

- [ ] The outbox insert commits in the same transaction as the order
- [ ] A publisher reads unpublished rows, produces to `orders.v1`, then sets `published_at`
- [ ] The message key is the order id
- [ ] With the publisher off, a created order stays unpublished and still exists
- [ ] Turning the publisher back on publishes that row without inserting a second order
- [ ] `kitchen` logs `order.created`. `notifier` logs a push
- [ ] Accept updates status and writes `order.accepted` in one transaction

## 7 — Reliability

- [ ] Empty idempotency key → 400
- [ ] Same user and same key returns the original order and does not insert another row
- [ ] Unique index is `(user_id, idempotency_key)`
- [ ] Accept only from `placed`. A second delivery changes zero rows and still commits the offset
- [ ] A database error in the consumer does not commit the offset
- [ ] Bad JSON is logged and skipped so the partition does not stall
- [ ] `CanTransition` matches the edges in `internal/order/status.go`

## 8 — Docker Compose

- [ ] Multi-stage Dockerfile builds the binary named by `SERVICE`, non-root, `CGO_ENABLED=0`
- [ ] `docker compose up --build` starts Postgres, Redpanda, migrate, order, gateway, kitchen
- [ ] App processes dial service DNS names, not `localhost`
- [ ] Migrate runs once and exits
- [ ] `down -v` then `up --build` still places an order
- [ ] README "Run" section has the commands and the curl

## 9 — Kubernetes

- [ ] `order` has `/healthz` (no database) and `/readyz` (pings Postgres) on the health port
- [ ] Deployment, Service, ConfigMap, and Secret apply on kind
- [ ] Image pull policy works with `kind load`
- [ ] A bad DSN makes the pod not Ready while liveness still passes
- [ ] Deleting the pod starts another one and does not drop the database volume
- [ ] `replicas` stays 1
- [ ] Gateway manifests are added

## 10 — AWS

- [ ] After a committed order, a receipt object is written at `receipts/<order_id>.json`
- [ ] Upload uses the default AWS credential chain
- [ ] IAM allows `s3:PutObject` only on that prefix
- [ ] A wrong prefix is denied, and you have seen that error
- [ ] If `RECEIPTS_BUCKET` is empty, create-order still succeeds
- [ ] S3 failure does not fail an already committed order
- [ ] No access key is committed
- [ ] The order image is tagged and pushed to ECR

## 11 — Monitoring

- [ ] Gateway `/metrics` has request count and a duration histogram
- [ ] Labels are method, route pattern, and status. No order id
- [ ] `campuseats_outbox_unpublished` climbs when publishing is off and drops when it is on
- [ ] `/metrics` and `/healthz` are not behind JWT
- [ ] Prometheus scrapes gateway and order. Grafana shows rate, p95, and the outbox gauge
- [ ] A log line for create-order includes `trace_id` and `order_id`

## 12 — Tests and CI

- [ ] Table test for `CanTransition`, including at least one illegal jump
- [ ] Integration test: two creates with one key leave one row and return one id
- [ ] Gateway test: gRPC `NotFound` becomes HTTP 404
- [ ] GitHub Actions runs `go test ./...` and is green
- [ ] `docs/incident.md` is filled in
- [ ] README demo script matches what you will show
