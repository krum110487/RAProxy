package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/elazarl/goproxy"
)

func main() {
	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = true
	proxy.NonproxyHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("Non Proxy Request: %s\n", r.URL.Host+r.URL.Path)
	})

	proxy.OnRequest().DoFunc(func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		fmt.Printf("Proxy Request: %s\n", r.URL.Host+r.URL.Path)
		newURL := *r.URL
		if strings.HasSuffix(newURL.Host, "real.com") {
			newURL.Host = "52.188.84.124"
		}

		if newURL.Host == "game-dl.real.com" {
			newURL.Host = "52.224.233.122"
		}

		//Redirect:
		if strings.HasPrefix(r.URL.Path, "/arcade/sites.html") {
			newURL.Path = "/games/slideshow/twistytracks/index.html"
			res := http.Response{}
			res.StatusCode = 301
			res.Header.Add("Location", newURL.String())
			return r, &res
		}

		//Make the request to the zip server.
		client := &http.Client{}
		proxyReq, err := http.NewRequest(r.Method, newURL.String(), r.Body)
		proxyReq.Header = r.Header
		proxyResp, err := client.Do(proxyReq)

		if proxyResp.StatusCode < 400 {
			fmt.Printf("Failed request with: ", err)
		}

		//Return the new response
		return r, proxyResp
	})

	log.Fatal(http.ListenAndServe(":3127", proxy))
}
