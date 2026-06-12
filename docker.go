package boxd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
)

type dockerClient struct {
	http    *http.Client
	baseURL string
}

type createBody struct {
	Image       string       `json:"Image"`
	Env         []string     `json:"Env,omitempty"`
	HostConfig  hostConfig   `json:"HostConfig"`
	Healthcheck *healthCheck `json:"Healthcheck,omitempty"`
}

type healthCheck struct {
	Test     []string `json:"Test"`
	Interval int64    `json:"Interval"`
	Timeout  int64    `json:"Timeout"`
	Retries  int      `json:"Retries"`
}

type hostConfig struct {
	PortBindings map[string][]portBinding `json:"PortBindings"`
}

type portBinding struct {
	HostIP   string `json:"HostIp"`
	HostPort string `json:"HostPort"`
}

type createResponse struct {
	ID string `json:"Id"`
}

type inspectResponse struct {
	State struct {
		Status string `json:"Status"`
		Health struct {
			Status string `json:"Status"`
		} `json:"Health"`
	} `json:"State"`
	NetworkSettings struct {
		Ports map[string][]struct {
			HostPort string `json:"HostPort"`
		} `json:"Ports"`
	} `json:"NetworkSettings"`
}

func newDockerClient() *dockerClient {
	socketPath := "/var/run/docker.sock"
	baseURL := "http://docker"

	if host := os.Getenv("DOCKER_HOST"); host != "" {
		switch {
		case strings.HasPrefix(host, "unix://"):
			socketPath = strings.TrimPrefix(host, "unix://")
		case strings.HasPrefix(host, "tcp://"):
			// TCP host: use it directly without a Unix socket transport.
			return &dockerClient{
				http:    &http.Client{},
				baseURL: "http://" + strings.TrimPrefix(host, "tcp://"),
			}
		}
	}

	return &dockerClient{
		baseURL: baseURL,
		http: &http.Client{
			Transport: &http.Transport{
				DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
					return net.Dial("unix", socketPath)
				},
			},
		},
	}
}

func (d *dockerClient) doRaw(ctx context.Context, method, path, contentType string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, d.baseURL+path, body)
	if err != nil {
		return nil, err
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	resp, err := d.http.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		return nil, fmt.Errorf("boxd: %s %s -> %d: %s", method, path, resp.StatusCode, b)
	}
	return resp, nil
}

func (d *dockerClient) do(ctx context.Context, method, path string, body any) (*http.Response, error) {
	var r io.Reader
	var ct string
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		r = bytes.NewReader(b)
		ct = "application/json"
	}
	return d.doRaw(ctx, method, path, ct, r)
}

func (d *dockerClient) pull(ctx context.Context, image string) error {
	name, tag, _ := strings.Cut(image, ":")
	path := "/images/create?fromImage=" + name
	if tag != "" {
		path += "&tag=" + tag
	}
	resp, err := d.do(ctx, http.MethodPost, path, nil)
	if err != nil {
		return err
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	return resp.Body.Close()
}

func (d *dockerClient) create(ctx context.Context, body createBody) (string, error) {
	resp, err := d.do(ctx, http.MethodPost, "/containers/create", body)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result createResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return result.ID, nil
}

func (d *dockerClient) start(ctx context.Context, id string) error {
	resp, err := d.do(ctx, http.MethodPost, "/containers/"+id+"/start", nil)
	if err != nil {
		return err
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	return resp.Body.Close()
}

func (d *dockerClient) inspect(ctx context.Context, id string) (inspectResponse, error) {
	resp, err := d.do(ctx, http.MethodGet, "/containers/"+id+"/json", nil)
	if err != nil {
		return inspectResponse{}, err
	}
	defer resp.Body.Close()

	var result inspectResponse
	return result, json.NewDecoder(resp.Body).Decode(&result)
}

func (d *dockerClient) logs(ctx context.Context, id string) (io.ReadCloser, error) {
	resp, err := d.do(ctx, http.MethodGet, "/containers/"+id+"/logs?follow=true&stdout=1&stderr=1&timestamps=0", nil)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

func (d *dockerClient) build(ctx context.Context, tar io.Reader, dockerfile string, noCache bool) (string, error) {
	path := "/build?dockerfile=" + dockerfile + "&rm=true"
	if noCache {
		path += "&nocache=1"
	}

	resp, err := d.doRaw(ctx, http.MethodPost, path, "application/x-tar", tar)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var imageID string
	dec := json.NewDecoder(resp.Body)
	for {
		var msg struct {
			Stream string `json:"stream"`
			Error  string `json:"error"`
		}
		if err := dec.Decode(&msg); err != nil {
			break
		}
		if msg.Error != "" {
			return "", fmt.Errorf("boxd: build error: %s", msg.Error)
		}
		if after, ok := strings.CutPrefix(msg.Stream, "Successfully built "); ok {
			imageID = strings.TrimSpace(after)
		}
	}
	if imageID == "" {
		return "", fmt.Errorf("boxd: build did not produce an image ID")
	}
	return imageID, nil
}

func (d *dockerClient) remove(ctx context.Context, id string) error {
	resp, err := d.do(ctx, http.MethodDelete, "/containers/"+id+"?force=true", nil)
	if err != nil {
		return err
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	return resp.Body.Close()
}
