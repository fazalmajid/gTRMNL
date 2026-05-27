# gTRMNL

A self-hosted [TRMNL](https://usetrmnl.com) API server. Instead of fetching plugin content from the TRMNL cloud, it renders a configured URL via headless Chrome, dithers the result to the 6-color [Spectra 6](https://www.waveshare.com/wiki/6inch_e-Paper_HAT_(F)) e-ink palette, and serves the image to your TRMNL device.

## How it works

1. The TRMNL device polls `GET /api/display`
2. The server renders the configured URL at 800×480 using headless Chrome (via [chromedp](https://github.com/chromedp/chromedp))
3. The image is Floyd-Steinberg dithered to the 6 Spectra primaries and cached in RAM
4. The device receives a JSON response with the image URL and fetches the PNG

The rendered image is cached for the duration of the refresh interval — Chrome is only launched when the cache expires.

## Requirements

- Go 1.22+
- Google Chrome or Chromium installed and on `$PATH`

## Installation

Build from source:

```sh
git clone ...
cd gTRMNL
go build -o gtrmnl .
```

## Usage

```sh
gtrmnl -url https://your-dashboard.example.com -base-url http://192.168.1.10:8080
```

Point your TRMNL device's custom server at `http://192.168.1.10:8080`.

## Options

| Flag | Default | Description |
|---|---|---|
| `-url` | *(required)* | URL to render on the device screen |
| `-base-url` | `http://localhost:8080` | Base URL the device uses to fetch the image (must be reachable from the device; include port if non-standard) |
| `-addr` | `:8080` | Address and port the server listens on |
| `-refresh` | `1800` | Cache TTL in seconds; also returned as `refresh_rate` to the device |
| `-token` | *(none)* | If set, requests to `/api/display` must include a matching `access-token` header |

## Docker

A multi-stage Dockerfile is provided. The final image is based on [Google's distroless](https://github.com/GoogleContainerTools/distroless) (`gcr.io/distroless/base-debian12`) — no shell, no package manager, minimal attack surface. Chromium and its shared libraries are copied in from a Debian build stage.

### docker compose

Create a `.env` file:

```sh
TRMNL_URL=https://your-dashboard.example.com
TRMNL_BASE_URL=http://192.168.1.10:8080
# optional
# TRMNL_TOKEN=yourtoken
# TRMNL_REFRESH=1800
```

Then:

```sh
docker compose up --build
```

### Plain Docker

```sh
docker build -t gtrmnl .
docker run --rm \
  --read-only --tmpfs /tmp --shm-size=128m \
  --cap-drop ALL --security-opt no-new-privileges \
  -p 8080:8080 \
  gtrmnl \
    -url https://your-dashboard.example.com \
    -base-url http://192.168.1.10:8080
```

## API endpoints

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/setup` | Device registration (called on first boot) |
| `GET` | `/api/display` | Returns JSON with the current screen image URL |
| `GET` | `/image/<name>` | Serves the cached PNG from RAM |
| `POST` | `/api/log` | Accepts device telemetry from firmware and logs it to stdout |

## Color palette

Images are dithered to the 6 Spectra primaries:

| Color | RGB |
|---|---|
| Black | `(0, 0, 0)` |
| White | `(192, 192, 192)` |
| Yellow | `(192, 192, 0)` |
| Red | `(192, 0, 0)` |
| Blue | `(0, 0, 192)` |
| Green | `(0, 192, 0)` |

## License

This project was generated with [Claude Code](https://claude.ai/code) and is dedicated to the public domain under [CC0 1.0 Universal](https://creativecommons.org/publicdomain/zero/1.0/). To the extent possible under law, all copyright and related rights are waived. You can copy, modify, and distribute this work without asking permission.
