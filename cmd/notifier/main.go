// Package main is the CampusEats notifier.
package main

import "log/slog"

// main starts the notifier consumer.
// Session 1: load config, log, block until SIGINT/SIGTERM.
// Session 6: consume orders.v1 in group "notifier" and log a fake push.
func main() {
	slog.Info("notifier: process entrypoint is intentionally unfinished")
}
