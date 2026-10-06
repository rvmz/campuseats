// Package order is the order lifecycle: statuses, transitions, and persistence.
package order

// Status values stored in orders.status.
// Session 2: keep docs/architecture.md in sync if you rename any of these.
const (
	StatusPlaced    = "placed"
	StatusAccepted  = "accepted"
	StatusPreparing = "preparing"
	StatusReady     = "ready"
	StatusPickedUp  = "picked_up"
	StatusDelivered = "delivered"
	StatusRejected  = "rejected"
	StatusCancelled = "cancelled"
)

// CanTransition reports whether an order may move from one status to another.
// Session 7: implement the edges. Commands from staff use this and return an
// error when the jump is illegal. Replayed Kafka events no-op instead of failing
// the consumer when zero rows change.
//
// Edges to encode:
//
//	placed    → accepted, rejected, cancelled
//	accepted  → preparing, cancelled
//	preparing → ready
//	ready     → picked_up
//	picked_up → delivered
func CanTransition(from, to string) bool {
	_, _ = from, to
	return false
}
