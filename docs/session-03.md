# Session 3 — The order is one transaction

Session 2 said Postgres is the source of truth. This session makes that a table you can roll back.

By the end:

- `goose up` creates the schema and `goose down` removes it
- money is integer cents, and `orders.status` only allows the names in `internal/order/status.go`
- `orders.idempotency_key` and an empty `outbox` table exist
- inserting an order and its items is one transaction
- a bad item leaves no order row
- the line price is copied from `menu_items` at insert time
- `scripts/seed.sql` loads the fixed users, restaurant, and two dishes

Kafka is not running. Do not publish anything from this session.

## Why a transaction

`Create` is three writes that mean one thing: the order, its lines, and (in session 6) the outbox row. If the second line fails, the order insert must not remain. Otherwise the student has an id for a lunch that has no food, or food for an order that does not exist.

The outbox table is created now and left empty. The insert into it belongs in this same transaction later. Do not write a publisher, and do not insert the outbox row from a second connection after commit.

## Postgres on your laptop

The user and database match `.env.example`. The default `postgres` superuser is a different URL, and goose will look logged in while talking to the wrong role.

```bash
docker run -d --name campuseats-pg \
  -e POSTGRES_USER=campuseats \
  -e POSTGRES_PASSWORD=campuseats \
  -e POSTGRES_DB=campuseats \
  -p 5432:5432 postgres:16
```

```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
```

```bash
export DATABASE_URL=postgres://campuseats:campuseats@localhost:5432/campuseats?sslmode=disable
goose -dir migrations postgres "$DATABASE_URL" up
```

`Down` drops tables in the reverse of `Up`: outbox, order items, orders, menu items, restaurants, users. A migration that has already been shared is history. Until this file has been applied anywhere that matters, you may still edit `001_init.sql`. After that, the next change is `002`.

## Schema

Replace `migrations/001_init.sql` with the schema below. `gen_random_uuid()` is built into Postgres 16. Prices are `integer` cents, never `float` or `numeric`. The status check uses the same strings as `internal/order/status.go`. `idempotency_key` stays nullable until session 7, which is when it becomes required and unique per user.

```sql
-- +goose Up
create table users (
    id uuid primary key default gen_random_uuid(),
    email text not null unique,
    role text not null check (role in ('student', 'kitchen', 'courier')),
    created_at timestamptz not null default now()
);

create table restaurants (
    id uuid primary key default gen_random_uuid(),
    name text not null,
    campus text not null,
    open boolean not null default true
);

create table menu_items (
    id uuid primary key default gen_random_uuid(),
    restaurant_id uuid not null references restaurants (id),
    name text not null,
    price_cents integer not null check (price_cents >= 0),
    available boolean not null default true
);

create table orders (
    id uuid primary key default gen_random_uuid(),
    user_id uuid not null references users (id),
    restaurant_id uuid not null references restaurants (id),
    status text not null default 'placed' check (status in (
        'placed', 'accepted', 'preparing', 'ready',
        'picked_up', 'delivered', 'rejected', 'cancelled'
    )),
    idempotency_key text,
    created_at timestamptz not null default now()
);

create table order_items (
    order_id uuid not null references orders (id),
    menu_item_id uuid not null references menu_items (id),
    qty integer not null check (qty > 0),
    price_cents integer not null check (price_cents >= 0),
    primary key (order_id, menu_item_id)
);

create table outbox (
    id bigserial primary key,
    aggregate_id uuid not null,
    event_type text not null,
    payload jsonb not null,
    created_at timestamptz not null default now(),
    published_at timestamptz
);

create index orders_user_created on orders (user_id, created_at desc);
create index outbox_unpublished on outbox (created_at) where published_at is null;

-- +goose Down
drop table outbox;
drop table order_items;
drop table orders;
drop table menu_items;
drop table restaurants;
drop table users;
```

`orders_user_created` is "my orders". `outbox_unpublished` is the publisher's queue. There is no password column. Session 5 uses a signed token, not a login form.

## Queries

`sqlc.yaml` already points at `queries/` and writes Go into `internal/gen/sqlc`. UUID columns are strings. A nullable column such as `idempotency_key` is a pointer.

Put this in `queries/orders.sql`:

```sql
-- name: CreateOrder :one
insert into orders (user_id, restaurant_id, status, idempotency_key)
values ($1, $2, 'placed', $3)
returning id, user_id, restaurant_id, status, created_at;

-- name: GetMenuItem :one
select id, restaurant_id, price_cents, available
from menu_items
where id = $1;

-- name: CreateOrderItem :exec
insert into order_items (order_id, menu_item_id, qty, price_cents)
values ($1, $2, $3, $4);

-- name: GetOrder :one
select id, user_id, restaurant_id, status, created_at
from orders
where id = $1;
```

Name the columns. `select *` breaks the generated code the day a column is added.

```bash
go get github.com/jackc/pgx/v5
sqlc generate
```

Commit `internal/gen/sqlc`. A classmate who clones the repo should not need `sqlc` on the path to build.

## The store

Open one pool in `cmd/order`, from `DATABASE_URL`, and pass it in. Do not open a pool inside a function that creates one order.

`Create` reads the menu row inside the transaction and copies `price_cents`. The client does not send a price. If the menu item is missing, sold by another restaurant, or not available, return an error before commit. `defer tx.Rollback(ctx)` is safe after a successful commit: pgx returns `ErrTxClosed`.

```go
func (s *Store) Create(ctx context.Context, in CreateParams) (Order, error) {
    tx, err := s.pool.Begin(ctx)
    if err != nil {
        return Order{}, err
    }
    defer tx.Rollback(ctx)

    q := s.q.WithTx(tx)
    row, err := q.CreateOrder(ctx, sqlc.CreateOrderParams{
        UserID:         in.UserID,
        RestaurantID:   in.RestaurantID,
        IdempotencyKey: in.IdempotencyKey, // *string; nil when the request has no key
    })
    if err != nil {
        return Order{}, err
    }
    for _, item := range in.Items {
        menu, err := q.GetMenuItem(ctx, item.MenuItemID)
        if err != nil {
            return Order{}, err
        }
        if menu.RestaurantID != in.RestaurantID || !menu.Available {
            return Order{}, fmt.Errorf("menu item %s is not sold here", item.MenuItemID)
        }
        if err := q.CreateOrderItem(ctx, sqlc.CreateOrderItemParams{
            OrderID:    row.ID,
            MenuItemID: item.MenuItemID,
            Qty:        item.Qty,
            PriceCents: menu.PriceCents,
        }); err != nil {
            return Order{}, err
        }
    }
    // Session 6: insert the outbox row here, before commit.
    if err := tx.Commit(ctx); err != nil {
        return Order{}, err
    }
    return toOrder(row), nil
}
```

`CreateOrderParams` will match what `sqlc generate` printed. Use those field names if they differ by a word.

## Prove the rollback

Apply the migration, then run `scripts/seed.sql`:

```bash
psql "$DATABASE_URL" -f scripts/seed.sql
```

Seed is not a goose migration. Goose on a fresh volume would re-insert those rows in every environment, including one you do not want seeded.

Create an order whose second item id does not exist. `select count(*) from orders` stays 0. Create an order for Miso ramen (`bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb`). The line's `price_cents` is 950, which is the menu price, not a number you typed into the request.

## Check

`goose down` then `goose up` on a database you can throw away. One failed item insert, zero orders. `outbox` has no rows. `internal/gen/sqlc` is committed.

## Homework

`GetOrder` is already a query. Call it from a small test or a temporary main. In `docs/architecture.md`, one sentence: the outbox insert will join this transaction, not follow it.

## What this session is not

An ORM. The SQL is the part you are grading yourself on. A JSON column for items would drop the foreign key to the menu. A cancelled order is a status, not a deleted row.
