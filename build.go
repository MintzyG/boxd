package boxd

import (
	"archive/tar"
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
)

type buildConfig struct {
	context    string
	dockerfile string
}

func WithDockerfile(contextPath string, dockerfile ...string) Option {
	df := "Dockerfile"
	if len(dockerfile) > 0 {
		df = dockerfile[0]
	}
	return func(c *config) {
		c.build = &buildConfig{context: contextPath, dockerfile: df}
	}
}

func buildImage(ctx context.Context, d *dockerClient, bc *buildConfig) (string, error) {
	archive, err := tarDir(bc.context)
	if err != nil {
		return "", err
	}
	return d.build(ctx, archive, bc.dockerfile)
}

func tarDir(dir string) (io.Reader, error) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

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
		return nil, err
	}
	return &buf, tw.Close()
}
