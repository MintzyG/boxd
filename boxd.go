package boxd

import (
	"context"
	"testing"
)

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
		ctx := context.Background()
		_ = d.remove(ctx, id)
		if cfg.build != nil {
			_ = d.removeImage(ctx, image)
		}
	})

	return c
}
