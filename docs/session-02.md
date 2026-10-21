# Session 2 — Where an order lives

Session 1 made a process that can start and stop. This session decides what the processes are allowed to do. Later sessions implement this document. They do not invent a second design.

Write `docs/architecture.md` in your own sentences. The names in this file are the contract: routes, RPCs, statuses, and the topic. If you rename one, change the proto, the OpenAPI file, and `internal/order/status.go` in the same commit.

By the end:

- every heading in `docs/architecture.md` is filled in
- the place-order sequence names HTTP, gRPC, SQL, and Kafka, and marks which step returns to the student
- two SLOs have numbers
- three failures say what is still true afterwards
- `api/proto` and `api/openapi/gateway.yaml` use the same names as the API table

No new dependency. Postgres is not running yet.

## The sentence

Postgres is the source of truth for an order. Kafka tells the rest of the campus that something changed. The gateway does not decide order status, and it does not write SQL.

## Who is outside

Three apps sit outside CampusEats: the student phone, the kitchen tablet, and the courier phone. Payments are a fake "this order exists" moment. Push notifications are a log line. Maps are out of scope.

## What runs

Draw these boxes and put a protocol on every arrow.

| From | To | Protocol | Caller waits? |
|---|---|---|---|
| phone | `gateway` | HTTP | yes |
| `gateway` | `order` | gRPC | yes |
| `order` | Postgres | SQL | yes, one transaction |
| `order` | topic `orders.v1` | Kafka | no |
| `kitchen`, `notifier` | topic `orders.v1` | Kafka | they read later |
| kitchen tablet | `gateway` | HTTP | yes, for accept / reject / ready |

S3 is a dotted box for session 10. Do not design it today.

One Postgres database for the course. `order` inserts orders. `kitchen` will update `orders.status`. A database per service would leave you with no single row that says what the student ordered. Schema ownership here is by table, not by cluster. That is a choice for a lab of this size.

`cmd/` already has the four processes. `notifier` only reacts. It is not on the path of `POST /v1/orders`. If the gateway waited for a push to be sent, a slow notifier would fail lunch.

## Sync and async

The caller waits when it cannot proceed without the answer: create an order, read an order, accept an order.

The caller does not wait when other parts of the campus react: the kitchen ticket, the push log. The HTTP response returns at status `placed`, before the notifier runs.

The design you are refusing: gateway calls kitchen, kitchen calls notifier, notifier calls email, and the student's request dies because email is slow.

## The crash

"The student tapped Pay. You inserted the row. The process died before Kafka."

The order exists exactly once. A later step is responsible for telling the kitchen. Publishing to Kafka and inserting afterwards is the wrong order: the kitchen would cook an order that has no row, and you cannot pull the message back.

You will build that later step in session 6. Today the document only forbids the wrong order.

## SLO

Write both of these down, with the numbers. "The system is fast" is not an SLO. Session 11 will put the first one on a graph.

- 99% of `POST /v1/orders` succeed or fail in under 300 ms. The kitchen's human is not included.
- 99% of committed orders are visible to the kitchen within 5 seconds.

Lunch fails when the kitchen is blind for twenty minutes. It does not fail because you missed five nines.

## Place order

Use this sequence. Mark step 4 as the HTTP response.

1. Student `POST /v1/orders` with `Authorization` and `Idempotency-Key`.
2. Gateway checks the token (session 5) and calls `OrderService.CreateOrder` with a deadline.
3. `order` inserts `orders`, `order_items`, and an outbox row in one transaction. Status is `placed`. The price is copied from `menu_items`.
4. Gateway returns `201` and the order id. The notifier has not run. That is success.
5. A publisher reads unpublished outbox rows and produces to `orders.v1`. The message key is the order id.
6. `kitchen` and `notifier` read the event. They are separate consumer groups, so both see it.
7. Staff `POST /v1/kitchen/orders/{id}/accept`. Status becomes `accepted` only from `placed`. That command is synchronous. Another outbox row, `order.accepted`, is written in the same transaction.

Statuses, already in `internal/order/status.go`: `placed`, `accepted`, `preparing`, `ready`, `picked_up`, `delivered`, `rejected`, `cancelled`.

## Three failures

Write four lines for each: what broke, what the student sees, what is still true, which later session handles it.

1. **Double tap.** Two POSTs, one idempotency key. One order. Session 7.
2. **Kafka down after commit.** The student has an order id. The kitchen does not, until the publisher retries. The order is not deleted. Session 6.
3. **Kitchen handles the same event twice.** The second update changes zero rows because status is no longer `placed`. Session 7.

## Non-goals

Real card payments, courier GPS, more than one campus, a second database, GraphQL, and event-sourcing the order. Status is a column. Events are notifications about that column.

## API to freeze

One URL per command. A generic `PATCH /v1/orders/{id}` with a status body has nowhere to hang "who may set `delivered`".

| Caller | Call | Waits? |
|---|---|---|
| student | `POST /v1/orders` | yes |
| student | `GET /v1/orders/{id}` | yes |
| anyone | `GET /v1/restaurants` | yes |
| anyone | `GET /v1/restaurants/{id}/menu` | yes |
| kitchen | `POST /v1/kitchen/orders/{id}/accept` | yes |
| kitchen | `POST /v1/kitchen/orders/{id}/reject` | yes |
| kitchen | `POST /v1/kitchen/orders/{id}/ready` | yes |
| gateway → order | `CreateOrder`, `GetOrder` | yes |
| gateway → kitchen | `AcceptOrder`, `RejectOrder`, `MarkReady` | yes |

`GET /healthz` is not part of this table. It stays unauthenticated. Declare the menu routes in the doc and in OpenAPI. Session 4 implements place and get. Menu handlers can wait.

Replace the empty services in the proto files with those RPCs. `CreateOrder` carries `user_id`, `restaurant_id`, `idempotency_key`, and items (`menu_item_id`, `qty`). It does not carry a price. `GetOrder` carries `order_id`, `caller_id`, and `caller_role`. The public JSON will not send `caller_id`; the gateway will, in session 5. Declaring the fields now means the contract does not change later.

`Order` returns `id`, `user_id`, `restaurant_id`, `status`.

Kitchen RPCs take an `order_id` and return `Order`. You declare them now. You implement them in session 6.

List the same HTTP paths in `api/openapi/gateway.yaml`. Request bodies can wait until session 4. Bearer auth can wait until session 5.

Topic name, for the document only this week: `orders.v1`. Event names: `order.created`, `order.accepted`, `order.rejected`, `order.ready`.

## Check

Read the sequence aloud. You can point at the step that returns HTTP, the step that commits SQL, and the step that may happen after the process dies. The proto RPC names match the table. `docs/architecture.md` has no empty heading.

## What this session is not

A microservice for its own sake, Mongo because the menu is JSON, and a new product. Browse and menu can stay seed data for a long time. The path that has to be true is place, get, and accept.
