package boxd

import "time"

// LogMode controls when container logs are emitted during a test.
type LogMode int

const (
	// LogOnFailure buffers container logs and dumps them only if the test fails.
	LogOnFailure LogMode = iota
	// LogAlways streams container logs to t.Log in real time.
	LogAlways
)

// HealthCheck defines a Docker healthcheck for a container.
// Pass it to WithHealthCheck. Use with WaitForHealthy to block until healthy.
type HealthCheck struct {
	// Test is the command Docker runs to check health, e.g. ["CMD-SHELL", "pg_isready"].
	Test []string
	// Interval is how often Docker runs the check.
	Interval time.Duration
	// Timeout is the maximum time Docker waits for a single check to complete.
	Timeout time.Duration
	// Retries is the number of consecutive failures before the container is unhealthy.
	Retries int
}
