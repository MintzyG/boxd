package nameddockerfile_test

import (
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/MintzyG/boxd"
)

// TestDevBuild demonstrates building from a non-default Dockerfile name.
// Useful when a repo has multiple Dockerfiles (e.g. Dockerfile.dev, Dockerfile.prod).
func TestDevBuild(t *testing.T) {
	c := boxd.Run(t,
		boxd.WithDockerfile("app", "Dockerfile.dev"),
		boxd.WithHealthCheck(boxd.HealthCheck{
			Test:     []string{"CMD-SHELL", "wget -qO- http://localhost:8080 || exit 1"},
			Interval: 2 * time.Second,
			Timeout:  1 * time.Second,
			Retries:  5,
		}),
		boxd.WithPort("8080/tcp"),
		boxd.WithWait(boxd.WaitForHealthy(30*time.Second)),
		boxd.WithLogs(boxd.LogAlways),
	)

	resp, err := http.Get("http://" + c.Host + ":" + c.Ports["8080/tcp"])
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(body), "hello from dev build") {
		t.Fatalf("unexpected response: %s", body)
	}
}
