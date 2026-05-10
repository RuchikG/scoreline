package app

import (
	"path/filepath"
	"testing"

	"github.com/RuchikG/scoreline/internal/cricket"
	"github.com/RuchikG/scoreline/internal/data"
	"github.com/RuchikG/scoreline/internal/sports"
)

func TestNewStartsAtSportSelectorWithoutSavedSport(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(t.TempDir(), "config"))

	m := New(true, false, true, false, "dev")
	if m.currentView != viewSportSelector {
		t.Fatalf("currentView = %v, want viewSportSelector", m.currentView)
	}
}

func TestNewStartsAtLastSelectedSport(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(t.TempDir(), "config"))

	settings := data.DefaultSettings()
	settings.SelectedSport = sports.Cricket
	if err := data.SaveSettings(settings); err != nil {
		t.Fatalf("SaveSettings: %v", err)
	}

	m := New(true, false, true, false, "dev")
	if m.currentView != viewCricketMain {
		t.Fatalf("currentView = %v, want viewCricketMain", m.currentView)
	}
}

func TestStaleSessionMessageIgnored(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(t.TempDir(), "config"))

	m := New(true, false, true, false, "dev")
	m.sportSessionID = 2
	m.loading = true

	updated, _ := m.Update(pollDisplayCompleteMsg{sessionID: 1})
	got := updated.(model)
	if !got.loading {
		t.Fatal("stale pollDisplayCompleteMsg should not clear loading")
	}
}

func TestStaleCricketDetailsForPreviousSelectionIgnored(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(t.TempDir(), "config"))

	m := New(true, false, true, false, "dev")
	m.currentView = viewCricketLive
	m.sportSessionID = 2
	m.selected = 1
	m.cricketMatches = []cricket.Match{
		{ID: "old"},
		{ID: "current"},
	}
	m.cricketDetails = &cricket.MatchDetails{Match: cricket.Match{ID: "current"}}

	updated, _ := m.Update(cricketMatchDetailsMsg{
		sessionID: 2,
		matchID:   "old",
		details:   &cricket.MatchDetails{Match: cricket.Match{ID: "old"}},
	})
	got := updated.(model)
	if got.cricketDetails == nil || got.cricketDetails.Match.ID != "current" {
		t.Fatalf("cricketDetails = %#v, want current selection preserved", got.cricketDetails)
	}
}
