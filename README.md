# starcharts

[![Build Status](https://img.shields.io/github/actions/workflow/status/caarlos0/starcharts/build.yml?style=for-the-badge)](https://github.com/caarlos0/starcharts/actions?workflow=build)
[![Coverage Status](https://img.shields.io/codecov/c/gh/caarlos0/starcharts.svg?logo=codecov&style=for-the-badge)](https://codecov.io/gh/caarlos0/starcharts)
[![](http://img.shields.io/badge/godoc-reference-5272B4.svg?style=for-the-badge)](http://godoc.org/github.com/caarlos0/starcharts)

Plot your repo stars over time!

## How it works

Star data comes from GitHub's [star history API][api], which returns the number
of stars a repository got on each day, grouped by week, without exposing who
starred it.

Star charts plots one data point per day in which the repository got at least
one star, so the resulting chart is accurate regardless of how many stars the
repository has.

Responses are cached in Redis and revalidated with `ETag`s, so repeated renders
of the same chart usually don't spend any API rate limit.

[api]: https://docs.github.com/rest/activity/starring#get-repository-star-history

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
| `GITHUB_MAX_RATE_LIMIT_USAGE` | `80` | API Rate Limit usage threshold percentage |
| `LISTEN` | `127.0.0.1:3000` | Server listen address |

## Example

[![starcharts stargazers over time](https://starchart.cc/caarlos0/starcharts.svg)](https://starchart.cc/caarlos0/starcharts)
