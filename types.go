package boxd

import "time"

type LogMode int

const (
	LogOnFailure LogMode = iota
	LogAlways
)

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

type HealthCheck struct {
	Test     []string
	Interval time.Duration
	Timeout  time.Duration
	Retries  int
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
