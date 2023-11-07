package proxychecker

import (
	"context"
	"log"
	"math/rand"
	"net/http"
	"time"
)

func NewProxyChecker() *ProxyChecker {
    return &ProxyChecker{
        Client: &http.Client{
            Timeout: 20 * time.Second,
        },
        Headers: headers,
        CheckLimit: 100,
        ConcurrencyLimit: 100,
    }
}

func (pc *ProxyChecker) GetGoodProxy(ctx context.Context) (Proxy, error) {
    pc.CacheLock.Lock()
    if len(pc.Cache) == 0 {
        pc.CacheLock.Unlock()
        err := pc.updateProxies(ctx)
        if err != nil {
            return Proxy{}, err
        }
        pc.CacheLock.Lock()
    }
    if len(pc.Cache) > 0 {
        proxy := pc.Cache[0]
        pc.Cache = pc.Cache[1:]
        pc.CacheLock.Unlock()

        schema := ""
        switch proxy.Type {
        case "HTTP":
            schema = "http://"
        case "SOCKS4":
            schema = "socks4://"
        case "SOCKS5":
            schema = "socks5://"
        default:
            schema = "http://"
        }
        fullAddress := schema + proxy.Address
        pc.Proxies.Store(Proxy{Address: fullAddress, Type: proxy.Type}, true)
        return Proxy{Address: fullAddress, Type: proxy.Type}, nil
    }
    pc.CacheLock.Unlock()
    return Proxy{}, nil
}


func (pc *ProxyChecker) GetAllProxies() []Proxy {
    pc.CacheLock.Lock()
    defer pc.CacheLock.Unlock()
    return pc.Cache
}

func (pc *ProxyChecker) ScheduleRecheck(stopChan <-chan struct{}) {
    ticker := time.NewTicker(1 * time.Hour)
    go func() {
        for {
            select {
            case <-ticker.C:
                ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
                defer cancel()
                if err := pc.updateProxies(ctx); err != nil {
                    log.Fatal(err)
                }
            case <-stopChan:
                ticker.Stop()
                return
            }
        }
    }()
}

func init() {
    rand.Seed(time.Now().UnixNano())
}