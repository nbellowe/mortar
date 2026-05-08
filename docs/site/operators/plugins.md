# Plugins

Plugins connect Mortar to your upstream services. Each plugin declares the capabilities it provides; Mortar only calls methods the plugin has declared.

## Supported plugins

| Type | Support | Capabilities | Notes |
|---|---|---|---|
| `jellyfin` | Experimental | `library.browse`, `library.exists`, `library.resume` | Browse library, check availability, continue watching |
| `jellyseerr` | Experimental | `requests.video`, `activity.read` | Search and request movies/TV |
| `sonarr` | Experimental | `activity.read` | TV download and import activity |
| `radarr` | Experimental | `activity.read` | Movie download and import activity |
| `sabnzbd` | Experimental | `downloads.read` | NZB download queue |

**Experimental** means: basic docs, minimal tests, not release-blocking if broken.

## Capability reference

| Capability | What it enables |
|---|---|
| `library.browse` | Recently added rows on the home screen |
| `library.exists` | "Available" badge on search results |
| `library.resume` | Continue watching row on the home screen |
| `requests.video` | Request button for movies and TV shows |
| `requests.audio` | Request button for audiobooks |
| `requests.ebook` | Request button for ebooks |
| `downloads.read` | Download queue view |
| `activity.read` | Activity feed entries |

## Per-plugin configuration

### Jellyfin

```yaml
- id: jellyfin
  type: jellyfin
  url: http://jellyfin:8096
  api_key: ${JELLYFIN_API_KEY}
```

Get an API key from Jellyfin: **Dashboard → API Keys → +**.

To enable **Continue watching**, link each Mortar user to their Jellyfin user ID. This is currently done via config:

```yaml
auth:
  users:
    - username: alice
      password_hash: "..."
      external_accounts:
        - plugin_id: jellyfin
          external_user_id: "abc123"  # Jellyfin user ID
```

### Jellyseerr

```yaml
- id: jellyseerr
  type: jellyseerr
  url: http://jellyseerr:5055
  api_key: ${JELLYSEERR_API_KEY}
```

Get an API key from Jellyseerr: **Settings → General → API Key**.

### Sonarr

```yaml
- id: sonarr
  type: sonarr
  url: http://sonarr:8989
  api_key: ${SONARR_API_KEY}
```

### Radarr

```yaml
- id: radarr
  type: radarr
  url: http://radarr:7878
  api_key: ${RADARR_API_KEY}
```

### SABnzbd

```yaml
- id: sabnzbd
  type: sabnzbd
  url: http://sabnzbd:8080
  api_key: ${SABNZBD_API_KEY}
```

Get an API key from SABnzbd: **Config → General → API Key**.
