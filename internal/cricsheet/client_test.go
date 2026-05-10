package cricsheet

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/RuchikG/scoreline/internal/cricket"
)

func TestParseCricsheetFile(t *testing.T) {
	match, err := parseCricsheetFile(filepath.Join("testdata", "sample_match.json"))
	if err != nil {
		t.Fatalf("parseCricsheetFile: %v", err)
	}
	if match.Source != "cricsheet" {
		t.Fatalf("Source = %q", match.Source)
	}
	if match.Status != cricket.MatchStatusFinished {
		t.Fatalf("Status = %q", match.Status)
	}
	if match.CurrentScoreSummary != "India won" {
		t.Fatalf("CurrentScoreSummary = %q", match.CurrentScoreSummary)
	}
	if len(match.Teams) != 2 {
		t.Fatalf("len(Teams) = %d, want 2", len(match.Teams))
	}
}

func TestIndexLocalFiltered(t *testing.T) {
	dir := t.TempDir()
	raw, err := os.ReadFile(filepath.Join("testdata", "sample_match.json"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "sample_match.json"), raw, 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	client := &Client{cacheDir: dir}
	matches, err := client.IndexLocalFiltered(ArchiveFilter{
		Formats: []string{"t20"},
		Teams:   []string{"India"},
		Limit:   1,
	})
	if err != nil {
		t.Fatalf("IndexLocalFiltered: %v", err)
	}
	if len(matches) != 1 {
		t.Fatalf("len(matches) = %d, want 1", len(matches))
	}

	matches, err = client.IndexLocalFiltered(ArchiveFilter{Formats: []string{"test"}})
	if err != nil {
		t.Fatalf("IndexLocalFiltered: %v", err)
	}
	if len(matches) != 0 {
		t.Fatalf("len(matches) = %d, want 0", len(matches))
	}
}

func TestDetails(t *testing.T) {
	dir := t.TempDir()
	raw, err := os.ReadFile(filepath.Join("testdata", "sample_match.json"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "sample_match.json"), raw, 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	client := &Client{cacheDir: dir}
	details, err := client.Details("sample_match")
	if err != nil {
		t.Fatalf("Details: %v", err)
	}
	if details.Result != "India won" {
		t.Fatalf("Result = %q, want India won", details.Result)
	}
}
