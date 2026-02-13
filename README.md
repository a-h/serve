# serve

Serve directory over HTTP.

## Usage

### Serve current directory on port 8080

```bash
serve
```

### Options

```bash
serve --help
```

```
-addr string
    Address to serve on. (Env: SERVE_ADDR) (default ":8080")
-auth string
    Username:Password for basic auth, no auth if not set. (Env: SERVE_AUTH)
-crt string
    Path to crt file for TLS. (Env: SERVE_CRT)
-dir string
    Directory to serve. (Env: SERVE_DIR) (default ".")
-help
    Print help.
-key string
    Path to key file for TLS. (Env: SERVE_KEY)
-log-remote-addr
    Log remote address. (Env: SERVE_LOG_REMOTE_ADDR)
-read-header-timeout duration
    Amount of time allowed to read request headers. (Env: SERVE_READ_HEADER_TIMEOUT) (default 5s)
-read-only
    Allow only GET and HEAD requests. (Env: SERVE_READ_ONLY) (default true)
-read-timeout duration
    Maximum duration for reading the entire request, including the body. (Env: SERVE_READ_TIMEOUT) (default 24h0m0s)
-write-timeout duration
    Maximum duration before timing out writes of the response. (Env: SERVE_WRITE_TIMEOUT) (default 12h0m0s)
```

## Tasks

### build

```bash
go build -o serve .
```

### test

```bash
go test -v ./...
```

### image-build

Interactive: true

```bash
nix build .#image
```

### image-run

Interactive: true

```bash
nix build .#image
docker load < result
# Use SERVE_READ_ONLY=false to allow file uploads.
docker run -p 8080:8080 -v "$PWD:/data" -e SERVE_READ_ONLY=false  ghcr.io/a-h/serve:latest
```

### image-push

Interactive: true

```bash
nix build .#image
gunzip -c result > serve.tar
skopeo copy docker-archive:serve.tar docker://ghcr.io/a-h/serve:latest
```

### file-upload

```bash
echo "Hello, world!" > upload.txt
curl -X POST -F "file=@upload.txt" http://localhost:8080/upload.txt
```
