package boxd

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
