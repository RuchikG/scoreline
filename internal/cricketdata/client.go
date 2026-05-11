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
	"sync"
	"time"

	"github.com/RuchikG/scoreline/internal/cricket"
)

const defaultBaseURL = "https://api.cricapi.com/v1"
const defaultCacheTTL = 30 * time.Second
const defaultMinRequestInterval = 2 * time.Second

// Client fetches live cricket data from CricketData.org.
type Client struct {
	baseURL            string
	apiKey             string
	httpClient         *http.Client
	cacheTTL           time.Duration
	minRequestInterval time.Duration

	mu          sync.Mutex
	lastRequest time.Time
	listCache   map[int]cachedMatches
	detailCache map[string]cachedDetails
}

type cachedMatches struct {
	matches   []cricket.Match
	fetchedAt time.Time
}

type cachedDetails struct {
	details   *cricket.MatchDetails
	fetchedAt time.Time
}

// NewClientFromEnv creates a client using an API key environment variable or a
// direct API key. Environment-style values are resolved from the process env;
// other values are treated as the key itself so pasting a key in settings works.
func NewClientFromEnv(envName string) *Client {
	apiKeySetting := strings.TrimSpace(envName)
	if apiKeySetting == "" {
		apiKeySetting = "CRICKETDATA_API_KEY"
	}
	apiKey := strings.TrimSpace(os.Getenv(apiKeySetting))
	if apiKey == "" && !looksLikeEnvName(apiKeySetting) {
		apiKey = apiKeySetting
	}
	return &Client{
		baseURL:            defaultBaseURL,
		apiKey:             apiKey,
		cacheTTL:           defaultCacheTTL,
		minRequestInterval: defaultMinRequestInterval,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
		listCache:   make(map[int]cachedMatches),
		detailCache: make(map[string]cachedDetails),
	}
}

// HasAPIKey reports whether the client can call authenticated endpoints.
func (c *Client) HasAPIKey() bool {
	return c != nil && c.apiKey != ""
}

func looksLikeEnvName(value string) bool {
	if value == "" {
		return false
	}
	hasUnderscore := false
	for i, r := range value {
		switch {
		case r == '_':
			hasUnderscore = true
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9' && i > 0:
		default:
			return false
		}
	}
	return hasUnderscore
}

// CurrentMatches fetches current matches from the free currentMatches endpoint.
func (c *Client) CurrentMatches(ctx context.Context, offset int) ([]cricket.Match, error) {
	if c == nil || c.apiKey == "" {
		return nil, fmt.Errorf("cricketdata api key is not configured")
	}
	if matches, ok := c.cachedCurrentMatches(offset); ok {
		return matches, nil
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
	resp, err := c.do(req)
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
	if err := payload.err(); err != nil {
		return nil, err
	}
	matches := mapCurrentMatches(payload.Data)
	c.setCurrentMatchesCache(offset, matches)
	return matches, nil
}

type currentMatchesResponse struct {
	Status string         `json:"status"`
	Info   string         `json:"info"`
	Reason string         `json:"reason"`
	Data   []currentMatch `json:"data"`
}

func (r currentMatchesResponse) err() error {
	if strings.EqualFold(r.Status, "failure") {
		reason := strings.TrimSpace(r.Reason)
		if reason == "" {
			reason = strings.TrimSpace(r.Info)
		}
		if reason == "" {
			reason = "request failed"
		}
		return fmt.Errorf("cricketdata current matches: %s", reason)
	}
	return nil
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

// MatchInfo fetches details for one match from CricketData's match_info endpoint.
func (c *Client) MatchInfo(ctx context.Context, id string) (*cricket.MatchDetails, error) {
	if c == nil || c.apiKey == "" {
		return nil, fmt.Errorf("cricketdata api key is not configured")
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, fmt.Errorf("cricketdata match id is required")
	}
	if details, ok := c.cachedMatchInfo(id); ok {
		return details, nil
	}

	endpoint, err := url.Parse(c.baseURL + "/match_info")
	if err != nil {
		return nil, err
	}
	q := endpoint.Query()
	q.Set("apikey", c.apiKey)
	q.Set("id", id)
	endpoint.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("cricketdata match info status: %s", resp.Status)
	}

	var payload matchInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}
	if err := payload.err(); err != nil {
		return nil, err
	}
	details := mapMatchInfo(payload.Data)
	c.setMatchInfoCache(id, details)
	return details, nil
}

type matchInfoResponse struct {
	Status string    `json:"status"`
	Info   string    `json:"info"`
	Reason string    `json:"reason"`
	Data   matchInfo `json:"data"`
}

func (r matchInfoResponse) err() error {
	if strings.EqualFold(r.Status, "failure") {
		reason := strings.TrimSpace(r.Reason)
		if reason == "" {
			reason = strings.TrimSpace(r.Info)
		}
		if reason == "" {
			reason = "request failed"
		}
		return fmt.Errorf("cricketdata match info: %s", reason)
	}
	return nil
}

type matchInfo struct {
	currentMatch
	TossWinner string `json:"tossWinner"`
	TossChoice string `json:"tossChoice"`
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
		matches = append(matches, mapMatch(row))
	}
	return matches
}

func mapMatch(row currentMatch) cricket.Match {
	teams := make([]cricket.Team, 0, len(row.Teams))
	for _, name := range row.Teams {
		teams = append(teams, cricket.Team{
			ID:        strings.ToLower(strings.ReplaceAll(name, " ", "-")),
			Name:      name,
			ShortName: shortTeamName(name),
		})
	}
	startTime := parseCricketTime(row.DateTime)
	return cricket.Match{
		ID:                  row.ID,
		Source:              "cricketdata",
		Competition:         row.Name,
		MatchType:           strings.ToUpper(row.MatchType),
		Venue:               row.Venue,
		StartTime:           startTime,
		Status:              mapStatus(row.Status),
		Teams:               teams,
		CurrentScoreSummary: scoreSummary(row.Score),
	}
}

func parseCricketTime(value string) time.Time {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}
	}
	for _, layout := range []string{time.RFC3339, "2006-01-02T15:04:05"} {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed
		}
	}
	return time.Time{}
}

func mapMatchInfo(row matchInfo) *cricket.MatchDetails {
	toss := ""
	if row.TossWinner != "" {
		toss = row.TossWinner
		if row.TossChoice != "" {
			toss += " chose to " + row.TossChoice
		}
	}
	return &cricket.MatchDetails{
		Match:       mapMatch(row.currentMatch),
		Toss:        toss,
		Result:      resultText(row.Status),
		LastUpdated: time.Now(),
	}
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

func resultText(status string) string {
	if mapStatus(status) == cricket.MatchStatusFinished {
		return status
	}
	return ""
}

func (c *Client) do(req *http.Request) (*http.Response, error) {
	if c.minRequestInterval > 0 {
		c.mu.Lock()
		wait := c.minRequestInterval - time.Since(c.lastRequest)
		c.mu.Unlock()
		if wait > 0 {
			timer := time.NewTimer(wait)
			select {
			case <-req.Context().Done():
				timer.Stop()
				return nil, req.Context().Err()
			case <-timer.C:
			}
		}
	}
	resp, err := c.httpClient.Do(req)
	if err == nil {
		c.mu.Lock()
		c.lastRequest = time.Now()
		c.mu.Unlock()
	}
	return resp, err
}

func (c *Client) cachedCurrentMatches(offset int) ([]cricket.Match, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, ok := c.listCache[offset]
	if !ok || c.cacheTTL <= 0 || time.Since(entry.fetchedAt) > c.cacheTTL {
		return nil, false
	}
	return append([]cricket.Match(nil), entry.matches...), true
}

func (c *Client) setCurrentMatchesCache(offset int, matches []cricket.Match) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.listCache == nil {
		c.listCache = make(map[int]cachedMatches)
	}
	c.listCache[offset] = cachedMatches{matches: append([]cricket.Match(nil), matches...), fetchedAt: time.Now()}
}

func (c *Client) cachedMatchInfo(id string) (*cricket.MatchDetails, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, ok := c.detailCache[id]
	if !ok || c.cacheTTL <= 0 || time.Since(entry.fetchedAt) > c.cacheTTL {
		return nil, false
	}
	return cloneDetails(entry.details), true
}

func (c *Client) setMatchInfoCache(id string, details *cricket.MatchDetails) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.detailCache == nil {
		c.detailCache = make(map[string]cachedDetails)
	}
	c.detailCache[id] = cachedDetails{details: cloneDetails(details), fetchedAt: time.Now()}
}

func cloneDetails(details *cricket.MatchDetails) *cricket.MatchDetails {
	if details == nil {
		return nil
	}
	next := *details
	next.Match.Teams = append([]cricket.Team(nil), details.Match.Teams...)
	next.Innings = append([]cricket.Innings(nil), details.Innings...)
	next.CurrentBatters = append([]cricket.PlayerBatting(nil), details.CurrentBatters...)
	next.RecentOvers = append([]string(nil), details.RecentOvers...)
	if details.CurrentBowler != nil {
		bowler := *details.CurrentBowler
		next.CurrentBowler = &bowler
	}
	return &next
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
