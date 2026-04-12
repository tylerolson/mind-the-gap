# mind-the-gap

![Go Version](https://img.shields.io/badge/go-1.26-00ADD8?logo=go)

Realtime transit tracker for BART data.

## What This Repository Is

`mind-the-gap` is a monorepo of a realtime transit platform:

- ingest GTFS-RT feeds in Go
- distribute updates through Redis
- serve clients through FastAPI (REST + WebSocket)
- render live arrivals and map state in React

## Repo Layout

```text
mind-the-gap/
├── gtfs-ingestor/              # active now
│   ├── main.go
│   └── gtfs/
│       ├── client.go
│       ├── decode.go
│       └── model.go
├── services/                   # planned backend services
│   └── api-fastapi/            # planned API + WebSocket gateway
├── frontend/
│   └── react-app/              # planned live UI
└── infra/
    ├── docker-compose.yml      # planned local environment
    ├── redis/
    └── postgres/
```

## Current Status

Implemented today in `gtfs-ingestor/`:

- fetches BART Trip Updates feed
- decodes GTFS-RT protobuf payloads
- normalizes arrival updates
- prints structured terminal output

Current feed URL:

- `https://api.bart.gov/gtfsrt/tripupdate.aspx`

Not implemented yet:

- Redis channels and cache
- FastAPI REST endpoints
- FastAPI WebSocket live stream
- React frontend experience
- Postgres integration
