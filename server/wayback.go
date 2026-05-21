package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var waybackClient = &http.Client{Timeout: 30 * time.Second}

type waybackResponse struct {
	ArchivedSnapshots struct {
		Closest struct {
			Available bool   `json:"available"`
			Status    string `json:"status"`
			URL       string `json:"url"`
			Timestamp string `json:"timestamp"`
		} `json:"closest"`
	} `json:"archived_snapshots"`
}

// fetchFromWayback queries the Wayback Machine availability API for originalURL,
// then downloads the raw (unmodified) snapshot using the id_ modifier so that
// no wayback toolbar or URL rewriting is applied.
func fetchFromWayback(originalURL string) (content []byte, contentType string, err error) {
	apiURL := "https://archive.org/wayback/available?url=" + originalURL
	resp, err := waybackClient.Get(apiURL)
	if err != nil {
		return nil, "", fmt.Errorf("availability check: %w", err)
	}
	defer resp.Body.Close()

	var wb waybackResponse
	if err := json.NewDecoder(resp.Body).Decode(&wb); err != nil {
		return nil, "", fmt.Errorf("decode availability: %w", err)
	}

	c := wb.ArchivedSnapshots.Closest
	if !c.Available || c.Status != "200" {
		return nil, "", fmt.Errorf("not available in wayback: %s", originalURL)
	}

	// Insert id_ after the timestamp so wayback serves the raw original bytes.
	rawURL := strings.Replace(c.URL, c.Timestamp+"/", c.Timestamp+"id_/", 1)

	rawResp, err := waybackClient.Get(rawURL)
	if err != nil {
		return nil, "", fmt.Errorf("fetch raw: %w", err)
	}
	defer rawResp.Body.Close()

	if rawResp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("wayback returned %d for %s", rawResp.StatusCode, rawURL)
	}

	data, err := io.ReadAll(rawResp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("read body: %w", err)
	}

	ct := rawResp.Header.Get("Content-Type")
	if ct == "" {
		ct = http.DetectContentType(data)
	}
	return data, ct, nil
}

// originalURLFromPath reconstructs the original http URL from a local request
// path like "/server/content/realguide.real.com/4plus/main_.js".
// Returns "" if the path doesn't start with serverRoot.
func originalURLFromPath(serverRoot, urlPath string) string {
	prefix := serverRoot + "/"
	trimmed := strings.TrimPrefix(urlPath, prefix)
	if trimmed == urlPath {
		return ""
	}
	slash := strings.Index(trimmed, "/")
	if slash < 0 {
		return ""
	}
	return "http://" + trimmed[:slash] + trimmed[slash:]
}

// saveLocally writes content to localPath, creating parent directories as needed.
func saveLocally(localPath string, content []byte) {
	if err := os.MkdirAll(filepath.Dir(localPath), 0755); err != nil {
		log.Printf("wayback: mkdir %s: %v", filepath.Dir(localPath), err)
		return
	}
	if err := os.WriteFile(localPath, content, 0644); err != nil {
		log.Printf("wayback: write %s: %v", localPath, err)
		return
	}
	log.Printf("wayback: saved %s", localPath)
}
