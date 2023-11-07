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

	"github.com/PuerkitoBio/goquery"
)

type Proxy struct {
    Address string
    Type    string
}

type ProxyChecker struct {
    Cache      []Proxy
    CacheLock  sync.Mutex
    Client     *http.Client
    Headers    map[string]string
    Proxies    sync.Map
}

func NewProxyChecker() *ProxyChecker {
    headers := map[string]string{
        "User-Agent":      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3",
        "Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
        "Accept-Language": "en-US,en;q=0.5",
    }
    return &ProxyChecker{
        Client: &http.Client{
            Timeout: 10 * time.Second,
        },
        Headers: headers,
    }
}

func (pc *ProxyChecker) makeRequest(ctx context.Context, url string) (*http.Response, error) {
    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, err
    }
    for key, value := range pc.Headers {
        req.Header.Set(key, value)
    }
    return pc.Client.Do(req)
}

func (pc *ProxyChecker) RecheckGoodProxies(ctx context.Context, interval time.Duration) {
    ticker := time.NewTicker(interval)
    defer ticker.Stop()
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            pc.CheckAndUpdateProxies(ctx)
        }
    }
}

func (pc *ProxyChecker) CheckAndUpdateProxies(ctx context.Context) {
    pc.CacheLock.Lock()
    defer pc.CacheLock.Unlock()
    
    validProxies := make([]Proxy, 0)
    proxyTypes := []string{"HTTP", "SOCKS4", "SOCKS5"}
    
    for _, proxy := range pc.Cache {
        if proxyType, valid := pc.CheckProxy(ctx, proxy, proxyTypes); valid {
            validProxies = append(validProxies, Proxy{Address: proxy.Address, Type: proxyType})
        }
    }
    
    pc.Cache = validProxies

    if len(validProxies) == 0 {
        pc.FetchAndStoreGoodProxies(ctx)
    }
}

func (pc *ProxyChecker) CheckProxy(ctx context.Context, p Proxy, proxyTypes []string) (string, bool) {
    httpServers := []string{
        "https://httpbin.org/ip",
        "https://icanhazip.com",
        "https://ifconfig.me/ip",
        "https://api.ipify.org",
        "https://ipinfo.io/ip",
        "https://ip.42.pl/raw",
        "https://checkip.amazonaws.com",
        "https://wtfismyip.com/text",
        "https://curlmyip.net",
        "https://ipapi.co/ip",
        "https://ipecho.net/plain",
        "https://ip.tyk.nu",
        "https://www.cloudflare.com/cdn-cgi/trace",
        "https://www.google.com",
        "https://www.youtube.com",
        "https://www.facebook.com",
        "https://www.amazon.com",
        "https://www.instagram.com",
        "https://www.whatsapp.com",
        "https://www.linkedin.com",
        "https://www.bing.com",
        "https://aws.amazon.com/",
        "https://www.x.com",
        "https://www.alibaba.com",
        "https://www.apple.com",
	}

    results := make(chan string, len(proxyTypes))

    for _, proxyType := range proxyTypes {
        go func(pt string) {
            proxyURL, err := url.Parse(fmt.Sprintf("%s://%s", strings.ToLower(pt), p.Address))
            if err != nil {
                results <- ""
                return
            }
            randomServer := httpServers[rand.Intn(len(httpServers))]
            tr := &http.Transport{
                Proxy: http.ProxyURL(proxyURL),
                DialContext: (&net.Dialer{
                    Timeout: 5 * time.Second,
                }).DialContext,
            }
            pc.Client.Transport = tr
            req, err := http.NewRequestWithContext(ctx, "GET", randomServer, nil)
            if err != nil {
                results <- ""
                return
            }
            _, err = pc.Client.Do(req)
            if err != nil {
                results <- ""
                return
            }
            results <- pt
        }(proxyType)
    }

    for i := 0; i < len(proxyTypes); i++ {
        result := <-results
        if result != "" {
            pc.Proxies.Store(Proxy{Address: p.Address, Type: result}, true)
            return result, true
        }
    }

    return "", false
}


func (pc *ProxyChecker) FetchAndStoreGoodProxies(ctx context.Context) {
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
		"https://hidemy.io/en/proxy-list/",
		"https://www.proxy-list.download/HTTP",
		"https://www.proxy-list.download/HTTPS",
		"https://www.proxy-list.download/SOCKS4",
		"https://www.proxy-list.download/SOCKS5",
		"https://free-proxy-list.net/",
		"https://www.sslproxies.org/",
		"http://www.us-proxy.org/",
		"https://www.proxynova.com/proxy-server-list/",
		"https://proxy-list.org/english/index.php",
		"https://hidemy.name/en/proxy-list/",
		"http://proxydb.net/",
		"https://multiproxy.org/txt_all/proxy.txt",
		"https://raw.githubusercontent.com/proxifly/free-proxy-list/main/proxies/all/data.txt",
	}

    var wg sync.WaitGroup

    for _, ep := range urls {
        wg.Add(1)
        go func(u string) {
            defer wg.Done()
            scrapedProxies, err := pc.scrapeProxies(ctx, u)
            if err != nil {
                fmt.Println("Error scraping proxies:", err)
                return
            }
            for _, proxy := range scrapedProxies {
                _, err := url.Parse(fmt.Sprintf("http://%s", proxy.Address))
                if err == nil {
                    pc.Proxies.Store(Proxy{Address: proxy.Address, Type: "HTTP"}, false)
                }
            }
        }(ep)
    }

    wg.Wait()

    pc.CacheLock.Lock()
    pc.Proxies.Range(func(key, value interface{}) bool {
        pc.Cache = append(pc.Cache, key.(Proxy))
        return true
    })
    pc.CacheLock.Unlock()
}

func (pc *ProxyChecker) scrapeProxies(ctx context.Context, url string) ([]Proxy, error) {
    resp, err := pc.makeRequest(ctx, url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }

    html := string(body)
    doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
    if err != nil {
        return nil, err
    }

    return pc.scrapeProxiesFromHTML(doc), nil
}

func (pc *ProxyChecker) scrapeProxiesFromHTML(doc *goquery.Document) []Proxy {
    var proxies []Proxy
	doc.Find("table").Each(func(_ int, tablehtml *goquery.Selection) {
		headers := []string{}
		tablehtml.Find("tr").Each(func(rowIndex int, rowhtml *goquery.Selection) {
			row := []string{}
			rowhtml.Find("th, td").Each(func(_ int, cellhtml *goquery.Selection) {
				text := strings.TrimSpace(cellhtml.Text())
				if rowIndex == 0 {
					headers = append(headers, text)
				} else {
					row = append(row, text)
				}
			})

			if rowIndex > 0 {
				proxy := Proxy{}
				for i, cell := range row {
					header := strings.ToLower(headers[i])
					if strings.Contains(header, "ip") || strings.Contains(header, "address") {
						proxy.Address = cell
					} else if strings.Contains(header, "port") {
						proxy.Address += ":" + cell
					}
				}
				if proxy.Address != "" {
					proxies = append(proxies, proxy)
				}
			}
		})
	})

	return proxies
}

func (pc *ProxyChecker) GetGoodProxy() Proxy {
    pc.CacheLock.Lock()
    defer pc.CacheLock.Unlock()
    if len(pc.Cache) > 0 {
        proxy := pc.Cache[0]
        pc.Cache = pc.Cache[1:]
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
        pc.Proxies.Store(Proxy{Address: schema + proxy.Address, Type: proxy.Type}, true)
        return Proxy{Address: schema + proxy.Address, Type: proxy.Type}
    }
    return Proxy{}
}

func (pc *ProxyChecker) GetAllProxies() []Proxy {
    pc.CacheLock.Lock()
    defer pc.CacheLock.Unlock()

    return pc.Cache
}


func init() {
    rand.Seed(time.Now().UnixNano())
}