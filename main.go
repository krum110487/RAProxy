package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/elazarl/goproxy"
	"github.com/krum110487/RAProxy/server"
)

/* func main() {
	//setupProxy()

	//Overwrite the other handlers
	http.HandleFunc("/games/arcade", arcadeHandler)
	http.HandleFunc("/arcade/download.html", downloadHandler)
	http.HandleFunc("/arcade/feeds.html", feedsHandler)
	http.HandleFunc("/arcade/sites.html", sitesHandler)
	http.HandleFunc("/pics/real/games/slideshow/", slideshowHandler)
	http.HandleFunc("/track/", trackHandler)

	//Handle the static files...
	http.HandleFunc("/", staticHandler)

	//Listen and Serve
	http.ListenAndServe(":80", nil)

} */

func main() {
	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = false

	//TODO: Update to Dynamic path
	ServerRoot := "/server/content"

	//Handle the requests seperate for when it requests
	rh := server.RouteHandler{ServerRoot: ServerRoot}

	//Load Game specific info into memory
	rh.LoadGameInfo("gameInfo.json")

	//Load Reviewer Specific info into memory
	rh.LoadReviewerInfo("reviewers.json")

	//Open SQLite DB, migrate schema, and seed from existing JSON review files
	db, err := server.NewDB("./reviews.db")
	if err != nil {
		log.Fatal("open DB:", err)
	}
	rh.DB = db
	rh.MigrateJSONReviews()

	//Start Discord sync if DISCORD_BOT_TOKEN is set (read-only, RealGuides category)
	ds, discordErr := server.NewDiscordSync(db, &rh.Games)
	if discordErr != nil {
		log.Printf("Discord sync disabled: %v", discordErr)
	} else {
		go ds.SyncPeriodically(10 * time.Minute)
	}

	proxy.NonproxyHandler = rh

	//Capture every request and route it to content
	proxy.OnRequest().DoFunc(func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		//fmt.Printf("\nProxy Request: http://%s\n", r.URL.Host+r.URL.Path)

		localFile := "http://127.0.0.1:33500" + ServerRoot + "/" + r.URL.Hostname() + r.URL.Path
		if r.URL.RawQuery != "" {
			localFile += "?" + r.URL.RawQuery
		}

		//fmt.Println("\tProxied to: " + localFile)

		//Make request to local file.
		client := &http.Client{}
		proxyReq, _ := http.NewRequest(r.Method, localFile, r.Body)
		proxyReq.Header = r.Header
		proxyResp, _ := client.Do(proxyReq)

		if proxyResp.StatusCode >= 400 {
			fmt.Printf("Failed request (%d): %s %s\n", proxyResp.StatusCode, proxyResp.Request.Method, proxyReq.URL)
		}

		//Return the new response
		return r, proxyResp
	})

	//TODO: Make non-static port...
	log.Fatal(http.ListenAndServe(":33500", proxy))
}
