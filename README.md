# yxorp

**yxorp** is a lightweight, single-binary reverse proxy and API gateway written in Go. Configure it with a single JSON file and get path-based routing, TLS termination, and rate limiting with zero dependencies beyond the standard library.

## Features

- **Path-based routing** — forward requests to different backends based on URL prefix
- **Path prefix rewriting** — optionally strip the matched prefix before forwarding
- **Rate limiting** — global token-bucket limiter (requests per second + burst)
- **TLS termination** — configurable minimum TLS version (1.0 – 1.3); omit for plain HTTP
- **Flexible logging** — writes to both stdout and a log file simultaneously
- **Loopback protection** — startup check prevents accidental self-referencing targets
- **Single binary** — no runtime dependencies; just build and run

## Getting Started

### Prerequisites

- Go 1.18 or later

### Build

```bash
git clone https://github.com/gberrante/yxorp.git
cd yxorp
go build -o yxorp .
```

### Run

Pass the config file path via a flag:

```bash
./yxorp --config ./config.json
```

Or via an environment variable (takes precedence over the flag):

```bash
export YXORP_CFG_FILE=./config.json
./yxorp
```

## Configuration

All settings live in a single JSON file.

### Server options

| Field | Type | Default | Description |
|---|---|---|---|
| `port` | string | — | Port to listen on (e.g. `"8080"`) |
| `logpath` | string | — | Path to the log file |
| `tls_version` | string | `""` | Minimum TLS version: `"1.0"`, `"1.1"`, `"1.2"`, `"1.3"`, or omit / set to anything else to disable TLS |
| `crt_path` | string | `""` | Path to the TLS certificate (required when TLS is enabled) |
| `key_path` | string | `""` | Path to the TLS private key (required when TLS is enabled) |
| `req_limit` | int | — | Sustained request rate limit (requests per second) |
| `req_burst` | int | — | Maximum burst size above the sustained limit |
| `read_timeout` | int | — | HTTP read timeout in seconds |
| `write_timeout` | int | — | HTTP write timeout in seconds |
| `idle_timeout` | int | — | HTTP keep-alive idle timeout in seconds |
| `targets` | array | — | List of routing rules (see below) |

### Target options

| Field | Type | Description |
|---|---|---|
| `name` | string | Human-readable label for the target |
| `prefix` | string | URL prefix to match (e.g. `"/api/"`) |
| `target` | string | Backend URL to proxy to (e.g. `"http://localhost:3000"`) |
| `rewrite` | bool | When `true`, strips the matched prefix before forwarding the request |

### Example (`config.json`)

```json
{
    "port": "8080",
    "logpath": "./yxorp.log",
    "tls_version": "1.2",
    "crt_path": "./cert.pem",
    "key_path": "./key.pem",
    "req_limit": 100,
    "req_burst": 200,
    "read_timeout": 10,
    "write_timeout": 10,
    "idle_timeout": 60,
    "targets": [
        {
            "name": "api",
            "prefix": "/api/",
            "target": "http://localhost:3000",
            "rewrite": true
        },
        {
            "name": "web",
            "prefix": "/",
            "target": "http://localhost:8081",
            "rewrite": false
        }
    ]
}
```

In this example:
- `GET /api/users` → forwarded to `http://localhost:3000/users` (prefix stripped)
- `GET /app/index.html` → forwarded to `http://localhost:8081/app/index.html` (prefix kept)

## How It Works

1. yxorp reads the JSON config and validates all targets at startup.
2. A token-bucket rate limiter is applied globally to every incoming request.
3. Incoming requests are matched against the registered prefixes in the order they are defined; the first match wins.
4. The matching request is reverse-proxied to the configured backend, optionally with the prefix stripped.
5. All log output is written to both stdout and the configured log file.

## Contributing

Contributions are welcome! Please open an issue to discuss your idea or submit a pull request.

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE) for details.