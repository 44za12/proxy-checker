package proxychecker

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

func (pc *ProxyChecker) checkProxy(ctx context.Context, p Proxy, proxyTypes []string) (string, bool) {
    if !isValidProxyFormat(p.Address) {
        return "", false
    }
    randomServer := httpServers[rand.Intn(len(httpServers))]
    proxyIP := strings.Split(p.Address, ":")[0]
    results := make(chan string, len(proxyTypes))
    var wg sync.WaitGroup
    for _, proxyType := range proxyTypes {
        wg.Add(1)
        go func(pt string) {
            defer wg.Done()
            proxyURL, err := url.Parse(fmt.Sprintf("%s://%s", strings.ToLower(pt), p.Address))
            if err != nil {
                return
            }
            transport := &http.Transport{
                Proxy: http.ProxyURL(proxyURL),
                DialContext: (&net.Dialer{
                    Timeout: 10 * time.Second,
                }).DialContext,
            }
            localClient := &http.Client{
                Transport: transport,
                Timeout:   pc.Client.Timeout,
            }
            req, err := http.NewRequestWithContext(ctx, "GET", randomServer, nil)
            if err != nil {
                return
            }
			for key, value := range pc.Headers {
				req.Header.Set(key, value)
			}
            resp, err := localClient.Do(req)
            if err != nil {
                return
            }
            defer resp.Body.Close()
            body, err := io.ReadAll(resp.Body)
            if err != nil || resp.StatusCode != http.StatusOK {
                return
            }
            ip := strings.TrimSpace(string(body))
            if ip == proxyIP {
                select {
                case results <- pt:
                default:
                }
            }
        }(proxyType)
    }
    go func() {
        wg.Wait()
        close(results)
    }()
    select {
    case result := <-results:
        if result != "" {
            fullAddress := fmt.Sprintf("%s://%s", result, p.Address)
            pc.CacheLock.Lock()
            pc.Cache = append(pc.Cache, Proxy{Address: fullAddress, Type: result})
            pc.CacheLock.Unlock()
            pc.Proxies.Store(Proxy{Address: p.Address, Type: result}, true)
            return result, true
        }
    case <-ctx.Done():
        return "", false
    }
    return "", false
}

func (pc *ProxyChecker) updateProxies(ctx context.Context) error {
    scrapedProxies, err := pc.scrapeProxies(ctx)
    if err != nil {
        return err
    }
    fmt.Println("Number of proxies scraped:", len(scrapedProxies))
    semaphore := make(chan struct{}, pc.ConcurrencyLimit)
    var wg sync.WaitGroup

    for _, proxy := range scrapedProxies {
        wg.Add(1)
        go func(p Proxy) {
            defer wg.Done()
            semaphore <- struct{}{}
            result, valid := pc.checkProxy(ctx, p, proxyTypes)
            <-semaphore
            if valid {
                fullAddress := fmt.Sprintf("%s://%s", result, p.Address)
                pc.Proxies.Store(fullAddress, Proxy{Address: fullAddress, Type: result})
            }
        }(proxy)
    }
    wg.Wait()
    return nil
}

