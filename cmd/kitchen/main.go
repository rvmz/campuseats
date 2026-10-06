// Package main is the CampusEats kitchen service.
package main

import "log/slog"

// main starts the kitchen consumer.
// Session 1: load config, log, block until SIGINT/SIGTERM.
// Session 6: consume orders.v1 in group "kitchen" and log order.created.
// Session 7: accept only from status placed; commit the Kafka offset after a successful handle.
func main() {
	slog.Info("kitchen: process entrypoint is intentionally unfinished")
}
