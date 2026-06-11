package boxd

import (
	"context"
	"fmt"
	"net"
	"time"
)

type WaitStrategy interface {
	Wait(c *Container) error
}

func WithWait(w WaitStrategy) Option {
	return func(c *config) { c.waitStrat = w }
}

type portConfig struct {
	port    string
	timeout time.Duration
}

func waitForPort(c *Container, pc portConfig) error {
	if pc.timeout == 0 {
		return nil
	}
	hostPort, ok := c.Ports[pc.port]
	if !ok {
		return fmt.Errorf("port %s not mapped", pc.port)
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
	return fmt.Errorf("timeout waiting for port %s", pc.port)
}

type statusWait struct {
	timeout time.Duration
}

func WaitForRunning(timeout time.Duration) WaitStrategy {
	return &statusWait{timeout: timeout}
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
		time.Sleep(200 * time.Millisecond)
	}
	return fmt.Errorf("timeout waiting for running")
}

type healthWait struct {
	timeout time.Duration
}

func WaitForHealthy(timeout time.Duration) WaitStrategy {
	return &healthWait{timeout: timeout}
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
		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("timeout waiting for healthy")
}
