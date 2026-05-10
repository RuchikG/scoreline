// Package cricket contains cricket-specific domain types.
package cricket

import "time"

// MatchStatus describes the current state of a cricket match.
type MatchStatus string

const (
	MatchStatusNotStarted MatchStatus = "not_started"
	MatchStatusLive       MatchStatus = "live"
	MatchStatusFinished   MatchStatus = "finished"
	MatchStatusAbandoned  MatchStatus = "abandoned"
	MatchStatusNoResult   MatchStatus = "no_result"
	MatchStatusPostponed  MatchStatus = "postponed"
	MatchStatusCancelled  MatchStatus = "cancelled"
)

// Team identifies a cricket team.
type Team struct {
	ID        string
	Name      string
	ShortName string
}

// Score contains one innings score line.
type Score struct {
	Runs          int
	Wickets       int
	Overs         string
	InningsNumber int
	BattingTeam   Team
}

// Match is the list-level cricket match model.
type Match struct {
	ID                  string
	Source              string
	Competition         string
	MatchType           string
	Venue               string
	StartTime           time.Time
	Status              MatchStatus
	Teams               []Team
	CurrentScoreSummary string
}

// PlayerBatting contains scorecard batting details.
type PlayerBatting struct {
	Player     string
	Runs       int
	Balls      int
	Fours      int
	Sixes      int
	StrikeRate string
	Dismissal  string
}

// PlayerBowling contains scorecard bowling details.
type PlayerBowling struct {
	Player  string
	Overs   string
	Maidens int
	Runs    int
	Wickets int
	Economy  string
}

// Innings contains a full innings scorecard section.
type Innings struct {
	BattingTeam    Team
	BowlingTeam    Team
	Runs           int
	Wickets        int
	Overs          string
	Target         int
	BattingCard    []PlayerBatting
	BowlingCard    []PlayerBowling
	FallOfWickets  []string
	RecentOvers    []string
}

// MatchDetails contains detail-panel data for a cricket match.
type MatchDetails struct {
	Match          Match
	Toss           string
	Result         string
	Innings        []Innings
	CurrentBatters []PlayerBatting
	CurrentBowler  *PlayerBowling
	RecentOvers    []string
	LastUpdated     time.Time
}
