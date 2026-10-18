# whereami

whereami is a web application that takes a location (a latitude and longitude, from a map click, a geocoded address, or the browser's location) and reports what is at or near that point.  It is a way to learn about your place in these United States.  

## Design

The application is a single Go binary built on the Echo web framework.

- **Handlers** (`handlers/`) serve the pages and the `/details` and `/geocode` endpoints.
- **Queries** (`queries/`) hold one query per result type. Each query runs against a local SQLite database with SpatiaLite, or calls an external API. Queries run concurrently for each request.
- **Summit index:** At startup, the application loads all summits with an elevation into an in-memory R-tree (`rtreego`).  This works around a limitation of the SpatialLite nearest neighbors algorithm, where filters (in this case elevation greater than current location) cannot be considered when choosing a nearest neighbor.
- **Templates** (`templates/`) are written with [templ](https://templ.guide) and compiled to Go code. The pages use htmx for partial updates.
- **Metrics:** Request and query timings are exported to Prometheus. The metrics endpoint runs on a separate port so it can be kept off the public interface.

The database is opened read-only. All data is loaded ahead of time by a separate process and is not modified by the application.

## Data Sources

The database is built from the following datasets.

| Dataset                         | Agency                                  |
|---------------------------------|-----------------------------------------|
| County and County Equivalents   | US Census                               |
| US Congressional Districts      | US Census                               |
| Incorporated Places             | US Census                               |
| Unincorporated Places           | US Census                               |
| Ecoregions                      | US Environmental Protection Agency      |
| Alaska Ecoregions               | US Environmental Protection Agency      |
| National Parks                  | US Geological Survey (The National Map) |
| Level 12 Hydrologic Units (WBD) | US Geological Survey (The National Map) |
| Streams (WBD)                   | US Geological Survey (The National Map) |
| GNIS                            | Geographic Names Information System     |
| Tide Stations                   | NOAA Tides and Stations                 

The following external services are called at request time.

| Service                                                                                         | Used for | Source                                                  |
|-------------------------------------------------------------------------------------------------|---|---------------------------------------------------------|
| Elevation Point Query Service (The National Map) | Elevation | https://apps.nationalmap.gov/epqs/                                                        |
| National Weather Service API   | Weather | https://www.weather.gov/documentation/services-web-api |
| NOAA Tides and Currents                                                                         | Tide predictions | https://tidesandcurrents.noaa.gov                       |
| AirNow                                                                                          | Air quality index | https://www.airnow.gov                                  |

**Database:** The SQLite database file is not included in this repository.

## SpatiaLite Integration

whereami uses [SpatiaLite](https://www.gaia-gis.it/fossil/libspatialite/index), a spatial extension for SQLite, for all geometry work. The `ST_*` functions and the `SpatialIndex` R-tree tables it provides are used in nearly every query.

The extension is loaded when each database connection opens (`queries/db.go`):

- The application registers a `spatialite` SQL driver that wraps `mattn/go-sqlite3` and loads `mod_spatialite` on connect.
- The library is looked up in the directory named by `SPATIALITE_PATH`. If that variable is not set, the default is `/opt/homebrew/Cellar/libspatialite/5.1.0_4/lib`, which is the Homebrew location on macOS. This default is tied to a specific Homebrew version and will need to change when that version is upgraded.
- Startup fails if the extension cannot be loaded.

Install SpatiaLite before running the application:

- **macOS (Homebrew):** `brew install libspatialite`
- **Debian or Ubuntu:** `sudo apt install libsqlite3-mod-spatialite`

If the library is installed somewhere other than the default path, set `SPATIALITE_PATH` to the directory that contains `mod_spatialite`.

## Configuration

whereami is configured with environment variables.

| Variable | Default | Description                                                                                    |
|---|---|------------------------------------------------------------------------------------------------|
| `DB_PATH` | none | Path to the SpatiaLite SQLite database file. The application cannot serve results without it.  |
| `SPATIALITE_PATH` | Homebrew path (see above) | Directory containing `mod_spatialite`.                                                         |
| `MAPBOX_TOKEN` | none | Mapbox access token used for reverse geocode queries.                                          |
| `AQI_API_KEY` | none | AirNow API key uses for determining AQI. Without it, air quality results will not be returned. |
| `API_TIMEOUT` | `10` | Timeout in seconds for calls to external services.                                             |

## Deployment

whereami is distributed as a Docker image. It can run on any host that supports Docker. The production instance runs on a VPS.

The image is built from the `Dockerfile` in this repository. It contains the compiled binary and the static assets, and it installs `libsqlite3-mod-spatialite` at runtime.

When a release is published on GitHub, the `release` workflow builds the image and pushes it to `ghcr.io/zmaillard/whereami`.

To build the image locally:

```sh
docker build --build-arg VERSION_APP=0.0.0 -t whereami .
```

To run it, mount the database into the container and set the environment variables:

```sh
docker run -d \
  -p 8080:8080 \
  -p 8081:8081 \
  -v /path/to/whereami.db:/data/whereami.db:ro \
  -e DB_PATH=/data/whereami.db \
  -e MAPBOX_TOKEN=... \
  -e AQI_API_KEY=... \
  whereami
```

- Port `8080` serves the web application.
- Port `8081` serves Prometheus metrics at `/metrics`. Do not expose it publicly.

## Developer Instructions

### Requirements

- Go 1.26 (see `go.mod`)
- CGO enabled, since the SQLite driver requires it
- SpatiaLite (see [SpatiaLite Integration](#spatialite-integration))
- [templ](https://templ.guide) CLI, to regenerate the template code

### Setup

```sh
git clone https://github.com/zmaillard/whereami.git
cd whereami
go mod download
```

### Running locally

Set the environment variables from the [Configuration](#configuration) section, then run:

```sh
export DB_PATH=/path/to/whereami.db
export MAPBOX_TOKEN=...
export AQI_API_KEY=...
go run main.go
```

Open http://localhost:8080.

### Templates

Templates are written in `templates/*.templ`. After editing one, regenerate the Go code:

```sh
cd templates && go generate
```

The generated `*_templ.go` files are committed and must be regenerated and committed along with the `.templ` source.

### Building the database index helper

`cmd/index.go` is a maintenance tool that precomputes data in the database. Its current task is building the closest-city table. It reads `DB_PATH` the same way the application does. Review the code before running it, because parts of it are commented out and it writes to the database.

### Tests

The repository does not currently include tests. CI runs `go build` only.

### Project layout

| Path | Contents |
|---|---|
| `main.go` | Entry point: loads config, opens the database, builds the summit index, and starts the servers. |
| `config/` | Environment variable configuration. |
| `handlers/` | HTTP handlers. |
| `queries/` | Database and external API queries, one file per result type. |
| `models/` | Request and result models. |
| `templates/` | templ templates and generated code. |
| `assets/` | CSS, JavaScript, and images served at `/assets`. |
| `middleware/` | Request logging. |
| `metrics/` | Custom Prometheus metrics. |
| `util/` | Helpers for FIPS codes, time, rendering, and vector math. |
| `cmd/` | Maintenance tools. |
