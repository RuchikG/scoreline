package cricketdata

import (
	"encoding/json"
	"os"
	"testing"

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
