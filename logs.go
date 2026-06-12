package boxd

import (
	"bufio"
	"context"
	"encoding/binary"
	"io"
	"strings"
	"testing"
)

// WithLogs enables container log streaming.
// LogAlways streams each line to t.Log in real time.
// LogOnFailure buffers logs and dumps them only if the test fails.
func WithLogs(mode LogMode) Option {
	return func(c *config) { c.logMode = &mode }
}

func startLogs(t *testing.T, d *dockerClient, id, image string, lm *LogMode) {
	if lm == nil {
		return
	}

	prefix := "[" + image + "]"

	ctx, cancel := context.WithCancel(context.Background())
	rc, err := d.logs(ctx, id)
	if err != nil {
		cancel()
		t.Log("boxd: could not open log stream:", err)
		return
	}

	if *lm == LogAlways {
		done := streamLines(rc, func(line string) { t.Log(prefix, line) })
		t.Cleanup(func() { cancel(); <-done; rc.Close() })
		return
	}

	var buf []string
	done := streamLines(rc, func(line string) { buf = append(buf, line) })
	t.Cleanup(func() {
		cancel()
		<-done
		rc.Close()
		if t.Failed() {
			for _, line := range buf {
				t.Log(prefix, line)
			}
		}
	})
}

func streamLines(rc io.Reader, fn func(string)) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		defer close(done)
		r := bufio.NewReader(rc)
		for {
			line, err := readDockerLine(r)
			if err != nil {
				return
			}
			fn(line)
		}
	}()
	return done
}

func readDockerLine(r *bufio.Reader) (string, error) {
	hdr := make([]byte, 8)
	if _, err := io.ReadFull(r, hdr); err != nil {
		return "", err
	}
	size := binary.BigEndian.Uint32(hdr[4:])
	line := make([]byte, size)
	if _, err := io.ReadFull(r, line); err != nil {
		return "", err
	}
	return strings.TrimRight(string(line), "\n"), nil
}
