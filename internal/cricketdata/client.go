// Package cricketdata integrates with CricketData.org/CricAPI.
package cricketdata

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/RuchikG/scoreline/internal/cricket"
)

const defaultBaseURL = "https://api.cricapi.com/v1"

// Client fetches live cricket data from CricketData.org.
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewClientFromEnv creates a client using an API key environment variable.
func NewClientFromEnv(envName string) *Client {
	if envName == "" {
		envName = "CRICKETDATA_API_KEY"
	}
	return &Client{
		baseURL: defaultBaseURL,
		apiKey:  strings.TrimSpace(os.Getenv(envName)),
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// HasAPIKey reports whether the client can call authenticated endpoints.
func (c *Client) HasAPIKey() bool {
	return c != nil && c.apiKey != ""
}

// CurrentMatches fetches current matches from the free currentMatches endpoint.
func (c *Client) CurrentMatches(ctx context.Context, offset int) ([]cricket.Match, error) {
	if c == nil || c.apiKey == "" {
		return nil, fmt.Errorf("cricketdata api key is not configured")
	}
	endpoint, err := url.Parse(c.baseURL + "/currentMatches")
	if err != nil {
		return nil, err
	}
	q := endpoint.Query()
	q.Set("apikey", c.apiKey)
	q.Set("offset", fmt.Sprintf("%d", offset))
	endpoint.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("cricketdata current matches status: %s", resp.Status)
	}

	var payload currentMatchesResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}
	return mapCurrentMatches(payload.Data), nil
}

type currentMatchesResponse struct {
	Data []currentMatch `json:"data"`
}

type currentMatch struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	MatchType string         `json:"matchType"`
	Status    string         `json:"status"`
	Venue     string         `json:"venue"`
	DateTime  string         `json:"dateTimeGMT"`
	Teams     []string       `json:"teams"`
	Score     []currentScore `json:"score"`
}

type currentScore struct {
	Runs    int     `json:"r"`
	Wickets int     `json:"w"`
	Overs   float64 `json:"o"`
	Inning  string  `json:"inning"`
}

func mapCurrentMatches(rows []currentMatch) []cricket.Match {
	matches := make([]cricket.Match, 0, len(rows))
	for _, row := range rows {
		teams := make([]cricket.Team, 0, len(row.Teams))
		for _, name := range row.Teams {
			teams = append(teams, cricket.Team{
				ID:        strings.ToLower(strings.ReplaceAll(name, " ", "-")),
				Name:      name,
				ShortName: shortTeamName(name),
			})
		}
		startTime, _ := time.Parse(time.RFC3339, row.DateTime)
		matches = append(matches, cricket.Match{
			ID:                  row.ID,
			Source:              "cricketdata",
			Competition:         row.Name,
			MatchType:           strings.ToUpper(row.MatchType),
			Venue:               row.Venue,
			StartTime:           startTime,
			Status:              mapStatus(row.Status),
			Teams:               teams,
			CurrentScoreSummary: scoreSummary(row.Score),
		})
	}
	return matches
}

func mapStatus(status string) cricket.MatchStatus {
	normalized := strings.ToLower(status)
	switch {
	case strings.Contains(normalized, "live"), strings.Contains(normalized, "innings"), strings.Contains(normalized, "need"):
		return cricket.MatchStatusLive
	case strings.Contains(normalized, "won"), strings.Contains(normalized, "draw"), strings.Contains(normalized, "complete"):
		return cricket.MatchStatusFinished
	case strings.Contains(normalized, "abandon"):
		return cricket.MatchStatusAbandoned
	case strings.Contains(normalized, "no result"):
		return cricket.MatchStatusNoResult
	case strings.Contains(normalized, "postpone"):
		return cricket.MatchStatusPostponed
	case strings.Contains(normalized, "cancel"):
		return cricket.MatchStatusCancelled
	default:
		return cricket.MatchStatusNotStarted
	}
}

func scoreSummary(scores []currentScore) string {
	if len(scores) == 0 {
		return "No score yet"
	}
	score := scores[len(scores)-1]
	return fmt.Sprintf("%s %d/%d (%.1f ov)", score.Inning, score.Runs, score.Wickets, score.Overs)
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
