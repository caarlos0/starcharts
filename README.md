# starcharts

[![Build Status](https://img.shields.io/github/actions/workflow/status/caarlos0/starcharts/build.yml?style=for-the-badge)](https://github.com/caarlos0/starcharts/actions?workflow=build)
[![Coverage Status](https://img.shields.io/codecov/c/gh/caarlos0/starcharts.svg?logo=codecov&style=for-the-badge)](https://codecov.io/gh/caarlos0/starcharts)
[![](http://img.shields.io/badge/godoc-reference-5272B4.svg?style=for-the-badge)](http://godoc.org/github.com/caarlos0/starcharts)

Plot your repo stars over time!

## Features

### 智能采样模式（大型仓库优化）

对于拥有大量 Star 的大型仓库，本项目采用**智能采样模式**来高效获取 Star 历史数据并渲染趋势图。

**工作原理：**

1. **自动检测**：首先请求 GitHub API 的第一页数据，通过解析 `Link` Header 获取总页数
2. **模式切换**：
   - 当总页数 ≤ `maxSamplePages`（默认 15 页，即约 1500 stars）时，获取所有数据
   - 当总页数 > `maxSamplePages` 时，自动切换到采样模式
3. **均匀采样**：在所有页面中均匀选取采样点，确保覆盖仓库 Star 增长的完整时间线
4. **数据点提取**：从每个采样页提取第一个 Stargazer 的时间戳和对应的 Star 计数
5. **趋势补全**：添加当前时间和总 Star 数作为最后一个数据点，确保图表延伸到最新状态

## Usage

```console
go run main.go
```

Then browse http://localhost:3000/me/myrepo .

## Configuration

通过环境变量配置：

| 环境变量 | 默认值 | 说明 |
|---------|--------|------|
| `REDIS_URL` | `redis://localhost:6379` | Redis 缓存地址 |
| `GITHUB_TOKENS` | - | GitHub API Token（支持多个，逗号分隔） |
| `GITHUB_PAGE_SIZE` | `100` | 每页获取的 Star 数量 |
| `GITHUB_MAX_SAMPLE_PAGES` | `15` | 最大采样页数（超过此数触发采样模式） |
| `GITHUB_MAX_RATE_LIMIT_USAGE` | `80` | API Rate Limit 使用上限百分比 |
| `LISTEN` | `127.0.0.1:3000` | 服务监听地址 |

## Example

示例图表：

[![starcharts stargazers over time](https://starchart.cc/caarlos0/starcharts.svg)](https://starchart.cc/caarlos0/starcharts)
