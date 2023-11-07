package proxychecker

import (
	"context"
	"regexp"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

func (pc *ProxyChecker) scrapeProxies(ctx context.Context) ([]Proxy, error) {
    var wg sync.WaitGroup
    var mu sync.Mutex
    var totalScraped []Proxy
    var scrapeErrors []error

    for _, url := range urls {
        wg.Add(1)
        go func(u string) {
            defer wg.Done()
            scraped, err := pc.scrapeProxy(ctx, u)
            if err != nil {
                scrapeErrors = append(scrapeErrors, err)
                return
            }
            mu.Lock()
            totalScraped = append(totalScraped, scraped...)
            for _, proxy := range scraped {
                pc.Proxies.Store(proxy.Address, proxy)
            }
            mu.Unlock()
        }(url)
    }
    wg.Wait()

    if len(scrapeErrors) > 0 {
        println("Number of errors:", len(scrapeErrors))
    }

    return totalScraped, nil
}



func (pc *ProxyChecker) scrapeProxy(ctx context.Context, url string) ([]Proxy, error) {
    bodyString, err := pc.makeRequest(ctx, url)
    if err != nil {
		return nil, err
	}
    proxyPattern := `(\d{1,3}(\.\d{1,3}){3}:\d{1,5})`
    r := regexp.MustCompile(proxyPattern)
    matches := r.FindAllString(bodyString, -1)
    if matches != nil {
        proxies := make([]Proxy, 0, len(matches))
        for _, match := range matches {
            if isValidProxyFormat(match) {
                proxies = append(proxies, Proxy{Address: match})
            }
        }
        if len(proxies) > 0 {
            return proxies, nil
        }
    }
    doc, err := goquery.NewDocumentFromReader(strings.NewReader(bodyString))
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
				if proxy.Address != "" && isValidProxyFormat(proxy.Address){
					proxies = append(proxies, proxy)
				}
			}
		})
	})

	return proxies
}