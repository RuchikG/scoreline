package app

import (
	"context"
	"fmt"
	"time"

	"github.com/RuchikG/scoreline/internal/api"
	"github.com/RuchikG/scoreline/internal/constants"
	"github.com/RuchikG/scoreline/internal/cricket"
	"github.com/RuchikG/scoreline/internal/cricketdata"
	"github.com/RuchikG/scoreline/internal/data"
	"github.com/RuchikG/scoreline/internal/fotmob"
	"github.com/RuchikG/scoreline/internal/sports"
	"github.com/RuchikG/scoreline/internal/ui"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// handleSportSelectorKeys processes keyboard input for the sport selector.
func (m model) handleSportSelectorKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "j", "down":
		if m.selected < 1 {
			m.selected++
		}
	case "k", "up":
		if m.selected > 0 {
			m.selected--
		}
	case "enter":
		if m.selected == 0 {
			return m.selectSport(sports.Soccer)
		}
		return m.selectSport(sports.Cricket)
	}
	return m, nil
}

func (m model) selectSport(sport sports.Sport) (tea.Model, tea.Cmd) {
	if m.loadCancel != nil {
		m.loadCancel()
	}
	m.sportSessionID++
	m.selectedSport = sport
	m.selected = 0
	m.loading = false
	m.mainViewLoading = false
	m.liveViewLoading = false
	m.statsViewLoading = false
	m.polling = false
	m.matchDetails = nil
	m.liveUpdates = nil
	m.lastEvents = nil
	m.lastHomeScore = 0
	m.lastAwayScore = 0
	m.cricketMatches = nil
	m.cricketDetails = nil
	m.cricketArchiveMatches = nil
	m.cricketArchiveDetails = nil
	m.cricketSettingsState = nil
	_ = data.SaveSelectedSport(sport)

	switch sport {
	case sports.Soccer:
		m.currentView = viewSoccerMain
	case sports.Cricket:
		m.currentView = viewCricketMain
	default:
		m.currentView = viewSportSelector
	}
	return m, nil
}

// handleCricketMainViewKeys processes keyboard input for the cricket main menu.
func (m model) handleCricketMainViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "j", "down":
		if m.selected < 2 {
			m.selected++
		}
	case "k", "up":
		if m.selected > 0 {
			m.selected--
		}
	case "enter":
		switch m.selected {
		case 0:
			m.currentView = viewCricketLive
			m.selected = 0
			m.cricketMatches = nil
			m.cricketDetails = nil
			m.lastError = ""
			m.loading = true
			return m, fetchCricketLiveMatches(m.sportSessionID, m.cricketClient, m.useMockData)
		case 1:
			m.currentView = viewCricketArchives
			m.selected = 0
			m.cricketArchiveMatches = nil
			m.cricketArchiveDetails = nil
			m.lastError = ""
			m.loading = true
			settings, err := data.LoadSettings()
			if err != nil {
				settings = data.DefaultSettings()
			}
			return m, loadCricketArchiveMatches(m.sportSessionID, m.cricsheetClient, settings.Cricket, m.useMockData)
		case 2:
			m.currentView = viewCricketSettings
			m.cricketSettingsState = ui.NewCricketSettingsState()
			m.lastError = ""
		}
		m.selected = 0
	}
	return m, nil
}

func (m model) handleCricketLiveMatches(msg cricketLiveMatchesMsg) (tea.Model, tea.Cmd) {
	m.loading = false
	if msg.err != nil {
		m.lastError = constants.ErrorLoadFailed
		return m, nil
	}
	m.lastError = ""
	m.cricketMatches = msg.matches
	m.selected = 0
	if len(m.cricketMatches) > 0 {
		m.cricketDetails = m.cricketDetailsForSelected()
	}
	return m, nil
}

// handleCricketLiveKeys processes keyboard input for the mock cricket dashboard.
func (m model) handleCricketLiveKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if len(m.cricketMatches) == 0 {
		return m, nil
	}
	switch msg.String() {
	case "j", "down":
		if m.selected < len(m.cricketMatches)-1 {
			m.selected++
			m.cricketDetails = m.cricketDetailsForSelected()
		}
	case "k", "up":
		if m.selected > 0 {
			m.selected--
			m.cricketDetails = m.cricketDetailsForSelected()
		}
	}
	return m, nil
}

func (m model) handleCricketArchiveMatches(msg cricketArchiveMatchesMsg) (tea.Model, tea.Cmd) {
	m.loading = false
	if msg.err != nil {
		m.lastError = "Unable to load Cricsheet archive cache"
		return m, nil
	}
	m.lastError = ""
	m.cricketArchiveMatches = msg.matches
	m.selected = 0
	if len(m.cricketArchiveMatches) > 0 {
		m.cricketArchiveDetails = m.cricketArchiveDetailsForSelected()
	}
	return m, nil
}

func (m model) handleCricketArchiveRefresh(msg cricketArchiveRefreshMsg) (tea.Model, tea.Cmd) {
	m.loading = false
	if msg.err != nil {
		m.lastError = "Unable to refresh Cricsheet archive cache"
		return m, nil
	}
	settings, err := data.LoadSettings()
	if err != nil {
		settings = data.DefaultSettings()
	}
	m.loading = true
	return m, loadCricketArchiveMatches(m.sportSessionID, m.cricsheetClient, settings.Cricket, m.useMockData)
}

// handleCricketArchiveKeys processes keyboard input for the archive browser.
func (m model) handleCricketArchiveKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "j", "down":
		if m.selected < len(m.cricketArchiveMatches)-1 {
			m.selected++
			m.cricketArchiveDetails = m.cricketArchiveDetailsForSelected()
		}
	case "k", "up":
		if m.selected > 0 {
			m.selected--
			m.cricketArchiveDetails = m.cricketArchiveDetailsForSelected()
		}
	case "r":
		if m.useMockData {
			settings, err := data.LoadSettings()
			if err != nil {
				settings = data.DefaultSettings()
			}
			m.loading = true
			return m, loadCricketArchiveMatches(m.sportSessionID, m.cricsheetClient, settings.Cricket, true)
		}
		m.loading = true
		m.lastError = ""
		return m, refreshCricketArchive(m.sportSessionID, m.cricsheetClient)
	}
	return m, nil
}

// handleCricketSettingsKeys processes keyboard input for cricket settings.
func (m model) handleCricketSettingsKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.cricketSettingsState == nil {
		m.cricketSettingsState = ui.NewCricketSettingsState()
	}
	if m.cricketSettingsState.Editing {
		switch msg.String() {
		case "enter":
			m.cricketSettingsState.CommitEdit()
		case "esc":
			m.cricketSettingsState.CancelEdit()
		case "backspace", "ctrl+h":
			m.cricketSettingsState.BackspaceEditText()
		default:
			if len(msg.Runes) > 0 {
				m.cricketSettingsState.InsertEditText(string(msg.Runes))
			}
		}
		return m, nil
	}
	switch msg.String() {
	case "j", "down":
		m.cricketSettingsState.Move(1)
	case "k", "up":
		m.cricketSettingsState.Move(-1)
	case " ":
		m.cricketSettingsState.Toggle()
	case "h", "left":
		m.cricketSettingsState.Adjust(-1)
	case "l", "right":
		m.cricketSettingsState.Adjust(1)
	case "s":
		if err := m.cricketSettingsState.Save(); err != nil {
			m.debugLog(fmt.Sprintf("failed to save cricket settings: %v", err))
			m.lastError = "Unable to save cricket settings"
			return m, nil
		}
		m.cricketClient = cricketdata.NewClientFromEnv(m.cricketSettingsState.Settings.APIKeyEnv)
		m.lastError = ""
		return m, nil
	case "enter":
		if m.cricketSettingsState.Cursor >= 6 {
			m.cricketSettingsState.BeginEdit()
			return m, nil
		}
		if err := m.cricketSettingsState.Save(); err != nil {
			m.debugLog(fmt.Sprintf("failed to save cricket settings: %v", err))
			m.lastError = "Unable to save cricket settings"
			return m, nil
		}
		m.cricketClient = cricketdata.NewClientFromEnv(m.cricketSettingsState.Settings.APIKeyEnv)
		m.lastError = ""
		return m, nil
	}
	return m, nil
}

func (m model) cricketDetailsForSelected() *cricket.MatchDetails {
	if m.selected < 0 || m.selected >= len(m.cricketMatches) {
		return nil
	}
	if m.useMockData {
		return data.MockCricketMatchDetails(m.cricketMatches[m.selected].ID)
	}
	return &cricket.MatchDetails{
		Match:       m.cricketMatches[m.selected],
		LastUpdated: time.Now(),
	}
}

func (m model) cricketArchiveDetailsForSelected() *cricket.MatchDetails {
	if m.selected < 0 || m.selected >= len(m.cricketArchiveMatches) {
		return nil
	}
	if m.useMockData {
		return data.MockCricketArchiveDetails(m.cricketArchiveMatches[m.selected].ID)
	}
	if m.cricsheetClient == nil {
		return &cricket.MatchDetails{Match: m.cricketArchiveMatches[m.selected], LastUpdated: time.Now()}
	}
	details, err := m.cricsheetClient.Details(m.cricketArchiveMatches[m.selected].ID)
	if err != nil {
		return &cricket.MatchDetails{Match: m.cricketArchiveMatches[m.selected], LastUpdated: time.Now()}
	}
	return details
}

// handleMainViewKeys processes keyboard input for the main menu view.
// Handles navigation (up/down) and selection (enter) to switch between views.
// On selection, immediately starts API preloading while showing spinner for 2 seconds.
func (m model) handleMainViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "j", "down":
		if m.selected < 2 && !m.mainViewLoading { // 3 menu items: 0, 1, 2
			m.selected++
		}
	case "k", "up":
		if m.selected > 0 && !m.mainViewLoading {
			m.selected--
		}
	case "enter":
		if m.mainViewLoading {
			return m, nil
		}

		// Handle Settings view separately (no API calls needed)
		if m.selected == 2 {
			m.settingsState = ui.NewSettingsState()
			m.currentView = viewSettings
			return m, nil
		}

		m.mainViewLoading = true
		m.pendingSelection = m.selected

		// Cancel any in-flight requests from previous view
		if m.loadCancel != nil {
			m.loadCancel()
		}
		m.loadCtx, m.loadCancel = context.WithCancel(context.Background())

		// Clear previous view state
		m.matches = nil
		m.upcomingMatches = nil
		m.matchDetails = nil
		m.liveUpdates = nil
		m.lastEvents = nil
		m.lastHomeScore = 0
		m.lastAwayScore = 0
		m.polling = false
		m.upcomingMatchesList.SetItems([]list.Item{})
		m.matchDetailsCache = make(map[int]*api.MatchDetails)

		// Start API calls immediately while showing main view spinner
		cmds := []tea.Cmd{
			m.spinner.Tick,
			performMainViewCheck(m.sportSessionID, m.selected),
		}

		switch m.selected {
		case 0: // Stats view - fetch data progressively (day by day)
			m.statsViewLoading = true
			m.loading = true
			m.statsData = nil                          // Clear cached data to force fresh fetch
			m.statsDaysLoaded = 0                      // Reset progress
			m.statsTotalDays = fotmob.StatsDataDays    // Set total days to load
			m.statsMatchesList.SetItems([]list.Item{}) // Clear list
			cmds = append(cmds, ui.SpinnerTick())
			// Start fetching day 0 (today) first - results shown immediately when it completes
			cmds = append(cmds, fetchStatsDayData(m.sportSessionID, m.loadCtx, m.fotmobClient, m.useMockData, 0, fotmob.StatsDataDays))
		case 1: // Live Matches view - preload live matches progressively (parallel batches)
			m.liveViewLoading = true
			m.loading = true
			m.liveBatchesLoaded = 0
			totalLeagues := fotmob.TotalLeagues()
			m.liveTotalBatches = (totalLeagues + LiveBatchSize - 1) / LiveBatchSize // Ceiling division
			m.liveMatchesBuffer = nil                                               // Clear buffer
			m.liveUpcomingBuffer = nil                                              // Clear upcoming buffer
			m.liveUpcomingMatches = nil                                             // Clear upcoming display
			m.liveMatchesList.SetItems([]list.Item{})
			cmds = append(cmds, ui.SpinnerTick())
			// Start fetching batch 0 (4 leagues in parallel) - results shown when batch completes
			cmds = append(cmds, fetchLiveBatchData(m.sportSessionID, m.loadCtx, m.fotmobClient, m.useMockData, 0))
		}

		return m, tea.Batch(cmds...)
	}
	return m, nil
}

// handleStatsViewKeys processes keyboard input for the stats view.
// Handles date range navigation (left/right) to change the time period.
// Uses client-side filtering from cached data - no new API calls needed!
func (m model) handleStatsViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "l", "right":
		// Cycle date range forward: 1 -> 3 -> 5 -> 1
		switch m.statsDateRange {
		case 1:
			m.statsDateRange = 3
		case 3:
			m.statsDateRange = 5
		default:
			m.statsDateRange = 1
		}
	case "h", "left":
		// Cycle date range backward: 1 -> 5 -> 3 -> 1
		switch m.statsDateRange {
		case 1:
			m.statsDateRange = 5
		case 5:
			m.statsDateRange = 3
		default:
			m.statsDateRange = 1
		}
	case "tab":
		// Tab = toggle focus between left and right panels
		m.statsRightPanelFocused = !m.statsRightPanelFocused
		// Reset scroll position when changing focus (both ways for consistency)
		m.statsScrollOffset = 0
		return m, nil
	default:
		return m, nil
	}

	// If we have cached stats data, just filter client-side (instant!)
	if m.statsData != nil {
		m.matchDetails = nil
		m.matchDetailsCache = make(map[int]*api.MatchDetails)
		m.applyStatsDateFilter()
		m.selected = 0

		// Load details for first match if available
		if len(m.matches) > 0 {
			m.statsMatchesList.Select(0)
			return m.loadStatsMatchDetails(m.matches[0].ID)
		}
		return m, nil
	}

	// No cached data - need to fetch (shouldn't happen normally)
	m.statsViewLoading = true
	m.loading = true
	m.statsDaysLoaded = 0
	m.statsTotalDays = fotmob.StatsDataDays
	if m.loadCancel != nil {
		m.loadCancel()
	}
	m.loadCtx, m.loadCancel = context.WithCancel(context.Background())
	return m, tea.Batch(m.spinner.Tick, ui.SpinnerTick(), fetchStatsDayData(m.sportSessionID, m.loadCtx, m.fotmobClient, m.useMockData, 0, fotmob.StatsDataDays))
}

// loadMatchDetails loads match details for the live matches view.
// Resets live updates and event history before fetching new details.
func (m model) loadMatchDetails(matchID int) (tea.Model, tea.Cmd) {
	return m.loadMatchDetailsWithRefresh(matchID, false)
}

// loadMatchDetailsWithRefresh loads match details for the live matches view with optional cache bypass.
func (m model) loadMatchDetailsWithRefresh(matchID int, forceRefresh bool) (tea.Model, tea.Cmd) {
	chainAlive := m.polling || m.liveViewLoading // check before mutation: if true, tick chain is already running
	m.liveUpdates = nil
	m.lastEvents = nil
	m.lastHomeScore = 0
	m.lastAwayScore = 0
	m.loading = true
	m.liveViewLoading = true
	m.polling = false // Reset polling state - this is a new match load, not a poll refresh
	m.pollGen++       // Invalidate any in-flight poll timers from the previous chain

	var cmd tea.Cmd
	if forceRefresh {
		cmd = fetchMatchDetailsForceRefresh(m.sportSessionID, m.fotmobClient, matchID, m.useMockData)
	} else {
		cmd = fetchMatchDetails(m.sportSessionID, m.fotmobClient, matchID, m.useMockData)
	}

	if chainAlive {
		return m, tea.Batch(m.spinner.Tick, cmd)
	}
	return m, tea.Batch(m.spinner.Tick, ui.SpinnerTick(), cmd)
}

// loadStatsMatchDetails loads match details for the stats view.
// Checks cache first to avoid redundant API calls.
func (m model) loadStatsMatchDetails(matchID int) (tea.Model, tea.Cmd) {
	return m.loadStatsMatchDetailsWithRefresh(matchID, false)
}

// loadStatsMatchDetailsWithRefresh loads match details with optional cache bypass.
func (m model) loadStatsMatchDetailsWithRefresh(matchID int, forceRefresh bool) (tea.Model, tea.Cmd) {
	m.debugLog(fmt.Sprintf("Loading match details for ID: %d (forceRefresh: %v)", matchID, forceRefresh))

	// Check cache unless force refresh is requested
	if !forceRefresh {
		if cached, ok := m.matchDetailsCache[matchID]; ok {
			m.matchDetails = cached
			m.debugLog(fmt.Sprintf("Using cached match details for ID: %d", matchID))
			return m, nil
		}
	} else {
		// Clear from cache to force fresh fetch
		delete(m.matchDetailsCache, matchID)
		m.debugLog(fmt.Sprintf("Cleared cache for match ID: %d", matchID))
	}

	// Fetch from API
	m.loading = true
	m.statsViewLoading = true
	m.debugLog(fmt.Sprintf("Fetching match details from API for ID: %d", matchID))
	return m, tea.Batch(m.spinner.Tick, ui.SpinnerTick(), fetchStatsMatchDetailsFotmob(m.sportSessionID, m.fotmobClient, matchID, m.useMockData))
}

// handleSettingsViewKeys processes keyboard input for the settings view.
// Follows the same pattern as handleStatsSelection for consistent behavior.
func (m model) handleSettingsViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.settingsState == nil {
		return m, nil
	}

	// Check if list is filtering - if so, let list handle ALL keys
	isFiltering := m.settingsState.List.FilterState() == list.Filtering

	// Only handle custom keys when NOT filtering
	if !isFiltering {
		switch msg.String() {
		case " ": // Space to toggle selection
			m.settingsState.Toggle()
			return m, nil
		case "right", "l": // Right arrow or 'l' to next tab
			m.settingsState.NextRegion()
			return m, nil
		case "left", "h": // Left arrow or 'h' to previous tab
			m.settingsState.PreviousRegion()
			return m, nil
		case "enter":
			// Save settings and return to main menu
			if err := m.settingsState.Save(); err != nil {
				m.debugLog(fmt.Sprintf("failed to save settings: %v", err))
			}
			m.settingsState = nil
			m.currentView = viewMain
			m.selected = 0
			return m, nil
		}
	}

	// Delegate to list component for navigation, filtering, etc.
	var listCmd tea.Cmd
	m.settingsState.List, listCmd = m.settingsState.List.Update(msg)
	return m, listCmd
}
