package server

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type Switchboard struct {
	ServerRoot    string
	GameList      *GameInfo
	ReviewersInfo *ReviewersInfo
}

func (sb *Switchboard) HandleSwitchboard(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case sb.ServerRoot + "/switchboard.real.com/arcade/sites.html":
		sb.sitesHandler(w, r)
	case sb.ServerRoot + "/switchboard.real.com/arcade/feed.html":
		fallthrough
	case sb.ServerRoot + "/switchboard.real.com/arcade/feeds.html":
		sb.feedsHandler(w, r)
	case sb.ServerRoot + "/switchboard.real.com/arcade/download.html":
		sb.downloadHandler(w, r)
	default:
		staticHandler(sb.ServerRoot, w, r)
	}
}

func (sb *Switchboard) slideshowHandler(w http.ResponseWriter, r *http.Request) {
	gameId := getQuery(r, "gameid", "")
	Id := getQuery(r, "ID", "")

	//If gameId is found, we just append it, redirect and stop
	if gameId != "" {
		newurl := sb.ServerRoot + "/switchboard.real.com/games/slideshow/" + gameId
		exists, _ := pathExists(newurl)
		if !exists {
			newurl = sb.ServerRoot + "/switchboard.real.com/games/slideshow/!default?gameid=" + gameId
			println("LOGS TEST: " + newurl)
		}
		redirect(w, r, newurl)
		return
	}

	//Gameid and Id are not found, give http response...
	missingFile(w, r, "Missing GameID or ID ("+Id+") is missing from '"+sb.ServerRoot+"/gameInfo.json', Either create the Id in JSON or determine why the GameID is missing.")
	return
}

func (sb *Switchboard) downloadHandler(w http.ResponseWriter, r *http.Request) {
	filePath := r.URL.Query()["file"]
	newurl, err := url.JoinPath(sb.ServerRoot, "/switchboard.real.com/arcade/download", filePath[0])
	if err != nil {
		fmt.Println(err)
		return
	}
	redirect(w, r, newurl)
}

func (sb *Switchboard) sitesHandler(w http.ResponseWriter, r *http.Request) {
	//fmt.Println("\tSite Request: ", r.URL.String())

	gameId := getQuery(r, "gameid", "")
	Id := getQuery(r, "ID", "")

	if Id != "" && gameId == "" {
		intId, err := IntConv(Id)
		if err != nil {
			fmt.Println(err)
			missingFile(w, r, "Error, check logs for details")
			return
		}
		gameId = sb.GameList.FindID(intId)
	}

	setQuery(r, "gameid", gameId)
	setQuery(r, "ID", Id)
	sb.slideshowHandler(w, r)
}

func (sb *Switchboard) feedsHandler(w http.ResponseWriter, r *http.Request) {
	//fmt.Println("\tFeed Request: ", r.URL.String())

	Id := getQuery(r, "Id", "")
	action := getQuery(r, "action", "UNDEFINED_ACTION")

	switch action {
	case "screen_shots":
		setQuery(r, "ID", Id)
		sb.slideshowHandler(w, r)
	default:
		missingFile(w, r, "Unhandled Request (action='"+action+"'): "+r.URL.String())
	}
}

func (sb *Switchboard) arcadeHandler(w http.ResponseWriter, r *http.Request) {
	//fmt.Println("Arcade Request: ", r.URL.String())

	s := r.URL.Query()["s"][0]
	ssplit := strings.Split(s, "_")

	isReview := false
	gameId := ""
	if len(ssplit) > 1 {
		gameId = ssplit[0]
		isReview = ssplit[1] == "review"
	}

	if isReview {
		sb.readReviews(gameId, w, r)
	}
}

func (sb *Switchboard) readReviews(gameId string, w http.ResponseWriter, r *http.Request) {
	//fmt.Println("Request Header", r.Header)

	newurl := sb.ServerRoot + "/realguide.real.com/games/review/" + gameId + "/index.html"
	redirect(w, r, newurl)
}

func (sb *Switchboard) messageBoard(domain string, gameId string, w http.ResponseWriter, r *http.Request) {
	//fmt.Println("Request Header", r.Header)

	newurl := sb.ServerRoot + "/gamesforum.real.com/games/slideshow/" + gameId + "/index.html"
	redirect(w, r, newurl)
}
