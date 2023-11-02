---

# Proxy Checker

A Go package for fetching and validating high-quality SOCKS5 proxies.

## Installation

To install the `proxy-checker` package, run:

```bash
go get github.com/44za12/proxy-checker
```

## Usage

First, import the package in your Go code:

```go
import "github.com/44za12/proxy-checker"
```

### Fetch a Valid Proxy

To fetch a valid SOCKS5 proxy:

```go
proxy := proxychecker.GetGoodProxy()
if proxy == "" {
    fmt.Println("Failed to fetch a valid proxy.")
    return
}
fmt.Println("Fetched Proxy:", proxy)
```

## How It Works

Upon the first call to `GetGoodProxy`:

- The tool will scrape proxies from predefined sources.
- Validate each proxy for performance and reliability.
- Store valid proxies in an in-memory cache for quick retrieval in subsequent calls.

The in-memory cache ensures that the scraping and validation process is not executed repeatedly, offering a balance between speed and freshness of the proxies.

---
