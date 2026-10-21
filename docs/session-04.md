# Session 4 — HTTP at the edge, gRPC behind it

The phone and curl speak HTTP and JSON. `gateway` and `order` speak gRPC, because the call has a method, a typed body, and a deadline. The order rules stay in `order`, next to the transaction from session 3. The gateway decodes JSON, calls one RPC, and maps the code. It does not write SQL, and it does not decide the price.

By the end:

- `CreateOrder` and `GetOrder` run on `order`
- `POST /v1/orders` and `GET /v1/orders/{id}` go through `gateway`
- an empty item list is 400, an unknown order is 404, and an unexpected failure is 500
- the gateway dials once, when the process starts, and puts a 2 second deadline on each call
- `Idempotency-Key` is forwarded and not stored in the gateway
- `api/openapi/gateway.yaml` matches the routes you serve
- a Postgres error is logged on the server and is not copied into the HTTP body

Kitchen RPCs stay declared and unimplemented. Accept is session 5's HTTP route and session 6's transaction.

## Generate the contract

Session 2 filled in the proto. From the repo root, with the plugins on your `PATH`:

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
protoc -I api/proto \
  --go_out=. --go_opt=module=campuseats \
  --go-grpc_out=. --go-grpc_opt=module=campuseats \
  api/proto/order/v1/order.proto \
  api/proto/kitchen/v1/kitchen.proto
```

`buf generate`, run from `api/`, writes the same tree because `api/buf.gen.yaml` sets `out` to the module root. Commit `internal/gen`. Add the module dependencies:

```bash
go get google.golang.org/grpc google.golang.org/protobuf
```

## The order service

`cmd/order` listens on `GRPC_ADDR` (`:9090`). Extend `config.Load` so that field is read. `order` refuses to start when `DATABASE_URL` is empty. The gateway does not need a database URL.

```go
func (s *Server) CreateOrder(ctx context.Context, req *orderv1.CreateOrderRequest) (*orderv1.Order, error) {
    if req.RestaurantId == "" || len(req.Items) == 0 {
        return nil, status.Error(codes.InvalidArgument, "items required")
    }
    for _, item := range req.Items {
        if item.Qty < 1 || item.MenuItemId == "" {
            return nil, status.Error(codes.InvalidArgument, "bad item")
        }
    }
    order, err := s.store.Create(ctx, toParams(req))
    if err != nil {
        s.log.Error("create order", "err", err)
        return nil, status.Error(codes.Internal, "create order")
    }
    return toProto(order), nil
}
```

`GetOrder` uses the query from session 3. `pgx.ErrNoRows` becomes `codes.NotFound`. Anything else is logged and returned as `codes.Internal`. Do not put `err.Error()` into the gRPC status. The text would include table names and constraints. The proto already has `caller_id` and `caller_role`. Leave them empty this week. Session 5 compares them with the row; doing that now, with an empty caller, would reject every read.

Pass `r.Context()` through to the query so a cancelled call stops the SQL. The store method must take `context.Context`.

Shut the gRPC server down the same way session 1 shuts HTTP down: `signal.NotifyContext`, serve in a goroutine, then `GracefulStop` after `ctx.Done()`.

## The gateway

Keep the `net/http` mux from session 1. One gRPC client is created in `main` and reused. `ORDER_ADDR` on the host is `localhost:9090`. Inside Compose, in session 8, that host name becomes `order`. `localhost` from inside the gateway container would be the gateway itself.

```go
conn, err := grpc.NewClient(cfg.OrderAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
```

Plaintext is a lab choice. Do not add TLS this week.

```go
func (h *Handler) createOrder(w http.ResponseWriter, r *http.Request) {
    var body createBody
    if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&body); err != nil {
        http.Error(w, "bad json", http.StatusBadRequest)
        return
    }
    ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
    defer cancel()
    resp, err := h.orders.CreateOrder(ctx, &orderv1.CreateOrderRequest{
        UserId:         body.UserID, // session 5 deletes this field and uses the token
        RestaurantId:   body.RestaurantID,
        IdempotencyKey: r.Header.Get("Idempotency-Key"),
        Items:          toItems(body.Items),
    })
    if err != nil {
        writeGRPCError(w, err)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(resp)
}
```

`user_id` in the JSON is a hole. Any caller can claim to be Ada. Session 5 closes it. Leave the hole visible; do not invent a fake check that trusts the body anyway.

`GET /v1/orders/{id}` reads `r.PathValue("id")`. If it is not a UUID, return 400 and do not call gRPC.

| gRPC code | HTTP | When |
|---|---|---|
| `InvalidArgument` | 400 | empty items, bad qty, bad id |
| `NotFound` | 404 | no such order |
| `DeadlineExceeded` | 504 | the 2 second deadline fired |
| `Unavailable` | 503 | order is not listening |
| anything else | 500 | you did not plan for it |

`PermissionDenied` becomes 403 in session 5. `AlreadyExists` becomes 409 in session 7, when the idempotency key is actually enforced. Forwarding the header now does not make a second POST safe.

`/healthz` stays on the mux and still does not touch Postgres.

## OpenAPI

Describe `POST /v1/orders` and `GET /v1/orders/{id}` with the status codes above. This file is the edge contract. You are not generating the server from it. Keep the menu paths from session 2 in the file if you listed them. Implementing those handlers is not this session. The checklist is the two order routes.

## Check

Postgres is up, the migration is applied, and `scripts/seed.sql` has been loaded.

```bash
curl -sS -D- -X POST localhost:8080/v1/orders \
  -H 'content-type: application/json' \
  -H 'Idempotency-Key: lab-1' \
  -d '{"user_id":"11111111-1111-1111-1111-111111111111","restaurant_id":"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa","items":[{"menu_item_id":"bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb","qty":1}]}'
```

You want `201`, a body with `"status":"placed"`, and one row. `{"items":[]}` returns 400. `GET /v1/orders/11111111-1111-1111-1111-111111111111` returns 404, because that id is Ada, not an order. Stop Postgres and curl again: the gateway returns 504 or 503, not a hang. Start Postgres again.

## What this session is not

gRPC-Gateway, the project that generates the HTTP layer from the proto, hides the mapping you are writing by hand. Protobuf is not the browser API. A new gRPC connection per request will exhaust ports; dial in `main`.
