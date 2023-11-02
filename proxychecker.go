package proxychecker

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"
)

type Proxy struct {
	Address string
	Type    string
}

type safeMap struct {
	m   map[Proxy]bool
	mux sync.Mutex
}

func (sm *safeMap) set(key Proxy, value bool) {
    sm.mux.Lock()
    sm.m[key] = value
    sm.mux.Unlock()
}

func (sm *safeMap) getAllKeys() []Proxy {
    sm.mux.Lock()
    defer sm.mux.Unlock()
    keys := make([]Proxy, 0, len(sm.m))
    for k := range sm.m {
        keys = append(keys, k)
    }
    return keys
}

var cache = make([]Proxy, 0)
var cacheMutex sync.Mutex

var httpServers = []string{
    "https://www.google.com",
    "https://www.cloudflare.com/cdn-cgi/trace",
}

func inferProxyTypeFromURL(proxyURL string) string {
	proxyURL = strings.ToLower(proxyURL)
	if strings.Contains(proxyURL, "socks5") || strings.Contains(proxyURL, "socks-5") {
		return "SOCKS5"
	} else if strings.Contains(proxyURL, "socks4") || strings.Contains(proxyURL, "socks-4") {
		return "SOCKS4"
	} else if strings.Contains(proxyURL, "https") {
		return "HTTPS"
	}
	return "HTTP"
}

func RecheckGoodProxies(interval time.Duration) {
    ticker := time.NewTicker(interval)
    for {
        <-ticker.C
        cacheMutex.Lock()
        validProxies := []Proxy{} // Changed to []Proxy
        for _, proxy := range cache {
            if CheckProxy(proxy) {
                validProxies = append(validProxies, proxy)
            }
        }
        if len(validProxies) == 0 {
            cacheMutex.Unlock()
            FetchAndStoreGoodProxies()
            continue
        }
        cache = validProxies
        cacheMutex.Unlock()
    }
}

func CheckProxy(p Proxy) bool {
    var proxyURL *url.URL
	var err error

	switch p.Type {
	case "HTTP":
		proxyURL, err = url.Parse("http://" + p.Address)
	case "HTTPS":
		proxyURL, err = url.Parse("http://" + p.Address)
	case "SOCKS4":
		proxyURL, err = url.Parse("socks4://" + p.Address)
	case "SOCKS5":
		proxyURL, err = url.Parse("socks5://" + p.Address)
	default:
		return false
	}

	if err != nil {
		return false
	}

	tr := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
		DialContext: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).DialContext,
	}

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    resultChan := make(chan bool, len(httpServers))

    for _, server := range httpServers {
        go func(server string) {
            req, err := http.NewRequestWithContext(ctx, "GET", server, nil)
            if err != nil {
                resultChan <- false
                return
            }
            resp, err := tr.RoundTrip(req)
            if err != nil {
                resultChan <- false
                return
            }
            defer resp.Body.Close()
            if resp.StatusCode == http.StatusOK {
                resultChan <- true
            } else {
                resultChan <- false
            }
        }(server)
    }

    for i := 0; i < len(httpServers); i++ {
        if <-resultChan {
            return true
        }
    }
    return false
}

func scrapeProxies(url string) []string {
	pattern := regexp.MustCompile(`\b(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\b:\d{2,5}`)
	resp, err := http.Get(url)
	if err != nil {
		return []string{}
	}
	defer resp.Body.Close()
	html, err := io.ReadAll(resp.Body)
	if err != nil {
		return []string{}
	}
	return pattern.FindAllString(string(html), -1)
}

func FetchAndStoreGoodProxies() {
	urls := []string{
		"https://raw.githubusercontent.com/TheSpeedX/PROXY-List/master/http.txt",
		"https://raw.githubusercontent.com/TheSpeedX/PROXY-List/master/socks4.txt",
		"https://raw.githubusercontent.com/TheSpeedX/PROXY-List/master/socks5.txt",
		"https://raw.githubusercontent.com/ShiftyTR/Proxy-List/master/http.txt",
		"https://raw.githubusercontent.com/ShiftyTR/Proxy-List/master/https.txt",
		"https://raw.githubusercontent.com/ShiftyTR/Proxy-List/master/proxy.txt",
		"https://raw.githubusercontent.com/ShiftyTR/Proxy-List/master/socks4.txt",
		"https://raw.githubusercontent.com/ShiftyTR/Proxy-List/master/socks5.txt",
		"https://raw.githubusercontent.com/roosterkid/openproxylist/main/HTTPS_RAW.txt",
		"https://raw.githubusercontent.com/roosterkid/openproxylist/main/SOCKS4_RAW.txt",
		"https://raw.githubusercontent.com/roosterkid/openproxylist/main/SOCKS5_RAW.txt",
		"https://raw.githubusercontent.com/hookzof/socks5_list/master/proxy.txt",
		"https://raw.githubusercontent.com/casals-ar/proxy-list/main/http",
		"https://raw.githubusercontent.com/casals-ar/proxy-list/main/https",
		"https://raw.githubusercontent.com/casals-ar/proxy-list/main/socks4",
		"https://raw.githubusercontent.com/casals-ar/proxy-list/main/socks5",
		"https://raw.githubusercontent.com/jetkai/proxy-list/main/online-proxies/txt/proxies.txt",
		"https://raw.githubusercontent.com/a2u/free-proxy-list/master/free-proxy-list.txt",
		"https://api.proxyscrape.com/proxytable.php?nf=true&country=all",
		"https://free-proxy-list.net/",
		"https://raw.githubusercontent.com/proxifly/free-proxy-list/main/proxies/all/data.txt",
	}

	var wg sync.WaitGroup
	proxies := &safeMap{m: make(map[Proxy]bool)}

	proxyChannel := make(chan Proxy, 500) // Buffered channel

	for _, url := range urls {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			for _, proxy := range scrapeProxies(url) {
				proxyType := inferProxyTypeFromURL(url)
				proxyStruct := Proxy{Address: proxy, Type: proxyType}
				proxyChannel <- proxyStruct
			}
		}(url)
	}

	// Close the proxyChannel after all proxies have been sent
	go func() {
		wg.Wait()
		close(proxyChannel)
	}()

	// Use multiple workers to check proxies concurrently
	workerCount := 500
	var workerWg sync.WaitGroup

	for i := 0; i < workerCount; i++ {
		workerWg.Add(1)
		go func() {
			defer workerWg.Done()
			for proxy := range proxyChannel {
				if CheckProxy(proxy) {
					proxies.set(proxy, true)
				}
			}
		}()
	}

	workerWg.Wait()

	cacheMutex.Lock()
	cache = append(cache, proxies.getAllKeys()...)
	cacheMutex.Unlock()
}

// GetGoodProxy returns a good proxy. If no cached proxies are available, it fetches and stores good proxies.
func GetGoodProxy() Proxy {
	cacheMutex.Lock()
	if len(cache) == 0 {
		cacheMutex.Unlock()
		FetchAndStoreGoodProxies()
		cacheMutex.Lock()
	}
	if len(cache) > 0 {
		proxy := cache[0]
		cache = cache[1:]
		cache = append(cache, proxy)
		cacheMutex.Unlock()
		return proxy
	}
	cacheMutex.Unlock()
	return Proxy{}
}

func GetAllProxies() []Proxy {
	cacheMutex.Lock()
	if len(cache) == 0 {
		cacheMutex.Unlock()
		FetchAndStoreGoodProxies()
		cacheMutex.Lock()
	}
	if len(cache) > 0 {
		cacheMutex.Unlock()
		return cache
	}
	cacheMutex.Unlock()
	return []Proxy{}
}

func init() {
    go RecheckGoodProxies(10 * time.Minute)
}