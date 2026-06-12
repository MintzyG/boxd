package boxd

import (
	"archive/tar"
	"context"
	"io"
	"os"
	"path/filepath"
)

type buildConfig struct {
	context    string
	dockerfile string
	noCache    bool
}

// WithDockerfile builds an image from a local Dockerfile and runs it.
// contextPath is the build context directory. dockerfile defaults to "Dockerfile"
// and can be overridden by passing a second argument.
// Mutually exclusive with WithImage.
func WithDockerfile(contextPath string, dockerfile ...string) Option {
	df := "Dockerfile"
	if len(dockerfile) > 0 {
		df = dockerfile[0]
	}
	return func(c *config) {
		c.build = &buildConfig{context: contextPath, dockerfile: df}
	}
}

// WithNoCache disables Docker's build cache for this build.
// Must be used together with WithDockerfile; has no effect with WithImage and Run will fatal.
func WithNoCache() Option {
	return func(c *config) { c.noCache = true }
}

func buildImage(ctx context.Context, d *dockerClient, bc *buildConfig) (string, error) {
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
