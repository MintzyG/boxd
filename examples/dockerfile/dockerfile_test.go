package dockerfile_test

import (
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/MintzyG/boxd"
)

// TestLocalApp demonstrates building an image from a local Dockerfile and
// running the resulting container. No registry involved.
func TestLocalApp(t *testing.T) {
	c := boxd.Run(t,
		boxd.WithDockerfile("app"),
		boxd.WithHealthCheck(boxd.HealthCheck{
			Test:     []string{"CMD-SHELL", "wget -qO- http://localhost:8080 || exit 1"},
			Interval: 2 * time.Second,
			Timeout:  1 * time.Second,
			Retries:  5,
		}),
		boxd.WithPort("8080"),
		boxd.WithWait(boxd.WaitForHealthy(30*time.Second)),
		boxd.WithLogs(boxd.LogAlways),
	)

	resp, err := http.Get("http://" + c.Host + ":" + c.Port("8080"))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(body), "hello from boxd") {
		t.Fatalf("unexpected response: %s", body)
	}
}
