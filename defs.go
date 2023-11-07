package proxychecker

import (
	"fmt"
	"net/http"
	"sync"
	"time"
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
    CheckLimit int
	ConcurrencyLimit int
}

var (
	httpServers = [7]string{
        "https://httpbin.org/ip",
        "https://icanhazip.com",
        "https://ifconfig.me/ip",
        "https://api.ipify.org",
        "https://ipinfo.io/ip",
        "https://ip.42.pl/raw",
        "https://checkip.amazonaws.com",
    }
	urls = []string{
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
		"https://www.proxynova.com/proxy-server-list/",
		"https://proxy-list.org/english/index.php",
		"https://hidemy.name/en/proxy-list/",
		"https://multiproxy.org/txt_all/proxy.txt",
		"https://raw.githubusercontent.com/proxifly/free-proxy-list/main/proxies/all/data.txt",
		fmt.Sprintf("https://checkerproxy.net/api/archive/%s", time.Now().Format("2006-01-02")),
        fmt.Sprintf("https://checkerproxy.net/api/archive/%s", time.Now().AddDate(0, 0, -1).Format("2006-01-02")),
		fmt.Sprintf("https://checkerproxy.net/api/archive/%s", time.Now().AddDate(0, 0, -2).Format("2006-01-02")),
		fmt.Sprintf("https://checkerproxy.net/api/archive/%s", time.Now().AddDate(0, 0, -3).Format("2006-01-02")),
		fmt.Sprintf("https://checkerproxy.net/api/archive/%s", time.Now().AddDate(0, 0, -4).Format("2006-01-02")),
		fmt.Sprintf("https://checkerproxy.net/api/archive/%s", time.Now().AddDate(0, 0, -5).Format("2006-01-02")),
		fmt.Sprintf("https://checkerproxy.net/api/archive/%s", time.Now().AddDate(0, 0, -6).Format("2006-01-02")),
		fmt.Sprintf("https://checkerproxy.net/api/archive/%s", time.Now().AddDate(0, 0, -7).Format("2006-01-02")),
		fmt.Sprintf("https://checkerproxy.net/api/archive/%s", time.Now().AddDate(0, 0, -8).Format("2006-01-02")),
		fmt.Sprintf("https://checkerproxy.net/api/archive/%s", time.Now().AddDate(0, 0, -9).Format("2006-01-02")),
		fmt.Sprintf("https://checkerproxy.net/api/archive/%s", time.Now().AddDate(0, 0, -10).Format("2006-01-02")),
		"https://vpnoverview.com/privacy/anonymous-browsing/free-proxy-servers/",
		"https://spys.one/",
	}
	headers = map[string]string{
        "User-Agent":      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3",
        "Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
        "Accept-Language": "en-US,en;q=0.5",
    }
	proxyTypes = []string{"http", "socks4", "socks5"}
)