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

## Docker

Build the image:

```powershell
docker build -t url-shortener-app .
```

Run the container with a persistent database volume:

```powershell
docker run --rm -p 8080:8080 -e BASE_URL=http://localhost:8080 -v "${PWD}/data:/app/data" url-shortener-app
```

The container defaults to:

| Variable | Default |
| --- | --- |
| `PORT` | `8080` |
| `DB_PATH` | `/app/data/urls.db` |

Test it:

```powershell
curl http://localhost:8080/health
```
