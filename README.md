# YXORP
Yxorp is a simple and flexible reverse proxy / API Gateway server written in Go.

## Features

- Easy configuration with JSON
- SSL/TLS termination (configurable TLS versions)
- Path-based routing to multiple backends
- Optional path prefix rewriting
- Basic rate limiting (requests per second and burst)
- Customizable timeouts

## Getting Started

### Installation

Clone the repository:

```bash
git clone https://github.com/gberrante/yxorp.git
cd yxorp
go build -o yxorp main.go
```

### Usage

Start the proxy with your configuration file:

```bash
./yxorp --config ./config.json
```

Or set the `YXORP_CFG_FILE` environment variable:

```bash
export YXORP_CFG_FILE=./config.json
./yxorp
```

### Example Configuration (`config.json`)

```json
{
    "port": "8080",
    "logpath": "./yxorp.log",
    "tls_version": "1.2",
    "crt_path": "./cert.pem",
    "key_path": "./key.pem",
    "req_limit": 10,
    "req_burst": 20,
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

## Contributing

Contributions are welcome! Please open issues or submit pull requests.

## License

This project is licensed under the MIT License.