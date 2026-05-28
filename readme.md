# crocsy

crocsy is a simple, lightweight HTTPS reverse proxy for forwarding requests to backend services.

It supports:

- reverse proxying
- virtual hosts
- path-based routing
- HTTPS termination
- HTTP → HTTPS redirects
- static file serving
- request/response header rewriting
- variable replacement in headers and redirects

---

## Features

- Reverse proxy for HTTP/HTTPS backends
- Multiple listening ports
- Virtual hosts
- Path-based routing
- TLS termination
- HTTP → HTTPS redirects
- Static file serving
- Custom request headers
- Custom response headers
- Variable replacement support
- Minimal YAML configuration

---

## Installation

Clone the repository and build crocsy:

```bash
git clone https://github.com/tvrzna/crocsy.git
cd crocsy
make build
```

Binary will be available in:

```bash
./dist/crocsy
```

---

## Usage

```text
Usage: crocsy [options]

Options:
        -h, --help              print this help
        -v, --version           print version
        -c, --config            set path to config file
        -C, --print-config      prints currently loaded configuration
```

Example:

```bash
crocsy -c crocsy.yaml
```

---

## Configuration

crocsy reads configuration from a YAML file.

Example:

```yaml
server:
  - listen: ":80"
    redirect: "https://$host$request_uri"

  - listen: ":443"

    tls:
      cert_file: "/etc/ssl/crocsy.crt"
      key_file: "/etc/ssl/crocsy.key"

    set-headers:
      Referrer-Policy: "origin"
      X-Content-Type-Options: "nosniff"
      Strict-Transport-Security: "max-age=31536000; includeSubDomains"

    route:
      - path: "/"
        target: "http://localhost:3000"

      - path: "/assets/"
        root: "/srv/www/assets"
        autoindex: true
        index: "index.html index.htm"

      - host: "api.example.com"
        path: "/"
        target: "http://localhost:8080"

      - host: "wiki.example.com"
        path: "/"
        target: "http://localhost:8081"

      - path: "/api/"
        target: "http://localhost:9000"

        proxy-set-headers:
          X-Forwarded-Host: "$host"
          X-Forwarded-Proto: "$scheme"
          X-Client-IP: "$remote_addr"
          X-Original-URI: "$request_uri"

        set-headers:
          X-Proxy: "crocsy"
```

---

## Configuration Reference

### Server

| Key | Description |
|---|---|
| `listen` | Address/port to listen on |
| `redirect` | Redirect requests to another URL |
| `tls.cert_file` | TLS certificate file |
| `tls.key_file` | TLS private key file |
| `set-headers` | Response headers applied globally |

---

### Route

A route can proxy requests or serve static files.

| Key | Description |
|---|---|
| `host` | Optional virtual host |
| `path` | URL path prefix |
| `target` | Backend target URL |
| `root` | Static files root directory |
| `autoindex` | Enable directory listing |
| `index` | Index files |
| `set-headers` | Response headers |
| `proxy-set-headers` | Headers added to proxied requests |

---

## Variables

Variables can be used in redirects, response headers, and proxy request headers.

### Built-in variables

| Variable | Description |
|---|---|
| `$host` | Request host |
| `$port` | Request port |
| `$scheme` | Request scheme (`http` or `https`) |
| `$request_uri` | Full request URI |
| `$path` | Request path |
| `$method` | HTTP request method |
| `$query` | Raw query string |
| `$remote_addr` | Remote client address |

---

### Request header variables

Request headers can be accessed using:

```text
$header.<header-name>
```

Example:

```yaml
proxy-set-headers:
  X-Custom-Token: "$header.x-custom-token"
```

---

## Examples

### HTTP → HTTPS redirect

```yaml
server:
  - listen: ":80"
    redirect: "https://$host$request_uri"
```

---

### Reverse proxy

```yaml
route:
  - path: "/api/"
    target: "http://localhost:8080"
```

---

### Virtual host

```yaml
route:
  - host: "api.example.com"
    path: "/"
    target: "http://localhost:8080"
```

---

### Static file serving

```yaml
route:
  - path: "/assets/"
    root: "/srv/www/assets"
```

---

### Custom proxy request headers

```yaml
proxy-set-headers:
  X-Forwarded-Host: "$host"
  X-Forwarded-Proto: "$scheme"
```

---

### Custom response headers

```yaml
set-headers:
  X-Request-Method: "$method"
  X-Original-Path: "$path"
```

---

## License

MIT License