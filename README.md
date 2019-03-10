# sslcheck

[![Go Report Card](https://goreportcard.com/badge/go.soon.build/sslcheck)](https://goreportcard.com/report/go.soon.build/sslcheck)

Tool to monitor SSL certificates.

```
sslcheck --console --host thisissoon.com
```

## Development

 - Go 1.11+
 - Dependencies managed with `go mod`

### Setup

These steps will describe how to setup this project for active development. Adjust paths to your desire.

1. Clone the repository: `git clone github.com/thisissoon/sslcheck sslcheck`
2. Build: `make build`
3. 🍻

### Dependencies

Dependencies are managed using `go mod` (introduced in 1.11), their versions
are tracked in `go.mod`.

To add a dependency:
```
go get url/to/origin
```

### Configuration

Configuration can be provided through a toml file, these are loaded
in order from:

- `/etc/sslcheck/sslcheck.toml`
- `$HOME/.config/sslcheck.toml`

Alternatively a config file path can be provided through the
-c/--config CLI flag.

#### Example sslcheck.toml
```toml
[log]
console = true
level = "debug"  # [debug|info|error]
```
