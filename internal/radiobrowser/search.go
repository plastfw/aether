package radiobrowser

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var apiBaseURL = "https://de1.api.radio-browser.info/json/stations/search"

type Station struct {
	Name        string `json:"name"`
	URL         string `json:"url"`
	URLResolved string `json:"url_resolved"`
	Tags        string `json:"tags"`
	Codec       string `json:"codec"`
	Bitrate     int    `json:"bitrate"`
	Votes       int    `json:"votes"`
	ClickCount  int    `json:"clickcount"`
	Country     string `json:"country"`
}

func Search(query string, limit int, offset int) ([]Station, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, fmt.Errorf("search query is empty")
	}
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	params := url.Values{}
	params.Set("name", query)
	params.Set("tag", query)
	params.Set("hidebroken", "true")
	params.Set("order", "clickcount")
	params.Set("reverse", "true")
	params.Set("limit", fmt.Sprint(limit))
	params.Set("offset", fmt.Sprint(offset))

	endpoint := apiBaseURL + "?" + params.Encode()
	client := http.Client{Timeout: 12 * time.Second}
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "aether/0.1 (+https://github.com/)")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("radio-browser status: %s", resp.Status)
	}

	var stations []Station
	if err := json.NewDecoder(resp.Body).Decode(&stations); err != nil {
		return nil, err
	}
	return stations, nil
}

func (s Station) StreamURL() string {
	if strings.TrimSpace(s.URLResolved) != "" {
		return s.URLResolved
	}
	return s.URL
}
