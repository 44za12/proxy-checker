package proxychecker

import (
	"context"
	"testing"
	"time"
)

// func TestScrapeProxiesFromHTML(t *testing.T) {
// 	pc := NewProxyChecker()
//     ctx := context.TODO()
// 	testURL := "https://checkerproxy.net/api/archive/2023-11-07"
// 	proxies, err := pc.scrapeProxies(ctx, testURL)
// 	if err != nil {
// 		t.Fatalf("Error scraping proxies: %v", err)
// 	}
// 	if len(proxies) == 0 {
// 		t.Fatal("No proxies found in the HTML")
// 	}
// 	fmt.Printf("Found %d proxies", len(proxies))
// }

// func TestGetGoodProxies(t *testing.T) {
//     pc := NewProxyChecker()
//     ctx := context.TODO()
//     proxy, err := pc.GetGoodProxy(ctx)
//     if err != nil {
//         log.Fatal(err)
//     }
//     _ = pc.SaveProxiesToFile("proxies.txt")
//     proxies := pc.GetAllProxies()
//     if len(proxies) < 1 {
//         t.Fatal("No good proxies found")
//     }
//     t.Logf("Example proxy with schema: %s", proxy.Address)
//     t.Logf("%d good proxies found", len(proxies))
// }

// TestCheckProxy tests the checkProxy function
func TestCheckProxy(t *testing.T) {
    pc := NewProxyChecker()
    proxy := Proxy{
        Address: "36.73.154.113:8080",
    }
    proxyTypes := []string{"http", "socks4", "socks5"}
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    proxyType, valid := pc.checkProxy(ctx, proxy, proxyTypes)
    println(proxyType, valid)
}

func TestUpdateProxy(t *testing.T) {
    pc := NewProxyChecker()
    // proxyToTest := fmt.Sprintf("https://checkerproxy.net/api/archive/%s", time.Now().Format("2006-01-02"))
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    err := pc.updateProxies(ctx)
    pc.SaveProxiesToFile("proxies.txt")
    if err != nil {
        t.Errorf("Expected no error, got %s", err)
    }
    println(len(pc.Cache))
}
