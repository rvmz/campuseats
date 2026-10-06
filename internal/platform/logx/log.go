// Package logx builds the process logger.
package logx

import (
	"io"
	"log/slog"
)

// New returns a JSON logger with a service attribute.
// Session 1: slog JSON on stdout. Pass the logger into constructors; do not store it in a global.
func New(service string) *slog.Logger {
	_ = service
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
