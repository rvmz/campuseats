// Package config loads process configuration from the environment.
package config

import "errors"

// Config is the runtime configuration shared by CampusEats processes.
// A field stays zero until the session that needs it.
type Config struct {
	// HTTPAddr is the gateway listen address, for example ":8080".
	HTTPAddr string
	// GRPCAddr is the order service listen address.
	GRPCAddr string
	// HealthAddr is the probe and metrics address on processes that are not the gateway.
	HealthAddr string
	// OrderAddr is where the gateway dials the order service, for example "order:9090".
	OrderAddr string
	// DatabaseURL is a Postgres DSN. Session 3.
	DatabaseURL string
	// KafkaBrokers is a comma-separated broker list. Session 6.
	KafkaBrokers string
	// JWTSecret signs and verifies access tokens. Session 5.
	JWTSecret string
	// PublishEnabled turns the outbox publisher on. Session 6 uses a false value to demo a crash.
	PublishEnabled bool
	// ReceiptsBucket is the S3 bucket for receipts. Empty disables upload. Session 10.
	ReceiptsBucket string
	// AWSRegion is the region for the AWS SDK.
	AWSRegion string
}

// Load reads configuration from the environment.
// Session 1: default HTTP_ADDR to ":8080" and return a Config.
// Later sessions: read the fields above. Missing JWT_SECRET and DATABASE_URL
// are errors in the process that requires them, not in every process.
func Load() (Config, error) {
	return Config{}, errors.New("config: Load not implemented")
}
