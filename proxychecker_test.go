package proxychecker

import (
	"fmt"
	"testing"
)

func TestGetGoodProxies(t *testing.T) {
    proxies := GetAllProxies()
    if len(proxies) < 1 {
        t.Fatal("No good proxy found")
    }
	proxy := GetGoodProxy()
	fmt.Printf("Example proxy: %s\n\n", proxy)
    fmt.Printf("%d good proxies found:", len(proxies))
}