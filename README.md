# boxd

boxd is a minimal Go library for running Docker containers in tests. It aims to do what [testcontainers-go](https://github.com/testcontainers/testcontainers-go) does, but with zero dependencies and a small, auditable surface area.

If you want container-based integration tests without pulling in a large dependency tree, boxd is for you.

## Install

```
go get github.com/MintzyG/boxd
```

## Quick start

```go
func TestMyService(t *testing.T) {
    c := boxd.Run(t,
        boxd.WithImage("redis:7-alpine"),
        boxd.WithPort("6379", 10*time.Second),
    )

    addr := c.Host + ":" + c.Port("6379")
    // connect, run your tests, container is removed when the test ends
}
```

`Run` pulls the image, starts the container, waits for the port, and registers cleanup, all in one call.

## Extending boxd

boxd is designed to be wrapped. If your tests use the same service repeatedly, build a typed helper around it:

```go
// myredis/myredis.go
type RedisContainer struct{ *boxd.Container }

func Run(t *testing.T) *RedisContainer {
    return &RedisContainer{boxd.Run(t,
        boxd.WithImage("redis:7-alpine"),
        boxd.WithPort("6379", 10*time.Second),
    )}
}

func (c *RedisContainer) Addr() string {
    return c.Host + ":" + c.Port("6379")
}
```

See [`examples/postgres`](examples/postgres) for a full example with `ConnStr()`, `DB()`, and `Clean()` helpers.

## Options

| Option                                | Description                                                                                                                 |
|---------------------------------------|-----------------------------------------------------------------------------------------------------------------------------|
| `WithImage(image)`                    | Pull and run an image. Mutually exclusive with `WithDockerfile`.                                                            |
| `WithDockerfile(ctx, ...BuildOption)` | Build from a local Dockerfile and run it. Accepts `WithDockerfileName` and `WithNoCache` as build options.                  |
| `WithDockerfileName(name)`            | Override the Dockerfile name (default `"Dockerfile"`). Pass as a `BuildOption` to `WithDockerfile`.                         |
| `WithNoCache()`                       | Disable Docker layer cache. Pass as a `BuildOption` to `WithDockerfile`.                                                    |
| `WithEnv(k, v)`                       | Set an environment variable.                                                                                                |
| `WithPort(port, ...timeout)`          | Expose a port (`"5432"` or `"5432/udp"`). Bare numbers default to TCP. Optional timeout waits until it accepts connections. |
| `WithHealthCheck(HealthCheck{...})`   | Attach a Docker healthcheck.                                                                                                |
| `WithWait(strategy)`                  | Block until a readiness strategy passes.                                                                                    |
| `WithLogs(mode)`                      | Stream logs. `LogAlways` streams in real time, `LogOnFailure` dumps on test failure.                                        |

## Container methods

| Method                  | Description                                                                                  |
|-------------------------|----------------------------------------------------------------------------------------------|
| `Port(port)`            | Returns the host-side port mapped to the given container port. Returns `""` if not mapped.   |
| `MustPort(t, port)`     | Like `Port`, but calls `t.Fatal` if the port is not mapped.                                  |

## Wait strategies

| Strategy                  | Description                                                                  |
|---------------------------|------------------------------------------------------------------------------|
| `WaitForHealthy(timeout)` | Wait until Docker reports the container healthy. Requires `WithHealthCheck`. |
| `WaitForRunning(timeout)` | Wait until the container status is `running`.                                |

## Examples

- [`examples/basic`](examples/basic) - pull an image and hit an HTTP endpoint
- [`examples/dockerfile`](examples/dockerfile) - build and run a local Dockerfile
- [`examples/named-dockerfile`](examples/named-dockerfile) - build from a non-default Dockerfile name
- [`examples/postgres`](examples/postgres) - typed Postgres wrapper with `Clean()` and `DB()` helpers