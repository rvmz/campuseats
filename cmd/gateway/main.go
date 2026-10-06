// Package main is the CampusEats public HTTP gateway.
package main

import "log/slog"

// main starts the gateway.
// Session 1: see docs/session-01.md. Load config, build a JSON logger, serve GET /healthz, shut down on SIGINT/SIGTERM.
// Session 4: route /v1 to the order gRPC client. One client for the life of the process.
// Session 5: authenticate every route except /healthz and /metrics.
func main() {
	slog.Info("gateway: process entrypoint is intentionally unfinished")
}
