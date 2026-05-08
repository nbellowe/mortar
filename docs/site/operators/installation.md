# Installation

::: warning Early development
Mortar is under active development. Pre-built binaries and a Docker image are not yet published. Run from source for now.
:::

## From source

**Requirements:** Go 1.21+, Git

```bash
git clone https://github.com/nbellowe/mortar.git
cd mortar
go build -o mortar ./cmd/server
```

Then run it, pointing at your config file:

```bash
./mortar --config /path/to/config.yaml
```

The server listens on port `8080` by default. Access the web UI at `http://<host>:8080`.

## With Docker

A Dockerfile and published image are planned. In the meantime, build from source.

## Reverse proxy

For production use, put Mortar behind a reverse proxy (nginx, Caddy, Traefik) that handles TLS termination. Mortar itself does not serve HTTPS.

Example Caddy block:

```
mortar.home.example.com {
    reverse_proxy localhost:8080
}
```

## Upgrading

Pull the latest code and rebuild:

```bash
git pull
go build -o mortar ./cmd/server
```

Restart the server after rebuilding. Mortar uses SQLite for state; no migrations are required for `v0.x` builds.
