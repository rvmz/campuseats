# Session 1 — A process that can start and stop

CampusEats is four programs. This session makes the first one, `gateway`, behave like a process you could run under Docker or Kubernetes later. The other three programs stay stubs. Copying this shape onto them is the homework.

By the end, `gateway`:

- reads its listen address from the environment
- writes JSON logs
- answers `GET /healthz` with `200` and `{"status":"ok"}`
- stops on Ctrl-C, and gives in-flight requests up to 10 seconds to finish
- has a test that does not open a port

No database, no framework, no config file.

## Where the code lives

A **process** is one `main` function, one binary. Each directory under `cmd/` is one process:

| Directory | Process |
|---|---|
| `cmd/gateway` | HTTP API the phones call |
| `cmd/order` | orders, Postgres, outbox |
| `cmd/kitchen` | kitchen staff and the Kafka consumer |
| `cmd/notifier` | the push log |

`internal/` is libraries those processes import. Another module cannot import `internal/`. That is the whole rule. Configuration and logging are shared, so they live in `internal/platform`. HTTP routes for the gateway live in `internal/gateway`, which you create in this session.

`main` is the only place that decides to exit. A library returns an error.

## Configuration

The process learns how to run from environment variables. A YAML file would be another thing to mount later. Kubernetes will set the same variables.

`HTTP_ADDR` defaults to `:8080` because every laptop can use that. A missing database URL must not get a made-up default: an empty string and "use localhost" are different situations, and a silent default hides a broken deploy. Session 1 does not need a database. Leave `DatabaseURL` empty.

`os.Getenv` cannot tell "unset" from "set to empty". Use `os.LookupEnv` when the distinction matters. For the listen address, empty and unset both mean "use the default".

Replace `Load` in `internal/platform/config/config.go`:

```go
func Load() (Config, error) {
    cfg := Config{
        HTTPAddr: getenv("HTTP_ADDR", ":8080"),
    }
    return cfg, nil
}

func getenv(key, fallback string) string {
    if v, ok := os.LookupEnv(key); ok && v != "" {
        return v
    }
    return fallback
}
```

Call `Load` from `main`. If it returns an error, log that error and exit 1. Do not continue on a zero `Config`.

## Logs

Use `log/slog` with a JSON handler on stdout. JSON has stable keys, so a collector in a later session can index `service` and `err` without parsing a sentence.

```go
func New(service string) *slog.Logger {
    return slog.New(slog.NewJSONHandler(os.Stdout, nil)).With("service", service)
}
```

Pass `*slog.Logger` into what you construct. A package-level `var logger` is shared with tests and is awkward to replace. The standard library `log` package has no levels and no fields. Session 11 will add `trace_id` as a field. Start with a logger that accepts fields.

`gateway` passes `"gateway"`. The other three processes will pass their own name, which is how you tell their lines apart.

## Health

Create `internal/gateway` with a mux and one route.

`GET /healthz` returns 200 and this body:

```json
{"status":"ok"}
```

It must not ping Postgres. This endpoint means "the process is alive". If it checked the database, a short database outage would kill a process that was fine, and a cluster would restart every copy at once. A second endpoint, `/readyz`, will mean "send me traffic" and will ping the database. That is session 9. Today there is only liveness.

```go
func NewMux() http.Handler {
    mux := http.NewServeMux()
    mux.HandleFunc("GET /healthz", live)
    return mux
}

func live(w http.ResponseWriter, _ *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    _ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
```

## Shutdown

`http.ListenAndServe` blocks until the process is killed. A kill drops the request that was in progress. Kubernetes sends SIGTERM and then waits. The process has to:

1. stop accepting new connections
2. finish or abandon the ones it already accepted
3. return from `main`

`context.Context` carries that cancellation. It is not a place to store configuration. Config is an argument. The root context is cancelled when the operating system asks the process to stop.

Order in `cmd/gateway/main.go`:

1. `signal.NotifyContext` cancels on SIGINT and SIGTERM.
2. `ListenAndServe` runs in a goroutine, because it blocks.
3. `main` waits on `<-ctx.Done()`.
4. `Shutdown` gets a **new** context with a 10 second timeout. The root context is already cancelled. Passing it to `Shutdown` makes `Shutdown` return immediately, which skips the wait.
5. Return from `main`. Deferred calls run. `os.Exit` skips them.

`ReadHeaderTimeout` limits how long the server waits for request headers, so a client that connects and sends nothing cannot hold a goroutine open. Five seconds is enough for this lab.

```go
func main() {
    cfg, err := config.Load()
    if err != nil {
        slog.Error("config", "err", err)
        os.Exit(1)
    }
    log := logx.New("gateway")

    srv := &http.Server{
        Addr:              cfg.HTTPAddr,
        Handler:           gateway.NewMux(),
        ReadHeaderTimeout: 5 * time.Second,
    }

    ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
    defer stop()

    go func() {
        log.Info("listening", "addr", cfg.HTTPAddr)
        if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
            log.Error("listen", "err", err)
            stop()
        }
    }()

    <-ctx.Done()
    shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    if err := srv.Shutdown(shutCtx); err != nil {
        log.Error("shutdown", "err", err)
    }
}
```

If the port is already taken, `ListenAndServe` returns an error that is not `http.ErrServerClosed`. Calling `stop()` unblocks `main` so the process exits instead of waiting forever for a signal.

## Test

`httptest` calls the handler without binding a port. Assert the status and the body.

```go
func TestHealth(t *testing.T) {
    req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
    rec := httptest.NewRecorder()
    NewMux().ServeHTTP(rec, req)
    if rec.Code != http.StatusOK {
        t.Fatalf("status %d", rec.Code)
    }
    if !bytes.Contains(rec.Body.Bytes(), []byte(`"status":"ok"`)) {
        t.Fatalf("body %s", rec.Body.Bytes())
    }
}
```

Put that in `internal/gateway/health_test.go`.

## Check

```bash
go test ./...
HTTP_ADDR=:8080 go run ./cmd/gateway
```

In another terminal:

```bash
curl -sS -D- localhost:8080/healthz
```

You want `200` and `{"status":"ok"}`. Ctrl-C in the first terminal should return within 10 seconds. The log line is JSON and contains `"service":"gateway"`.

## Homework

`cmd/order`, `cmd/kitchen`, and `cmd/notifier` still exit immediately. Give each the same start and stop shape: load config, log `service` as its own name, and block until SIGINT or SIGTERM. They do not listen on HTTP yet. A process that exits 0 on startup looks healthy and does nothing.

Read `docs/architecture.md` before session 2. The headings are the worksheet, and [session-02.md](session-02.md) is the write-up.

## What this session is not

Gin, Zap, Viper, and Wire are real tools. This course stays on the standard library here so the timeout, the logger, and the shutdown stay visible. Request-scoped data, such as the caller identity, goes on a context in session 5. Process config does not.
