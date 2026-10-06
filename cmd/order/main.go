// Package main is the CampusEats order service.
package main

import "log/slog"

// main starts the order gRPC server and the outbox publisher.
// Session 1: load config, log, block until SIGINT/SIGTERM.
// Session 3: open one pgx pool in main and pass it in.
// Session 4: listen on GRPC_ADDR.
// Session 6: run the publisher until the root context is cancelled.
// Session 9: serve /healthz and /readyz on HEALTH_ADDR.
func main() {
	slog.Info("order: process entrypoint is intentionally unfinished")
}
