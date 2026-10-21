// Package main is the CampusEats public HTTP gateway.
package main

import "log/slog"

// main starts the gateway.
// Session 1: see docs/session-01.md. Load config, build a JSON logger, serve GET /healthz, shut down on SIGINT/SIGTERM.
// Session 4: docs/session-04.md. Route /v1 to one order gRPC client for the life of the process.
// Session 5: docs/session-05.md. Authenticate every route except /healthz.
func main() {
	slog.Info("gateway: process entrypoint is intentionally unfinished")
}
