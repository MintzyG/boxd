package boxd

import (
	"context"
	"strings"
	"testing"
)

// normalizePort appends "/tcp" if port has no protocol suffix.
func normalizePort(port string) string {
	if strings.Contains(port, "/") {
		return port
	}
	return port + "/tcp"
}

// Port returns the host-side port mapped to the given container port.
// Accepts a bare number ("5432") or a full port/protocol pair ("5432/tcp").
// Returns an empty string if the port was not mapped.
func (c *Container) Port(port string) string {
	return c.Ports[normalizePort(port)]
}

func createContainer(t *testing.T, ctx context.Context, d *dockerClient, image string, cfg *config) string {
	t.Helper()

	if cfg.build == nil {
		if err := d.pull(ctx, image); err != nil {
			t.Fatal("pull failed:", err)
		}
		t.Log("pulled", image)
	}

	bindings := map[string][]portBinding{}
	for _, pc := range cfg.ports {
		bindings[pc.port] = []portBinding{{HostIP: "0.0.0.0", HostPort: ""}}
	}

	id, err := d.create(ctx, createBody{
		Image:       image,
		Env:         cfg.env,
		HostConfig:  hostConfig{PortBindings: bindings},
		Healthcheck: cfg.healthCheck,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log("created", id)

	if err = d.start(ctx, id); err != nil {
		t.Fatal(err)
	}
	t.Log("started", id)

	return id
}

func inspectContainer(t *testing.T, ctx context.Context, d *dockerClient, id string) *Container {
	t.Helper()

	info, err := d.inspect(ctx, id)
	if err != nil {
		t.Fatal(err)
	}

	ports := map[string]string{}
	for port, bindings := range info.NetworkSettings.Ports {
		if len(bindings) > 0 {
			ports[port] = bindings[0].HostPort
		}
	}

	return &Container{ID: id, Host: d.host, Ports: ports, d: d}
}
