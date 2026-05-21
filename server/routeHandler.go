package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type RouteHandler struct {
	ServerRoot string
	Games      GameInfo
	Reviewers  ReviewersInfo
	DB         *DB
}

// staticHandler serves a local file, falling back to the Wayback Machine when
// the file is not found. On a successful Wayback fetch the file is saved locally
// so subsequent requests are served from disk.
func staticHandler(serverRoot string, w http.ResponseWriter, r *http.Request) {
	localPath := "." + r.URL.Path
	if info, err := os.Stat(localPath); err == nil && !info.IsDir() {
		http.FileServer(http.Dir("./")).ServeHTTP(w, r)
		return
	} else if err == nil && info.IsDir() {
		// Let the file server handle directory/index serving.
		http.FileServer(http.Dir("./")).ServeHTTP(w, r)
		return
	}

	// File not found locally — try archive.org.
	originalURL := originalURLFromPath(serverRoot, r.URL.Path)
	if originalURL == "" {
		http.NotFound(w, r)
		return
	}

	log.Printf("wayback: fetching %s", originalURL)
	content, contentType, err := fetchFromWayback(originalURL)
	if err != nil {
		log.Printf("wayback: %v", err)
		http.NotFound(w, r)
		return
	}

	saveLocally(localPath, content)
	w.Header().Set("Content-Type", contentType)
	w.Write(content)
}

func (rh RouteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sb := Switchboard{ServerRoot: rh.ServerRoot, GameList: &rh.Games, ReviewersInfo: &rh.Reviewers}
	rg := Realguide{ServerRoot: rh.ServerRoot, GameList: &rh.Games, ReviewersInfo: &rh.Reviewers, DB: rh.DB}
	gf := GamesForum{ServerRoot: rh.ServerRoot, GameList: &rh.Games, ReviewersInfo: &rh.Reviewers}

	//Custom handling switchboard
	if strings.HasPrefix(r.URL.String(), rh.ServerRoot+"/switchboard.real.com") {
		sb.HandleSwitchboard(w, r)
		return
	}

	//Custom handling reviews
	if strings.HasPrefix(r.URL.String(), rh.ServerRoot+"/realguide.real.com") {
		rg.HandlerRealGuide(w, r)
		return
	}

	//Custom handling messageBoard
	if strings.HasPrefix(r.URL.String(), rh.ServerRoot+"/gamesforum.real.com") {
		gf.HandlerGamesForum(w, r)
		return
	}

	//If not one of the items above, just serve the file.
	staticHandler(rh.ServerRoot, w, r)
}

func (rh *RouteHandler) LoadGameInfo(filename string) {
	// Open the JSON file
	fp, err := filepath.Abs("." + rh.ServerRoot + "/" + filename)
	if err != nil {
		fmt.Println("Cannot find path:", err)
		return
	}

	file, err := os.Open(fp)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	// Decode the JSON data into the struct
	var gl GameInfo
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&gl)
	if err != nil {
		fmt.Println("Error decoding JSON:", err)
		return
	}

	rh.Games = gl
}

// MigrateJSONReviews reads all review JSON files under
// {ServerRoot}/realguide.real.com/games/arcade/<gameId>/*.json
// and inserts them into the DB, enriching each with data from reviewers.json.
// Safe to call on every startup — INSERT OR IGNORE skips already-imported rows.
func (rh *RouteHandler) MigrateJSONReviews() {
	arcadeDir, err := filepath.Abs(filepath.Join(".", rh.ServerRoot, "realguide.real.com", "games", "arcade"))
	if err != nil {
		fmt.Println("MigrateJSONReviews: path error:", err)
		return
	}

	_ = filepath.WalkDir(arcadeDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".json") || strings.Contains(d.Name(), "template") {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			fmt.Println("MigrateJSONReviews: read error:", err)
			return nil
		}

		var rev ReviewData
		if err := json.Unmarshal(data, &rev); err != nil {
			fmt.Println("MigrateJSONReviews: decode error:", path, err)
			return nil
		}
		if rev.GameID == "" || rev.UserID == "" {
			return nil
		}

		if reviewer := rh.Reviewers.GetReviewerByUserID(rev.UserID); reviewer != nil {
			rev.Username = reviewer.Username
			rev.Reviewer = reviewer.Username
			rev.Location = reviewer.Location
			rev.FavoriteGames = strings.Join(reviewer.FavoriteGames, ", ")
		}
		if rev.Reviewer == "" {
			rev.Reviewer = rev.UserID
		}

		if err := rh.DB.InsertLocalReview(rev); err != nil {
			fmt.Printf("MigrateJSONReviews: insert %s: %v\n", path, err)
		}
		return nil
	})
}

func (rh *RouteHandler) LoadReviewerInfo(filename string) {
	// Open the JSON file
	fp, err := filepath.Abs("." + rh.ServerRoot + "/" + filename)
	if err != nil {
		fmt.Println("Cannot find path:", err)
		return
	}

	file, err := os.Open(fp)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	// Decode the JSON data into the struct
	var ri ReviewersInfo
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&ri)
	if err != nil {
		fmt.Println("Error decoding JSON:", err)
		return
	}

	rh.Reviewers = ri
}

/*
---------------------------------------------------------------------------------------
--------------------------------- Helper Functions ------------------------------------
---------------------------------------------------------------------------------------
*/

func redirect(w http.ResponseWriter, r *http.Request, newUrl string) {
	//fmt.Println("\tRedirect: ", r.URL.String())
	invalidateCache(w)
	http.Redirect(w, r, newUrl, http.StatusMovedPermanently)
}

func missingFile(w http.ResponseWriter, r *http.Request, message string) {
	fmt.Println("\tError: ", message)
	invalidateCache(w)
	http.Error(w, message, 404)
}

func invalidateCache(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-cache, private, max-age=0")
	w.Header().Set("Expires", time.Unix(0, 0).Format(http.TimeFormat))
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("X-Accel-Expires", "0")
}

func getQuery(r *http.Request, Param string, Default string) string {
	paramVal := r.URL.Query()[Param]
	if len(paramVal) > 0 {
		return paramVal[0]
	} else {
		return Default
	}
}

func setQuery(r *http.Request, Param string, Value string) {
	q := r.URL.Query()
	q.Set(Param, Value)
	r.URL.RawQuery = q.Encode()
}

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, fs.ErrNotExist) {
		_, err := os.Stat("." + path)
		if err == nil {
			return true, nil
		}
		return false, nil
	}
	return false, err
}

func IntConv(arg interface{}) (int, error) {
	switch x := arg.(type) {
	case int:
		return x, nil
	case string:
		return strconv.Atoi(x)
	}
	return 0, errors.New("IntConv: invalid argument ")
}
