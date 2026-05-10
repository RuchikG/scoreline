package app

import (
	"time"

	"github.com/RuchikG/scoreline/internal/constants"
	tea "github.com/charmbracelet/bubbletea"
)

// mainViewCheckMsg is sent after the check delay completes.
type mainViewCheckMsg struct {
	sessionID uint64
	selection int // 0 for Stats, 1 for Live Matches
}

// performMainViewCheck performs a delay check before navigating.
func performMainViewCheck(sessionID uint64, selection int) tea.Cmd {
	return tea.Tick(constants.MainViewCheckDelay, func(t time.Time) tea.Msg {
		return mainViewCheckMsg{sessionID: sessionID, selection: selection}
	})
}
