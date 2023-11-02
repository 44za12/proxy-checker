package proxychecker

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"sync"
	"time"
)

var testUrls = []string{
	"http://ifconfig.me",
	"http://httpbin.org/ip",
	"https://api.myip.com",
	"http://checkip.amazonaws.com",
	"https://www.cloudflare.com/cdn-cgi/trace",
	"https://checkip.dyndns.org",
	"http://icanhazip.com",
	"http://www.trackip.net/ip",
	"https://ipinfo.io/ip",
}

var cache = make([]string, 0)
var cacheMutex sync.Mutex

// CheckProxy checks the latency of a proxy and returns true if it is below the threshold
func CheckProxy(proxy string) bool {
	start := time.Now()
	proxyURL, _ := url.Parse("socks5://" + proxy)
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	for _, testUrl := range testUrls {
		req, err := http.NewRequestWithContext(ctx, "GET", testUrl, nil)
		if err != nil {
			continue
		}
		resp, err := client.Do(req)
		if err != nil || resp.StatusCode != 200 {
			continue
		}
		defer resp.Body.Close()
		elapsed := time.Since(start)
		if elapsed.Seconds() <= 60 {
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
	html, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []string{}
	}
	return pattern.FindAllString(string(html), -1)
}

func FetchAndStoreGoodProxies() {
	urls := []string{
		"https://raw.githubusercontent.com/TheSpeedX/PROXY-List/master/socks5.txt",
		"https://raw.githubusercontent.com/ShiftyTR/Proxy-List/master/socks5.txt",
		"https://raw.githubusercontent.com/roosterkid/openproxylist/main/SOCKS5.txt",
	}

	var wg sync.WaitGroup
	proxies := make(map[string]bool)

	for _, url := range urls {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			for _, proxy := range scrapeProxies(url) {
				if CheckProxy(proxy) {
					proxies[proxy] = true
				}
			}
		}(url)
	}
	wg.Wait()

	cacheMutex.Lock()
	for proxy := range proxies {
		cache = append(cache, proxy)
	}
	cacheMutex.Unlock()
}

// GetGoodProxy returns a good proxy. If no cached proxies are available, it fetches and stores good proxies.
func GetGoodProxy() string {
	cacheMutex.Lock()
	if len(cache) == 0 {
		cacheMutex.Unlock()
		FetchAndStoreGoodProxies()
		cacheMutex.Lock()
	}
	if len(cache) > 0 {
		proxy := cache[0]
		cache = cache[1:]
		cacheMutex.Unlock()
		return proxy
	}
	cacheMutex.Unlock()
	return ""
}