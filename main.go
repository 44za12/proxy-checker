package main

import (
	"bufio"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"sync"
	"time"
)

// ProxyChecker is a struct that holds the input and output file paths
type ProxyChecker struct {
	InputFile  string
	OutputFile string
}

// CheckProxy checks the latency of a proxy and returns true if it is below the threshold
func (pc *ProxyChecker) CheckProxy(proxy string) bool {
	start := time.Now()
	proxyURL, _ := url.Parse(proxy)
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", "http://bing.com", nil)
	if err != nil {
		return false
	}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	elapsed := time.Since(start)
	if elapsed.Seconds() > 8 {
		return false
	}
	return true
}

// ProcessFile reads the input file line by line and checks the latency of each proxy
// It then writes the good proxies to the output file
func (pc *ProxyChecker) ProcessFile() {
	file, err := os.Open(pc.InputFile)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	outFile, err := os.Create(pc.OutputFile)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer outFile.Close()

	scanner := bufio.NewScanner(file)
	var wg sync.WaitGroup
	for scanner.Scan() {
		wg.Add(1)
		go func(proxy string) {
			defer wg.Done()
			if pc.CheckProxy(proxy) {
				_, err := outFile.WriteString(proxy + "\n")
				if err != nil {
					fmt.Println(err)
				}
			}
		}(scanner.Text())
	}
	wg.Wait()
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

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: proxychecker <file_to_put_all_scraped_proxies_to> <output_file_of_good_proxies>")
		return
	}

	inputFile := os.Args[1]
	outputFile := os.Args[2]

	// List of URLs to scrape proxies from
	urls := []string{
		"https://www.sslproxies.org/",
		"https://free-proxy-list.net/",
		"https://www.us-proxy.org/",
		"https://www.socks-proxy.net/",
		"https://raw.githubusercontent.com/saschazesiger/Free-Proxies/master/proxies/all.txt",
		"https://raw.githubusercontent.com/andigwandi/free-proxy/main/proxy_list.txt",
		"https://raw.githubusercontent.com/TheSpeedX/PROXY-List/master/socks5.txt",
		"https://raw.githubusercontent.com/TheSpeedX/PROXY-List/master/socks4.txt",
		"https://raw.githubusercontent.com/TheSpeedX/PROXY-List/master/http.txt",
		"https://raw.githubusercontent.com/ShiftyTR/Proxy-List/master/proxy.txt",
		"https://raw.githubusercontent.com/ShiftyTR/Proxy-List/master/socks4.txt",
		"https://raw.githubusercontent.com/ShiftyTR/Proxy-List/master/socks5.txt",
	}

	var wg sync.WaitGroup
	proxies := make(map[string]bool)

	for _, url := range urls {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			for _, proxy := range scrapeProxies(url) {
				proxies[proxy] = true
			}
		}(url)
	}
	wg.Wait()

	f, err := os.Create(inputFile)
	if err != nil {
		return
	}
	defer f.Close()
	for proxy := range proxies {
		f.WriteString(proxy + "\n")
	}

	pc := &ProxyChecker{
		InputFile:  inputFile,
		OutputFile: outputFile,
	}

	start := time.Now()
	pc.ProcessFile()
	elapsed := time.Since(start)
	fmt.Printf("Processed %d proxies in %s\n", pc.CountLines(), elapsed)
}

// CountLines counts the number of lines in the input file
func (pc *ProxyChecker) CountLines() int {
	file, err := os.Open(pc.InputFile)
	if err != nil {
		fmt.Println(err)
		return 0
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	count := 0
	for scanner.Scan() {
		count++
	}
	return count
}
