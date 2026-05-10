package data

import (
	"time"

	"github.com/RuchikG/scoreline/internal/cricket"
)

// MockCricketLiveMatches returns representative cricket matches for mock mode.
func MockCricketLiveMatches() []cricket.Match {
	india := cricket.Team{ID: "ind", Name: "India", ShortName: "IND"}
	australia := cricket.Team{ID: "aus", Name: "Australia", ShortName: "AUS"}
	england := cricket.Team{ID: "eng", Name: "England", ShortName: "ENG"}
	southAfrica := cricket.Team{ID: "sa", Name: "South Africa", ShortName: "SA"}

	return []cricket.Match{
		{
			ID:                  "mock-cricket-1",
			Source:              "mock",
			Competition:         "T20 International",
			MatchType:           "T20",
			Venue:               "Wankhede Stadium",
			StartTime:           time.Now().Add(-90 * time.Minute),
			Status:              cricket.MatchStatusLive,
			Teams:               []cricket.Team{india, australia},
			CurrentScoreSummary: "IND 148/3 (15.2 ov)",
		},
		{
			ID:                  "mock-cricket-2",
			Source:              "mock",
			Competition:         "ODI Series",
			MatchType:           "ODI",
			Venue:               "Lord's",
			StartTime:           time.Now().Add(2 * time.Hour),
			Status:              cricket.MatchStatusNotStarted,
			Teams:               []cricket.Team{england, southAfrica},
			CurrentScoreSummary: "Starts soon",
		},
	}
}

// MockCricketArchiveMatches returns representative completed matches for archive browsing.
func MockCricketArchiveMatches() []cricket.Match {
	india := cricket.Team{ID: "ind", Name: "India", ShortName: "IND"}
	australia := cricket.Team{ID: "aus", Name: "Australia", ShortName: "AUS"}
	england := cricket.Team{ID: "eng", Name: "England", ShortName: "ENG"}
	southAfrica := cricket.Team{ID: "sa", Name: "South Africa", ShortName: "SA"}

	return []cricket.Match{
		{
			ID:                  "mock-archive-1",
			Source:              "mock",
			Competition:         "World Cup",
			MatchType:           "ODI",
			Venue:               "Narendra Modi Stadium",
			StartTime:           time.Now().AddDate(0, 0, -3),
			Status:              cricket.MatchStatusFinished,
			Teams:               []cricket.Team{india, australia},
			CurrentScoreSummary: "Australia won by 6 wickets",
		},
		{
			ID:                  "mock-archive-2",
			Source:              "mock",
			Competition:         "Test Series",
			MatchType:           "TEST",
			Venue:               "The Oval",
			StartTime:           time.Now().AddDate(0, 0, -11),
			Status:              cricket.MatchStatusFinished,
			Teams:               []cricket.Team{england, southAfrica},
			CurrentScoreSummary: "England won by 84 runs",
		},
	}
}

// MockCricketMatchDetails returns mock scorecard details for a cricket match.
func MockCricketMatchDetails(matchID string) *cricket.MatchDetails {
	matches := MockCricketLiveMatches()
	match := matches[0]
	for _, candidate := range matches {
		if candidate.ID == matchID {
			match = candidate
			break
		}
	}

	if match.Status == cricket.MatchStatusNotStarted {
		return &cricket.MatchDetails{
			Match:       match,
			Toss:        "Toss pending",
			Result:      "",
			LastUpdated: time.Now(),
		}
	}

	return &cricket.MatchDetails{
		Match: match,
		Toss:  "Australia won the toss and chose to bowl",
		Innings: []cricket.Innings{
			{
				BattingTeam: match.Teams[0],
				BowlingTeam: match.Teams[1],
				Runs:        148,
				Wickets:     3,
				Overs:       "15.2",
				Target:      0,
				BattingCard: []cricket.PlayerBatting{
					{Player: "R. Sharma", Runs: 54, Balls: 31, Fours: 6, Sixes: 2, StrikeRate: "174.19", Dismissal: "c Maxwell b Starc"},
					{Player: "V. Kohli", Runs: 38, Balls: 29, Fours: 4, Sixes: 1, StrikeRate: "131.03", Dismissal: "not out"},
				},
				BowlingCard: []cricket.PlayerBowling{
					{Player: "M. Starc", Overs: "3.2", Maidens: 0, Runs: 31, Wickets: 1, Economy: "9.30"},
					{Player: "P. Cummins", Overs: "3.0", Maidens: 0, Runs: 28, Wickets: 1, Economy: "9.33"},
				},
				FallOfWickets: []string{"41/1 (4.5)", "92/2 (10.1)", "121/3 (13.4)"},
				RecentOvers:   []string{"12: 1 4 0 1 6 1", "13: 0 W 1 1 4 2", "14: 1 1 0 6 1 1", "15: 4 0"},
			},
		},
		CurrentBatters: []cricket.PlayerBatting{
			{Player: "V. Kohli", Runs: 38, Balls: 29, Fours: 4, Sixes: 1, StrikeRate: "131.03", Dismissal: "not out"},
			{Player: "H. Pandya", Runs: 12, Balls: 7, Fours: 1, Sixes: 1, StrikeRate: "171.43", Dismissal: "not out"},
		},
		CurrentBowler: &cricket.PlayerBowling{Player: "M. Starc", Overs: "3.2", Maidens: 0, Runs: 31, Wickets: 1, Economy: "9.30"},
		RecentOvers:   []string{"12: 1 4 0 1 6 1", "13: 0 W 1 1 4 2", "14: 1 1 0 6 1 1", "15: 4 0"},
		LastUpdated:   time.Now(),
	}
}

// MockCricketArchiveDetails returns a completed-match detail view for archive browsing.
func MockCricketArchiveDetails(matchID string) *cricket.MatchDetails {
	matches := MockCricketArchiveMatches()
	match := matches[0]
	for _, candidate := range matches {
		if candidate.ID == matchID {
			match = candidate
			break
		}
	}

	return &cricket.MatchDetails{
		Match:       match,
		Result:      match.CurrentScoreSummary,
		LastUpdated: time.Now(),
		Innings: []cricket.Innings{
			{
				BattingTeam: match.Teams[0],
				BowlingTeam: match.Teams[1],
				Runs:        241,
				Wickets:     8,
				Overs:       "50.0",
				BattingCard: []cricket.PlayerBatting{
					{Player: match.Teams[0].ShortName + " opener", Runs: 73, Balls: 91, Fours: 8, Sixes: 0, StrikeRate: "80.21", Dismissal: "c keeper b seamer"},
					{Player: match.Teams[0].ShortName + " captain", Runs: 48, Balls: 55, Fours: 4, Sixes: 1, StrikeRate: "87.27", Dismissal: "lbw b spinner"},
				},
				BowlingCard: []cricket.PlayerBowling{
					{Player: match.Teams[1].ShortName + " strike bowler", Overs: "10.0", Maidens: 1, Runs: 42, Wickets: 3, Economy: "4.20"},
					{Player: match.Teams[1].ShortName + " spinner", Overs: "10.0", Maidens: 0, Runs: 39, Wickets: 2, Economy: "3.90"},
				},
			},
			{
				BattingTeam: match.Teams[1],
				BowlingTeam: match.Teams[0],
				Runs:        242,
				Wickets:     4,
				Overs:       "43.0",
				Target:      242,
				BattingCard: []cricket.PlayerBatting{
					{Player: match.Teams[1].ShortName + " anchor", Runs: 89, Balls: 101, Fours: 9, Sixes: 1, StrikeRate: "88.11", Dismissal: "not out"},
					{Player: match.Teams[1].ShortName + " finisher", Runs: 36, Balls: 24, Fours: 3, Sixes: 2, StrikeRate: "150.00", Dismissal: "not out"},
				},
			},
		},
	}
}
