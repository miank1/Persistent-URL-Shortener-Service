# Persistent URL Shortener Service

## Configuration

The service reads configuration from environment variables at startup. If a variable is not set, the default value is used.

| Variable | Description | Default |
| --- | --- | --- |
| `PORT` | HTTP server port. | `8080` |
| `DB_PATH` | SQLite database file path. | `urls.db` |
| `BASE_URL` | Base URL used when constructing shortened URLs. | `http://localhost:<PORT>` |

Example:

```powershell
$env:PORT = "8080"
$env:DB_PATH = "urls.db"
$env:BASE_URL = "http://localhost:8080"
go run .
```
