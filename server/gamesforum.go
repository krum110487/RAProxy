package server

import "net/http"

type GamesForum struct {
	ServerRoot    string
	GameList      *GameInfo
	ReviewersInfo *ReviewersInfo
}

func (rg *GamesForum) HandlerGamesForum(w http.ResponseWriter, r *http.Request) {
	//TODO: Implement the RealGuide stuff...
	switch r.URL.Path {
	//case sb.ServerRoot + "/switchboard.real.com/arcade/sites.html":
	default:
		staticHandler(rg.ServerRoot, w, r)
	}
}
