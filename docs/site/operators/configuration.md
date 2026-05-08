# Configuration

Mortar is configured with a YAML file. Secrets (API keys) should be passed as environment variables — never hardcode them.

## Minimal example

```yaml
server:
  port: 8080

auth:
  users:
    - username: alice
      # Generate with: htpasswd -bnBC 12 "" yourpassword | tr -d ':\n'
      password_hash: "$2b$12$..."

plugins:
  - id: jellyfin
    type: jellyfin
    url: http://jellyfin:8096
    api_key: ${JELLYFIN_API_KEY}

  - id: jellyseerr
    type: jellyseerr
    url: http://jellyseerr:5055
    api_key: ${JELLYSEERR_API_KEY}

routing:
  requests:
    video: jellyseerr
```

## Environment variable substitution

Any value can reference an environment variable with `${VAR_NAME}`. Mortar substitutes these at startup. If a referenced variable is unset, startup fails with an error identifying the missing variable.

## Reference

### `server`

| Field | Default | Description |
|---|---|---|
| `port` | `8080` | Port the HTTP server listens on |

### `auth.users`

A list of local users. Each entry requires `username` and `password_hash` (bcrypt).

### `plugins`

Each plugin entry requires:

| Field | Description |
|---|---|
| `id` | Unique identifier used in routing and logs (e.g. `jellyfin`) |
| `type` | Plugin type — must match a registered type (see [Plugins](./plugins)) |
| `url` | Base URL of the upstream service |
| `api_key` | Service API key (use an env var reference) |

### `routing.requests`

Maps media types to the plugin responsible for handling requests.

| Key | Value |
|---|---|
| `video` | Plugin `id` that handles movie and TV requests |
| `audio` | Plugin `id` that handles audiobook requests |
| `ebook` | Plugin `id` that handles ebook requests |

If a media type has no routing entry, Mortar will not offer a Request button for that type.
