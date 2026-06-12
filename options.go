package boxd

import "time"

type config struct {
	image       string
	env         []string
	ports       []portConfig
	waitStrat   WaitStrategy
	logMode     *LogMode
	healthCheck *healthCheck
	build       *buildConfig
}

// Option configures a container before it is started.
type Option func(*config)

// WithImage sets the Docker image to pull and run.
// Mutually exclusive with WithDockerfile.
func WithImage(image string) Option {
	return func(c *config) { c.image = image }
}

// WithEnv sets an environment variable on the container.
func WithEnv(k, v string) Option {
	return func(c *config) { c.env = append(c.env, k+"="+v) }
}

// WithHealthCheck attaches a Docker healthcheck to the container.
// Use with WaitForHealthy to block until the container reports healthy.
func WithHealthCheck(hc HealthCheck) Option {
	return func(c *config) {
		c.healthCheck = &healthCheck{
			Test:     hc.Test,
			Interval: hc.Interval.Nanoseconds(),
			Timeout:  hc.Timeout.Nanoseconds(),
			Retries:  hc.Retries,
		}
	}
}

type portConfig struct {
	port    string
	timeout time.Duration
}

// WithPort exposes a container port and maps it to a random host port.
// Accepts a bare number ("5432") which defaults to TCP, or a full pair ("5432/udp").
// If a timeout is given, Run will block until the port accepts connections
// or the timeout expires. Without a timeout, the port is mapped but not waited on.
func WithPort(port string, timeout ...time.Duration) Option {
	var t time.Duration
	if len(timeout) > 0 {
		t = timeout[0]
	}
	return func(c *config) { c.ports = append(c.ports, portConfig{port: normalizePort(port), timeout: t}) }
}
