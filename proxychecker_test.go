package proxychecker

import (
	"context"
	"testing"
)

func TestGetGoodProxies(t *testing.T) {
    pc := NewProxyChecker()
    ctx := context.TODO()
    pc.FetchAndStoreGoodProxies(ctx)

    proxies := pc.GetAllProxies()
    if len(proxies) < 1 {
        t.Fatal("No good proxies found")
    }
    proxy := pc.GetGoodProxy()
    t.Logf("Example proxy with schema: %s", proxy.Address)
    t.Logf("%d good proxies found", len(proxies))
}

