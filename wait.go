package boxd

import (
	"context"
	"fmt"
	"net"
	"time"
)

// WaitStrategy blocks Run until a container is considered ready.
// Implement this interface to define custom readiness checks.
type WaitStrategy interface {
	Wait(c *Container) error
}

// WithWait sets the readiness strategy Run will block on after the container starts.
func WithWait(w WaitStrategy) Option {
	return func(c *config) { c.waitStrat = w }
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

func waitForPort(c *Container, pc portConfig) error {
	if pc.timeout == 0 {
		return nil
	}
	hostPort, ok := c.Ports[pc.port]
	if !ok {
		return fmt.Errorf("boxd: port %s not mapped", pc.port)
	}
	deadline := time.Now().Add(pc.timeout)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", c.Host+":"+hostPort, time.Second)
		if err == nil {
			_ = conn.Close()
			return nil
		}
		time.Sleep(200 * time.Millisecond)
	}
	return fmt.Errorf("boxd: timeout waiting for port %s", pc.port)
}

type statusWait struct {
	timeout  time.Duration
	interval time.Duration
}

// WaitForRunning waits until the container status is "running".
// An optional second argument sets the poll interval (default 200ms).
func WaitForRunning(timeout time.Duration, interval ...time.Duration) WaitStrategy {
	d := 200 * time.Millisecond
	if len(interval) > 0 {
		d = interval[0]
	}
	return &statusWait{timeout: timeout, interval: d}
}

func (w *statusWait) Wait(c *Container) error {
	deadline := time.Now().Add(w.timeout)
	for time.Now().Before(deadline) {
		info, err := c.d.inspect(context.Background(), c.ID)
		if err != nil {
			return err
		}
		if info.State.Status == "running" {
			return nil
		}
		time.Sleep(w.interval)
	}
	return fmt.Errorf("boxd: timeout waiting for running")
}

type healthWait struct {
	timeout  time.Duration
	interval time.Duration
}

// WaitForHealthy waits until Docker reports the container as healthy.
// Requires a healthcheck to be configured via WithHealthCheck.
// An optional second argument sets the poll interval (default 500ms).
func WaitForHealthy(timeout time.Duration, interval ...time.Duration) WaitStrategy {
	d := 500 * time.Millisecond
	if len(interval) > 0 {
		d = interval[0]
	}
	return &healthWait{timeout: timeout, interval: d}
}

func (w *healthWait) Wait(c *Container) error {
	deadline := time.Now().Add(w.timeout)
	for time.Now().Before(deadline) {
		info, err := c.d.inspect(context.Background(), c.ID)
		if err != nil {
			return err
		}
		if info.State.Health.Status == "healthy" {
			return nil
		}
		time.Sleep(w.interval)
	}
	return fmt.Errorf("boxd: timeout waiting for healthy")
}
