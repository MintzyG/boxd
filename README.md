# boxd

Lightweight library for running Docker containers in Go tests. No dependencies.

## Install

```
go get github.com/MintzyG/boxd
```

## Usage

```go
c := boxd.Run(t,
    boxd.WithImage("postgres:16"),
    boxd.WithEnv("POSTGRES_PASSWORD", "pass"),
    boxd.WithHealthCheck(boxd.HealthCheck{
        Test:     []string{"CMD-SHELL", "pg_isready -U postgres"},
        Interval: 2 * time.Second,
        Timeout:  1 * time.Second,
        Retries:  5,
    }),
    boxd.WithPort("5432/tcp"),
    boxd.WithWait(boxd.WaitForHealthy(30*time.Second)),
)

// c.Host, c.Ports["5432/tcp"] are available after Run returns
// container is removed automatically when the test ends
```

### Build from a local Dockerfile

```go
boxd.Run(t,
    boxd.WithDockerfile("./path/to/context"),
    boxd.WithPort("8080/tcp", 15*time.Second),
)
```

Use `WithDockerfile(ctx, "Custom.dockerfile")` to specify a non-default filename.

## Options

| Option                               | Description                                                        |
|--------------------------------------|--------------------------------------------------------------------|
| `WithImage(image)`                   | Pull and run an image                                              |
| `WithDockerfile(ctx, ...dockerfile)` | Build from a local Dockerfile and run it                           |
| `WithEnv(k, v)`                      | Set an environment variable                                        |
| `WithPort(port, ...timeout)`         | Expose a port; optional timeout waits until it accepts connections |
| `WithHealthCheck(HealthCheck{...})`  | Attach a Docker healthcheck                                        |
| `WithWait(strategy)`                 | Block until a wait strategy passes                                 |
| `WithLogs(mode)`                     | Stream container logs (`LogAlways` or `LogOnFailure`)              |

`WithImage` and `WithDockerfile` are mutually exclusive.

## Wait strategies

| Strategy                  | Description                                     |
|---------------------------|-------------------------------------------------|
| `WaitForHealthy(timeout)` | Wait until Docker reports the container healthy |
| `WaitForRunning(timeout)` | Wait until the container status is running      |