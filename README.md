# Proxy Checker

`ProxyChecker` is a comprehensive Go library designed for the retrieval, validation, and management of proxies, including HTTP, SOCKS4, and SOCKS5 types. It streamlines the process of obtaining proxies that are tested for connectivity and performance, and provides an in-memory cache for fast access to validated proxies.

## Installation

Install the `proxy-checker` package with the following command:

```bash
go get github.com/44za12/proxy-checker
```

This will retrieve the package and include it in your project.

## Usage

To use `proxy-checker`, import it into your Go application:

```go
import "github.com/44za12/proxy-checker"
```

### Creating a New ProxyChecker

Instantiate a `ProxyChecker` object:

```go
checker := proxychecker.NewProxyChecker()
```

### Fetching a Valid Proxy

To obtain a valid proxy, call `GetGoodProxy`:

```go
proxy := checker.GetGoodProxy()
if proxy.Address == "" {
    fmt.Println("No valid proxy found.")
} else {
    fmt.Printf("Fetched a valid proxy: %s\n", proxy.Address)
}
```

### Revalidating Proxies

Automatically revalidate the list of good proxies at specified intervals:

```go
interval := 1 * time.Hour // Revalidation interval
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

checker.RecheckGoodProxies(ctx, interval)
```

## How It Works

The `ProxyChecker` performs several key operations:

- **Scraping**: Proxies are scraped from predefined sources and are then processed.
- **Validation**: Each proxy is validated for connectivity by making requests to test URLs.
- **Caching**: Valid proxies are stored in an in-memory cache for fast retrieval.
- **Revalidation**: Optionally, proxies can be periodically revalidated to ensure they remain functional.

This caching strategy prevents unnecessary repeated validation, saving resources and providing quick access to a list of working proxies.

## Features

- **Support for Multiple Proxy Types**: HTTP, SOCKS4, and SOCKS5.
- **Concurrent Validation**: Proxies are validated concurrently for efficiency.
- **Automatic Revalidation**: Maintain a fresh pool of proxies with automatic revalidation.
- **In-Memory Caching**: Quick access to validated proxies without additional network calls.
- **Custom Headers**: Set custom headers for HTTP requests during proxy validation.

## Advanced Usage

### Custom HTTP Headers

Customize HTTP headers for validation requests:

```go
checker.Headers["Custom-Header"] = "YourValue"
```

### Direct Proxy Validation

Validate a specific proxy directly:

```go
ctx := context.Background()
proxyType, isValid := checker.CheckProxy(ctx, proxychecker.Proxy{Address: "ip:port"}, []string{"HTTP", "SOCKS4", "SOCKS5"})
if isValid {
    fmt.Printf("Proxy is valid and of type: %s\n", proxyType)
}
```

### Retrieving All Proxies

Fetch all proxies from the cache:

```go
proxies := checker.GetAllProxies()
fmt.Printf("There are %d proxies in the cache.\n", len(proxies))
```

## Contributing

We welcome contributions to the `proxy-checker` library. Please submit any issues or pull requests through the project's GitHub repository.