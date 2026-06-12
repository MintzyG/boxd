package boxd

import (
	"archive/tar"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type buildConfig struct {
	context    string
	dockerfile string
	noCache    bool
}

// BuildOption configures a Dockerfile build.
type BuildOption func(*buildConfig)

// WithDockerfileName overrides the Dockerfile name within the build context (default "Dockerfile").
func WithDockerfileName(name string) BuildOption {
	return func(bc *buildConfig) { bc.dockerfile = name }
}

// WithNoCache disables Docker's layer cache for this build.
func WithNoCache() BuildOption {
	return func(bc *buildConfig) { bc.noCache = true }
}

// WithDockerfile builds an image from a local Dockerfile and runs it.
// contextPath is the build context directory. Defaults to "Dockerfile".
// Mutually exclusive with WithImage.
func WithDockerfile(contextPath string, opts ...BuildOption) Option {
	return func(c *config) {
		bc := &buildConfig{context: contextPath, dockerfile: "Dockerfile"}
		for _, o := range opts {
			o(bc)
		}
		c.build = bc
	}
}

func buildImage(ctx context.Context, d *dockerClient, bc *buildConfig) (string, error) {
	if _, err := os.Stat(bc.context); err != nil {
		return "", fmt.Errorf("boxd: build context not found: %s", bc.context)
	}
	return d.build(ctx, tarDir(bc.context), bc.dockerfile, bc.noCache)
}

func tarDir(dir string) io.Reader {
	pr, pw := io.Pipe()
	go func() {
		tw := tar.NewWriter(pw)
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			hdr, err := tar.FileInfoHeader(info, "")
			if err != nil {
				return err
			}
			hdr.Name, err = filepath.Rel(dir, path)
			if err != nil {
				return err
			}
			if err = tw.WriteHeader(hdr); err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			f, err := os.Open(path)
			if err != nil {
				return err
			}
			defer f.Close()
			_, err = io.Copy(tw, f)
			return err
		})
		if err != nil {
			pw.CloseWithError(err)
			return
		}
		pw.CloseWithError(tw.Close())
	}()
	return pr
}
