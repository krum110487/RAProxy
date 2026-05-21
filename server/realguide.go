package server

import (
	"bytes"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type RealGuideData struct {
	GameID   string       `json:"gameId"`
	Title    string       `json:"title"`
	ReviewID string       `json:"reviewId"`
	TopImage string       `json:"topImage"`
	Rating   string       `json:"rating"`
	Reviews  []ReviewData `json:"reviews"`
}

type ReviewData struct {
	GameID              string `json:"gameId"`
	UserID              string `json:"userId"`
	Username            string `json:"username"`
	Reviewer            string `json:"reviewer"`
	Location            string `json:"location"`
	FavoriteGames       string `json:"favoriteGames"`
	Version             string `json:"version"`
	RatingOverall       int    `json:"ratingOverall"`
	RatingGraphics      int    `json:"ratingGraphics"`
	RatingLearningCurve int    `json:"ratingLearningCurve"`
	RatingSound         int    `json:"ratingSound"`
	RatingLastingAppeal int    `json:"ratingLastingAppeal"`
	Comments            string `json:"comments"`
}

type Realguide struct {
	ServerRoot    string
	GameList      *GameInfo
	ReviewersInfo *ReviewersInfo
	DB            *DB
}

// Example request: http://127.0.0.1:33500/server/content/realguide.real.com/games/arcade/?s=luxor2_review
func (rg *Realguide) HandlerRealGuide(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case rg.ServerRoot + "/realguide.real.com/games/arcade/index.html":
		fallthrough
	case rg.ServerRoot + "/realguide.real.com/games/arcade/":
		rg.reviewHandler(w, r)
	default:
		staticHandler(rg.ServerRoot, w, r)
	}
}

func (rg *Realguide) reviewHandler(w http.ResponseWriter, r *http.Request) {
	reviewID := getQuery(r, "s", "")
	gameID := strings.ReplaceAll(reviewID, "_review", "")

	if gameID == "" {
		http.Error(w, "missing ?s= parameter", http.StatusBadRequest)
		return
	}

	game := rg.GameList.GetGameByGameID(gameID)
	if game == nil {
		http.Error(w, "game not found: "+gameID, http.StatusNotFound)
		return
	}

	reviews, err := rg.DB.GetReviewsByGameID(gameID)
	if err != nil {
		log.Printf("realguide: db error for %s: %v", gameID, err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	data := RealGuideData{
		GameID:   gameID,
		ReviewID: reviewID,
		Title:    game.Title,
		TopImage: game.ReviewTopImage,
		Reviews:  reviews,
	}

	rg.renderReviewPage(w, data)
}

func (rg *Realguide) renderReviewPage(w http.ResponseWriter, data RealGuideData) {
	tmplPath, err := filepath.Abs(filepath.Join(".", rg.ServerRoot, "realguide.real.com/games/arcade/index.html"))
	if err != nil {
		http.Error(w, "template path error", http.StatusInternalServerError)
		return
	}

	raw, err := os.ReadFile(tmplPath)
	if err != nil {
		http.Error(w, "template not found", http.StatusInternalServerError)
		return
	}

	tmpl, err := template.New("realguide").Parse(string(raw))
	if err != nil {
		log.Printf("realguide: template parse: %v", err)
		http.Error(w, "template error", http.StatusInternalServerError)
		return
	}

	// Build review maps to match the template's lowercase camelCase field names.
	reviewMaps := make([]map[string]interface{}, len(data.Reviews))
	for i, rev := range data.Reviews {
		reviewMaps[i] = map[string]interface{}{
			"reviewer":            rev.Reviewer,
			"version":             rev.Version,
			"location":            rev.Location,
			"favoriteGames":       rev.FavoriteGames,
			"ratingOverall":       rev.RatingOverall,
			"ratingGraphics":      rev.RatingGraphics,
			"ratingLearningCurve": rev.RatingLearningCurve,
			"ratingSound":         rev.RatingSound,
			"ratingLastingAppeal": rev.RatingLastingAppeal,
			"comments":            rev.Comments,
		}
	}

	tplData := map[string]interface{}{
		"gameId":   data.GameID,
		"reviewId": data.ReviewID,
		"title":    data.Title,
		"topImage": data.TopImage,
		"Rating":   data.Rating,
		"reviews":  reviewMaps,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, tplData); err != nil {
		log.Printf("realguide: template execute: %v", err)
		http.Error(w, "render error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(buf.Bytes())
}
