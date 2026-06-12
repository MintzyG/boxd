package boxd

import (
	"context"
	"testing"
)

// Container holds the runtime details of a started container.
// It is returned by Run and is valid for the lifetime of the test.
type Container struct {
	// ID is the Docker container ID.
	ID string
	// Host is the hostname to reach the container on, typically "localhost".
	Host string
	// Ports maps container ports (e.g. "5432/tcp") to their host-side port numbers.
	Ports map[string]string
	d     *dockerClient
}

type config struct {
	image       string
	env         []string
	ports       []portConfig
	waitStrat   WaitStrategy
	logMode     *LogMode
	healthCheck *healthCheck
	build       *buildConfig
	noCache     bool
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

// Run starts a Docker container for the duration of the test.
// The container is removed automatically when the test ends.
// Requires exactly one of WithImage or WithDockerfile.
func Run(t *testing.T, opts ...Option) *Container {
	t.Helper()
	ctx := context.Background()

	cfg := &config{}
	for _, o := range opts {
		o(cfg)
	}

	if cfg.image != "" && cfg.build != nil {
		t.Fatal("boxd: WithImage and WithDockerfile are mutually exclusive")
	}
	if cfg.image == "" && cfg.build == nil {
		t.Fatal("boxd: one of WithImage or WithDockerfile is required")
	}
	if cfg.noCache && cfg.build == nil {
		t.Fatal("boxd: WithNoCache requires WithDockerfile")
	}
	if cfg.build != nil {
		cfg.build.noCache = cfg.noCache
	}

	d := newDockerClient()

	image := cfg.image
	if cfg.build != nil {
		built, err := buildImage(ctx, d, cfg.build)
		if err != nil {
			t.Fatal("build failed:", err)
		}
		image = built
	}

	id := createContainer(t, ctx, d, image, cfg)
	c := inspectContainer(t, ctx, d, id)

	logLabel := image
	if cfg.build != nil {
		logLabel = cfg.build.context + "/" + cfg.build.dockerfile
	}
	startLogs(t, d, id, logLabel, cfg.logMode)

	for _, pc := range cfg.ports {
		if err := waitForPort(c, pc); err != nil {
			_ = d.remove(ctx, id)
			t.Fatalf("port %s never opened: %v", pc.port, err)
		}
	}

	if cfg.waitStrat != nil {
		if err := cfg.waitStrat.Wait(c); err != nil {
			_ = d.remove(ctx, id)
			t.Fatal("wait failed:", err)
		}
	}

	t.Cleanup(func() {
		_ = d.remove(context.Background(), id)
	})

	return c
}
