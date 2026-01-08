# starcharts

[![Build Status](https://img.shields.io/github/actions/workflow/status/caarlos0/starcharts/build.yml?style=for-the-badge)](https://github.com/caarlos0/starcharts/actions?workflow=build)
[![Coverage Status](https://img.shields.io/codecov/c/gh/caarlos0/starcharts.svg?logo=codecov&style=for-the-badge)](https://codecov.io/gh/caarlos0/starcharts)
[![](http://img.shields.io/badge/godoc-reference-5272B4.svg?style=for-the-badge)](http://godoc.org/github.com/caarlos0/starcharts)

Plot your repo stars over time!

## Features

### Smart Sampling Mode (Large Repository Optimization)

For large repositories with massive amounts of stars, this project uses **Smart Sampling Mode** to efficiently fetch star history data and render trend charts.

**How it works:**

1. **Auto Detection**: First requests the first page of GitHub API data and parses the `Link` Header to get total page count
2. **Mode Switching**:
   - When total pages ≤ `maxSamplePages` (default 15 pages, ~1500 stars), fetches all data
   - When total pages > `maxSamplePages`, automatically switches to sampling mode
3. **Uniform Sampling**: Evenly selects sample points across all pages to ensure coverage of the complete star growth timeline
4. **Data Point Extraction**: Extracts the timestamp and corresponding star count from the first Stargazer of each sampled page
5. **Trend Completion**: Adds current time and total star count as the final data point to ensure the chart extends to the latest state

## Usage

```console
go run main.go
```

Then browse http://localhost:3000/me/myrepo .

## Configuration

Configure via environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `REDIS_URL` | `redis://localhost:6379` | Redis cache URL |
| `GITHUB_TOKENS` | - | GitHub API Token (supports multiple, comma-separated) |
| `GITHUB_PAGE_SIZE` | `100` | Number of stars per page |
| `GITHUB_MAX_SAMPLE_PAGES` | `15` | Max sample pages (triggers sampling mode when exceeded) |
| `GITHUB_MAX_RATE_LIMIT_USAGE` | `80` | API Rate Limit usage threshold percentage |
| `LISTEN` | `127.0.0.1:3000` | Server listen address |

## Example

[![starcharts stargazers over time](https://starchart.cc/caarlos0/starcharts.svg)](https://starchart.cc/caarlos0/starcharts)
