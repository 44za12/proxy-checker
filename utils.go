package proxychecker

import (
	"context"
	"io"
	"net/http"
	"os"
	"regexp"
)

func (pc *ProxyChecker) makeRequest(ctx context.Context, url string) (string, error) {
    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return "", err
    }

    for key, value := range pc.Headers {
        req.Header.Set(key, value)
    }

    resp, err := pc.Client.Do(req)
	if err != nil {
        return "", err
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return "", err
    }
	return string(body), nil
}

func isValidProxyFormat(proxy string) bool {
    proxyPattern := `^\d{1,3}(\.\d{1,3}){3}:\d{1,5}$`
    r := regexp.MustCompile(proxyPattern)
    return r.MatchString(proxy)
}

func (pc *ProxyChecker) SaveProxiesToFile(filename string) error {
    pc.CacheLock.Lock()
    defer pc.CacheLock.Unlock()

    file, err := os.Create(filename)
    if err != nil {
        return err
    }
    defer file.Close()

    for _, proxy := range pc.Cache {
        fullProxy := proxy.Address
        _, err := file.WriteString(fullProxy + "\n")
        if err != nil {
            return err
        }
    }

    return nil
}