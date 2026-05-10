// Package cricsheet handles Cricsheet archive data.
package cricsheet

import (
	"archive/zip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/RuchikG/scoreline/internal/cricket"
	"github.com/RuchikG/scoreline/internal/data"
)

const cacheSubdir = "cricsheet"
const allJSONDownloadURL = "https://cricsheet.org/downloads/all_json.zip"

// ArchiveFilter filters locally indexed Cricsheet matches.
type ArchiveFilter struct {
	Formats      []string
	Teams        []string
	Competitions []string
	RecentDays   int
	Limit        int
}

// Client reads and indexes local Cricsheet JSON files.
type Client struct {
	cacheDir string
}

// NewClient creates a Cricsheet archive client using Scoreline's cache path.
func NewClient() (*Client, error) {
	dir, err := data.EnsureCacheSubdir(cacheSubdir)
	if err != nil {
		return nil, err
	}
	return &Client{cacheDir: dir}, nil
}

// CacheDir returns the local Cricsheet cache directory.
func (c *Client) CacheDir() string {
	if c == nil {
		return ""
	}
	return c.cacheDir
}

// IndexLocal scans cached JSON files and returns lightweight match entries.
func (c *Client) IndexLocal() ([]cricket.Match, error) {
	return c.IndexLocalFiltered(ArchiveFilter{})
}

// IndexLocalFiltered scans cached JSON files and returns filtered match entries.
func (c *Client) IndexLocalFiltered(filter ArchiveFilter) ([]cricket.Match, error) {
	if c == nil || c.cacheDir == "" {
		return nil, fmt.Errorf("cricsheet client is not configured")
	}
	entries, err := os.ReadDir(c.cacheDir)
	if err != nil {
		return nil, err
	}

	matches := make([]cricket.Match, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		match, err := parseCricsheetFile(filepath.Join(c.cacheDir, entry.Name()))
		if err == nil {
			if matchPassesFilter(match, filter) {
				matches = append(matches, match)
			}
		}
	}
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].StartTime.After(matches[j].StartTime)
	})
	if filter.Limit > 0 && len(matches) > filter.Limit {
		matches = matches[:filter.Limit]
	}
	return matches, nil
}

// Details reads a cached Cricsheet JSON match and maps it into a detail model.
func (c *Client) Details(matchID string) (*cricket.MatchDetails, error) {
	if c == nil || c.cacheDir == "" {
		return nil, fmt.Errorf("cricsheet client is not configured")
	}
	match, err := parseCricsheetFile(filepath.Join(c.cacheDir, matchID+".json"))
	if err != nil {
		return nil, err
	}
	return &cricket.MatchDetails{
		Match:       match,
		Result:      match.CurrentScoreSummary,
		LastUpdated: time.Now(),
	}, nil
}

// RefreshAllJSON downloads and extracts the official Cricsheet JSON archive.
func (c *Client) RefreshAllJSON(ctx context.Context) error {
	if c == nil || c.cacheDir == "" {
		return fmt.Errorf("cricsheet client is not configured")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, allJSONDownloadURL, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("cricsheet archive returned %s", resp.Status)
	}

	zipPath := filepath.Join(c.cacheDir, "all_json.zip")
	tmpPath := zipPath + ".tmp"
	out, err := os.Create(tmpPath)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, resp.Body); err != nil {
		_ = out.Close()
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpPath, zipPath); err != nil {
		return err
	}
	return extractJSONZip(zipPath, c.cacheDir)
}

type cricsheetMatch struct {
	Meta struct {
		DataVersion string `json:"data_version"`
	} `json:"meta"`
	Info struct {
		Dates     []string `json:"dates"`
		Event     struct {
			Name string `json:"name"`
		} `json:"event"`
		MatchType string   `json:"match_type"`
		Teams     []string `json:"teams"`
		Venue     string   `json:"venue"`
		Outcome   struct {
			Winner string `json:"winner"`
		} `json:"outcome"`
	} `json:"info"`
}

func extractJSONZip(zipPath string, destDir string) error {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer reader.Close()

	for _, file := range reader.File {
		if file.FileInfo().IsDir() || !strings.HasSuffix(file.Name, ".json") {
			continue
		}
		base := filepath.Base(file.Name)
		if base == "" || base == "." {
			continue
		}
		if err := extractZipFile(file, filepath.Join(destDir, base)); err != nil {
			return err
		}
	}
	return nil
}

func extractZipFile(file *zip.File, destPath string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	tmpPath := destPath + ".tmp"
	dst, err := os.Create(tmpPath)
	if err != nil {
		return err
	}
	if _, err := io.Copy(dst, src); err != nil {
		_ = dst.Close()
		return err
	}
	if err := dst.Close(); err != nil {
		return err
	}
	return os.Rename(tmpPath, destPath)
}

func matchPassesFilter(match cricket.Match, filter ArchiveFilter) bool {
	if filter.RecentDays > 0 && !match.StartTime.IsZero() {
		cutoff := time.Now().AddDate(0, 0, -filter.RecentDays)
		if match.StartTime.Before(cutoff) {
			return false
		}
	}
	if len(filter.Formats) > 0 && !containsFold(filter.Formats, match.MatchType) {
		return false
	}
	if len(filter.Competitions) > 0 && !containsFold(filter.Competitions, match.Competition) {
		return false
	}
	if len(filter.Teams) > 0 {
		for _, team := range match.Teams {
			if containsFold(filter.Teams, team.Name) || containsFold(filter.Teams, team.ShortName) {
				return true
			}
		}
		return false
	}
	return true
}

func containsFold(values []string, needle string) bool {
	for _, value := range values {
		if strings.EqualFold(strings.TrimSpace(value), strings.TrimSpace(needle)) {
			return true
		}
	}
	return false
}

func parseCricsheetFile(path string) (cricket.Match, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return cricket.Match{}, err
	}
	var payload cricsheetMatch
	if err := json.Unmarshal(raw, &payload); err != nil {
		return cricket.Match{}, err
	}

	teams := make([]cricket.Team, 0, len(payload.Info.Teams))
	for _, name := range payload.Info.Teams {
		teams = append(teams, cricket.Team{ID: strings.ToLower(strings.ReplaceAll(name, " ", "-")), Name: name, ShortName: shortTeamName(name)})
	}

	var start time.Time
	if len(payload.Info.Dates) > 0 {
		start, _ = time.Parse("2006-01-02", payload.Info.Dates[0])
	}

	summary := "Completed"
	if payload.Info.Outcome.Winner != "" {
		summary = payload.Info.Outcome.Winner + " won"
	}

	return cricket.Match{
		ID:                  strings.TrimSuffix(filepath.Base(path), ".json"),
		Source:              "cricsheet",
		Competition:         payload.Info.Event.Name,
		MatchType:           strings.ToUpper(payload.Info.MatchType),
		Venue:               payload.Info.Venue,
		StartTime:           start,
		Status:              cricket.MatchStatusFinished,
		Teams:               teams,
		CurrentScoreSummary: summary,
	}, nil
}

func shortTeamName(name string) string {
	parts := strings.Fields(name)
	if len(parts) == 0 {
		return "TBD"
	}
	if len(parts) == 1 {
		if len(parts[0]) <= 3 {
			return strings.ToUpper(parts[0])
		}
		return strings.ToUpper(parts[0][:3])
	}
	var b strings.Builder
	for _, part := range parts {
		if part != "" {
			b.WriteByte(part[0])
		}
	}
	return strings.ToUpper(b.String())
}
