package boxd_test

import (
	"testing"

	"github.com/MintzyG/boxd"
)

func TestBuild(t *testing.T) {
	boxd.Run(t,
		boxd.WithDockerfile("testdata/hello"),
		boxd.WithLogs(boxd.LogAlways),
	)
}
