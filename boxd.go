package boxd

import (
	"context"
	"testing"
	"time"
)

type Container struct {
	ID    string
	Host  string
	Ports map[string]string
	d     *dockerClient
}

type config struct {
	image       string
	env         []string
	ports       []portConfig
	waitStrat   WaitStrategy
	logMode     *logConfig
	healthCheck *healthCheck
	build       *buildConfig
}

type Option func(*config)

func WithImage(image string) Option {
	return func(c *config) { c.image = image }
}

func WithEnv(k, v string) Option {
	return func(c *config) { c.env = append(c.env, k+"="+v) }
}

func WithPort(port string, timeout ...time.Duration) Option {
	var t time.Duration
	if len(timeout) > 0 {
		t = timeout[0]
	}
	return func(c *config) { c.ports = append(c.ports, portConfig{port: port, timeout: t}) }
}

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

	startLogs(t, d, id, image, cfg.logMode)

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
