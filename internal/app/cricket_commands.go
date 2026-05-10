package app

import (
	"context"
	"errors"
	"time"

	"github.com/RuchikG/scoreline/internal/cricketdata"
	"github.com/RuchikG/scoreline/internal/cricsheet"
	"github.com/RuchikG/scoreline/internal/data"
	tea "github.com/charmbracelet/bubbletea"
)

var errCricketArchiveUnavailable = errors.New("cricsheet archive client unavailable")

func fetchCricketLiveMatches(sessionID uint64, client *cricketdata.Client, useMockData bool) tea.Cmd {
	return func() tea.Msg {
		if useMockData {
			return cricketLiveMatchesMsg{sessionID: sessionID, matches: data.MockCricketLiveMatches()}
		}
		if client == nil || !client.HasAPIKey() {
			return cricketLiveMatchesMsg{sessionID: sessionID, matches: nil}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		matches, err := client.CurrentMatches(ctx, 0)
		return cricketLiveMatchesMsg{sessionID: sessionID, matches: matches, err: err}
	}
}

func fetchCricketMatchDetails(sessionID uint64, client *cricketdata.Client, matchID string, useMockData bool) tea.Cmd {
	return func() tea.Msg {
		if useMockData {
			return cricketMatchDetailsMsg{sessionID: sessionID, matchID: matchID, details: data.MockCricketMatchDetails(matchID)}
		}
		if client == nil || !client.HasAPIKey() {
			return cricketMatchDetailsMsg{sessionID: sessionID, matchID: matchID, details: nil}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		details, err := client.MatchInfo(ctx, matchID)
		return cricketMatchDetailsMsg{sessionID: sessionID, matchID: matchID, details: details, err: err}
	}
}

func loadCricketArchiveMatches(sessionID uint64, client *cricsheet.Client, settings data.CricketSettings, useMockData bool) tea.Cmd {
	return func() tea.Msg {
		if useMockData {
			return cricketArchiveMatchesMsg{sessionID: sessionID, matches: data.MockCricketArchiveMatches()}
		}
		if client == nil {
			return cricketArchiveMatchesMsg{sessionID: sessionID, err: errCricketArchiveUnavailable}
		}
		matches, err := client.IndexLocalFiltered(cricsheet.ArchiveFilter{
			Formats:      settings.SelectedFormats,
			Teams:        settings.SelectedTeams,
			Competitions: settings.SelectedCompetitions,
			RecentDays:   settings.ArchiveRecentDays,
			Limit:        200,
		})
		return cricketArchiveMatchesMsg{sessionID: sessionID, matches: matches, err: err}
	}
}

func refreshCricketArchive(sessionID uint64, client *cricsheet.Client) tea.Cmd {
	return func() tea.Msg {
		if client == nil {
			return cricketArchiveRefreshMsg{sessionID: sessionID, err: errCricketArchiveUnavailable}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		return cricketArchiveRefreshMsg{sessionID: sessionID, err: client.RefreshAllJSON(ctx)}
	}
}
