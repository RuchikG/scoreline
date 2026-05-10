package cricketdata

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/RuchikG/scoreline/internal/cricket"
)

func TestMapCurrentMatches(t *testing.T) {
	raw, err := os.ReadFile("testdata/current_matches.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	var payload currentMatchesResponse
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("unmarshal fixture: %v", err)
	}

	matches := mapCurrentMatches(payload.Data)
	if len(matches) != 1 {
		t.Fatalf("len(matches) = %d, want 1", len(matches))
	}
	match := matches[0]
	if match.ID != "sample-current-1" {
		t.Fatalf("match.ID = %q", match.ID)
	}
	if match.Status != cricket.MatchStatusLive {
		t.Fatalf("match.Status = %q", match.Status)
	}
	if match.CurrentScoreSummary != "India Inning 1 147/3 (15.2 ov)" {
		t.Fatalf("summary = %q", match.CurrentScoreSummary)
	}
}

func TestCurrentMatchesUsesCache(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if got := r.URL.Query().Get("apikey"); got != "test-key" {
			t.Fatalf("apikey = %q", got)
		}
		_, _ = w.Write([]byte(`{"status":"success","data":[{"id":"m1","name":"India vs Australia","matchType":"t20","status":"Live","venue":"MCG","dateTimeGMT":"2026-01-01T10:00:00","teams":["India","Australia"],"score":[{"r":42,"w":1,"o":5.2,"inning":"India Inning 1"}]}]}`))
	}))
	defer server.Close()

	client := testClient(server.URL)
	for i := 0; i < 2; i++ {
		matches, err := client.CurrentMatches(context.Background(), 0)
		if err != nil {
			t.Fatalf("CurrentMatches: %v", err)
		}
		if len(matches) != 1 {
			t.Fatalf("len(matches) = %d, want 1", len(matches))
		}
	}
	if requests != 1 {
		t.Fatalf("requests = %d, want 1", requests)
	}
}

func TestMatchInfoMapsDetails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/match_info" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		if got := r.URL.Query().Get("id"); got != "m1" {
			t.Fatalf("id = %q", got)
		}
		_, _ = w.Write([]byte(`{"status":"success","data":{"id":"m1","name":"India vs Australia","matchType":"odi","status":"India won by 3 wickets","venue":"MCG","dateTimeGMT":"2026-01-01T10:00:00","teams":["India","Australia"],"score":[{"r":250,"w":7,"o":48.4,"inning":"India Inning 2"}],"tossWinner":"Australia","tossChoice":"bat"}}`))
	}))
	defer server.Close()

	client := testClient(server.URL)
	details, err := client.MatchInfo(context.Background(), "m1")
	if err != nil {
		t.Fatalf("MatchInfo: %v", err)
	}
	if details.Match.ID != "m1" {
		t.Fatalf("Match.ID = %q", details.Match.ID)
	}
	if details.Match.Status != cricket.MatchStatusFinished {
		t.Fatalf("status = %q", details.Match.Status)
	}
	if details.Toss != "Australia chose to bat" {
		t.Fatalf("toss = %q", details.Toss)
	}
	if details.Result != "India won by 3 wickets" {
		t.Fatalf("result = %q", details.Result)
	}
}

func TestCricketDataFailureDoesNotExposeAPIKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"status":"failure","reason":"Invalid API key"}`))
	}))
	defer server.Close()

	client := testClient(server.URL)
	_, err := client.CurrentMatches(context.Background(), 0)
	if err == nil {
		t.Fatal("CurrentMatches err = nil, want error")
	}
	if got := err.Error(); got == "" || got == "test-key" {
		t.Fatalf("unexpected error text: %q", got)
	}
}

func testClient(baseURL string) *Client {
	return &Client{
		baseURL:            baseURL,
		apiKey:             "test-key",
		httpClient:         http.DefaultClient,
		cacheTTL:           time.Minute,
		minRequestInterval: 0,
		listCache:          make(map[int]cachedMatches),
		detailCache:        make(map[string]cachedDetails),
	}
}
