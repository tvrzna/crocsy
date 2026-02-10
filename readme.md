# crocsy

**crocsy** is a simple, lightweight HTTPS reverse proxy for forwarding requests to one or more backend services. It supports multiple listening ports, path-based routing, and TLS termination.

## Features

- Forward HTTP/HTTPS requests to one or more backend services
- Multiple listening ports
- Path-based routing
- TLS termination for secure connections
- Minimal and easy-to-configure

## Usage
```
Usage: crocsy [options]
Options:
        -h, --help              print this help
        -v, --version           print version
        -c, --config            set path to config file
```

## Installation
```bash
git clone https://github.com/tvrzna/crocsy.git
cd crocsy
make build
```

## Configuration
crocsy reads its configuration from a YAML file. Example:

```yaml
server:
  - listen: ":80"
    redirect: "https://$host$request_uri"

  - listen: ":443"
    tls:
      cert_file: "/etc/ssl/crocsy.crt"
      key_file: "/etc/ssl/crocsy.key"
    route:
      - path: "/api/"
        target: "http://localhost:8080/"
        set-headers:
          Content-Security-Policy: "default-src 'self'; script-src 'none'; style-src 'none'; font-src 'self'; connect-src 'self' https: data:; img-src 'self';"

      - path: "/api2/"
        target: "http://localhost:8081/"
```

