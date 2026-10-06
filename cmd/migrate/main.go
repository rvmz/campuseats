// Package main applies CampusEats SQL migrations and exits.
package main

import "log/slog"

// main runs goose once against DATABASE_URL.
// Session 8: embed the migrations directory, apply it, and exit 0.
// Compose treats this container as a one-shot, not a long-running service.
func main() {
	slog.Info("migrate: process entrypoint is intentionally unfinished")
}
