# Office Foosball Stats

Single-binary Go API + SvelteKit SPA packaged into one Docker image with SQLite persistence.

## Run with Docker

Build the image:

```sh
docker build -t office-foosball-stats:latest .
```

Run the app with a mounted data volume:

```sh
docker run --rm -p 8080:8080 -v "$(pwd)/data:/data" office-foosball-stats:latest
```

The app stores SQLite data at `/data/foosball.db` by default.

## Environment variables

- `PORT` (default: `8080`)
- `SQLITE_PATH` (default: `/data/foosball.db` in Docker image)
