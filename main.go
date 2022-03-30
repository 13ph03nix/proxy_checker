// Version = "0.0.1"

package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"golang.org/x/net/proxy"
)

type proxyInfo struct {
	Country string `json:"country_name"`
	State   string `json:"state"`
	Timeout time.Duration
	Alive   bool
}

func main() {

	// concurrency flag
	var concurrency int
	flag.IntVar(&concurrency, "c", 128, "set the concurrency level")

	// timeout flag
	var to int
	flag.IntVar(&to, "t", 10000, "timeout (milliseconds)")

	// skip http proxy check
	var skip bool
	flag.BoolVar(&skip, "s", false, "only socks5, skip the http proxy check")

	flag.Parse()

	// make an actual time.Duration out of the timeout
	timeout := time.Duration(to * 1000000)

	socks5Addrs := make(chan string)
	httpAddrs := make(chan string)
	output := make(chan string)

	const format = "  %-21s  --  %-21s  --  %-21s  --  %-21s  --  %-v"

	// SOCKS5 checker workers
	var socks5WG sync.WaitGroup
	for i := 0; i < concurrency/2; i++ {
		socks5WG.Add(1)
		go func() {
			for address := range socks5Addrs {
				info := checkProxySOCKS5(address, timeout)
				if info.Alive {
					output <- fmt.Sprintf(format, address, "socks5", info.Country, info.State, info.Timeout)
					continue
				} else if !skip {
					httpAddrs <- address
				}
			}
			socks5WG.Done()
		}()
	}

	// HTTP workers
	var httpWG sync.WaitGroup
	for i := 0; i < concurrency/2; i++ {
		httpWG.Add(1)
		go func() {
			for address := range httpAddrs {
				info := checkProxyHTTP(address, timeout)
				if info.Alive {
					output <- fmt.Sprintf(format, address, "http", info.Country, info.State, info.Timeout)
					continue
				}
			}
			httpWG.Done()
		}()
	}

	// Close the httpURLs channel when the SOCKS5 workers are done
	go func() {
		socks5WG.Wait()
		close(httpAddrs)
	}()

	// Output worker
	var outputWG sync.WaitGroup
	outputWG.Add(1)
	go func() {
		for o := range output {
			fmt.Println(o)
		}
		outputWG.Done()
	}()

	// Close the output channel when the HTTP workers are done
	go func() {
		httpWG.Wait()
		close(output)
	}()

	// accept domains on stdin
	sc := bufio.NewScanner(os.Stdin)
	for sc.Scan() {
		// submit proxy checks
		socks5Addrs <- sc.Text()
	}

	close(socks5Addrs)

	// check there were no errors reading stdin (unlikely)
	if err := sc.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to read input: %s\n", err)
	}

	// Wait until the output waitgroup is done
	outputWG.Wait()
}

func checkProxySOCKS5(address string, timeout time.Duration) *proxyInfo {
	var info proxyInfo
	d := net.Dialer{
		Timeout:   timeout,
		KeepAlive: timeout,
	}
	dialer, _ := proxy.SOCKS5("tcp", address, nil, &d)
	httpClient := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			DisableKeepAlives: true,
			Dial:              dialer.Dial,
		},
	}
	startTime := time.Now()
	response, err := httpClient.Get("https://geoip-db.com/json/")

	if err != nil {
		return &info
	}
	info.Alive = true
	defer response.Body.Close()

	body, _ := ioutil.ReadAll(response.Body)
	info.Timeout = time.Since(startTime)
	json.Unmarshal(body, &info)
	return &info
}

func checkProxyHTTP(address string, timeout time.Duration) *proxyInfo {
	var info proxyInfo
	proxyURL, _ := url.Parse("http://" + address)

	httpClient := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			DisableKeepAlives: true,
			Proxy:             http.ProxyURL(proxyURL),
		},
	}
	startTime := time.Now()
	response, err := httpClient.Get("https://geoip-db.com/json/")
	if err != nil {
		return &info
	}
	info.Alive = true
	defer response.Body.Close()

	body, _ := ioutil.ReadAll(response.Body)
	info.Timeout = time.Since(startTime)
	json.Unmarshal(body, &info)
	return &info
}
