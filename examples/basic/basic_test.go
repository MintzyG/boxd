package basic_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/MintzyG/boxd"
)

// TestNginx demonstrates the minimal usage of boxd: pull an image, expose a
// port, wait for it to accept connections, then make a request.
func TestNginx(t *testing.T) {
	c := boxd.Run(t,
		boxd.WithImage("nginx:alpine"),
		boxd.WithHealthCheck(boxd.HealthCheck{
			Test:     []string{"CMD-SHELL", "wget -qO- http://localhost || exit 1"},
			Interval: 2 * time.Second,
			Timeout:  1 * time.Second,
			Retries:  5,
		}),
		boxd.WithPort("80/tcp"),
		boxd.WithWait(boxd.WaitForHealthy(30*time.Second)),
	)

	resp, err := http.Get("http://" + c.Host + ":" + c.Port("80"))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}
