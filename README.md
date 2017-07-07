# starchart

Plot your repo stars over time!

## Usage

```bash
go get github.com/marianogappa/chart
go get github.com/caarlos0/starchart
export GITHUB_TOKEN=my-token

starchart goreleaser/goreleaser | chart line --date-format 2006-01-02T15:04:05Z
```

And that's it!

Example:

![gorelease stars over time](https://user-images.githubusercontent.com/245435/27939013-5df2718c-6298-11e7-8f92-7e03bb994d91.png)

## Thanks

@marianogappa and his awesome [chart](github.com/marianogappa/chart) tool!
