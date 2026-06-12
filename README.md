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
        boxd.WithPort("6379/tcp", 10*time.Second),
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
        boxd.WithPort("6379/tcp", 10*time.Second),
    )}
}

func (c *RedisContainer) Addr() string {
    return c.Host + ":" + c.Port("6379")
}
```

See [`examples/postgres`](examples/postgres) for a full example with `ConnStr()`, `DB()`, and `Clean()` helpers.

## Options

| Option                              | Description                                                                          |
|-------------------------------------|--------------------------------------------------------------------------------------|
| `WithImage(image)`                  | Pull and run an image. Mutually exclusive with `WithDockerfile`.                     |
| `WithDockerfile(ctx, ...name)`      | Build from a local Dockerfile and run it. Defaults to `"Dockerfile"`.                |
| `WithNoCache()`                     | Disable Docker layer cache for the build.                                            |
| `WithEnv(k, v)`                     | Set an environment variable.                                                         |
| `WithPort(port, ...timeout)`        | Expose a port. Optional timeout waits until it accepts TCP connections.              |
| `WithHealthCheck(HealthCheck{...})` | Attach a Docker healthcheck.                                                         |
| `WithWait(strategy)`                | Block until a readiness strategy passes.                                             |
| `WithLogs(mode)`                    | Stream logs. `LogAlways` streams in real time, `LogOnFailure` dumps on test failure. |

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