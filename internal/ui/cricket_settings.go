package ui

import (
	"fmt"
	"slices"
	"strings"

	"github.com/RuchikG/scoreline/internal/constants"
	"github.com/RuchikG/scoreline/internal/data"
	"github.com/RuchikG/scoreline/internal/ui/design"
	"github.com/charmbracelet/lipgloss"
)

var cricketFormatOptions = []string{"test", "odi", "t20"}

// CricketSettingsState holds editable cricket-specific settings.
type CricketSettingsState struct {
	Settings data.CricketSettings
	Cursor   int
	Saved    bool
	Editing  bool
	EditRow  int
	EditText string
}

// NewCricketSettingsState loads cricket settings for the settings view.
func NewCricketSettingsState() *CricketSettingsState {
	settings, err := data.LoadSettings()
	if err != nil {
		settings = data.DefaultSettings()
	}
	return &CricketSettingsState{Settings: settings.Cricket}
}

// Move moves the selected settings row.
func (s *CricketSettingsState) Move(delta int) {
	if s == nil {
		return
	}
	s.Cursor += delta
	if s.Cursor < 0 {
		s.Cursor = 0
	}
	if s.Cursor > s.maxRow() {
		s.Cursor = s.maxRow()
	}
}

// Toggle toggles a format row.
func (s *CricketSettingsState) Toggle() {
	if s == nil || s.Cursor >= len(cricketFormatOptions) || s.Editing {
		return
	}
	format := cricketFormatOptions[s.Cursor]
	if slices.Contains(s.Settings.SelectedFormats, format) {
		s.Settings.SelectedFormats = removeString(s.Settings.SelectedFormats, format)
		s.Saved = false
		return
	}
	s.Settings.SelectedFormats = append(s.Settings.SelectedFormats, format)
	s.Saved = false
}

// Adjust changes numeric settings on the selected row.
func (s *CricketSettingsState) Adjust(delta int) {
	if s == nil || s.Editing {
		return
	}
	switch s.Cursor {
	case 3:
		s.Settings.LiveRefreshSeconds = clamp(s.Settings.LiveRefreshSeconds+delta*30, 60, 1800)
	case 4:
		s.Settings.DetailRefreshSeconds = clamp(s.Settings.DetailRefreshSeconds+delta*15, 30, 600)
	case 5:
		s.Settings.ArchiveRecentDays = clamp(s.Settings.ArchiveRecentDays+delta*7, 7, 365)
	}
	s.Saved = false
}

// BeginEdit starts text editing for list-style settings.
func (s *CricketSettingsState) BeginEdit() {
	if s == nil {
		return
	}
	switch s.Cursor {
	case 6:
		s.startEdit(strings.Join(s.Settings.SelectedTeams, ", "))
	case 7:
		s.startEdit(strings.Join(s.Settings.SelectedCompetitions, ", "))
	case 8:
		s.startEdit(s.Settings.APIKeyEnv)
	}
}

// InsertEditText appends typed text to the active editor.
func (s *CricketSettingsState) InsertEditText(text string) {
	if s == nil || !s.Editing {
		return
	}
	s.EditText += text
}

// BackspaceEditText removes the last byte from the active editor.
func (s *CricketSettingsState) BackspaceEditText() {
	if s == nil || !s.Editing || s.EditText == "" {
		return
	}
	s.EditText = s.EditText[:len(s.EditText)-1]
}

// CommitEdit saves the active text editor into settings.
func (s *CricketSettingsState) CommitEdit() {
	if s == nil || !s.Editing {
		return
	}
	switch s.EditRow {
	case 6:
		s.Settings.SelectedTeams = splitCSV(s.EditText)
	case 7:
		s.Settings.SelectedCompetitions = splitCSV(s.EditText)
	case 8:
		s.Settings.APIKeyEnv = strings.TrimSpace(s.EditText)
	}
	s.Editing = false
	s.Saved = false
}

// CancelEdit exits the active text editor without applying changes.
func (s *CricketSettingsState) CancelEdit() {
	if s == nil {
		return
	}
	s.Editing = false
}

func (s *CricketSettingsState) startEdit(text string) {
	s.Editing = true
	s.EditRow = s.Cursor
	s.EditText = text
}

// Save persists cricket-specific settings without changing soccer settings.
func (s *CricketSettingsState) Save() error {
	if s == nil {
		return nil
	}
	settings, err := data.LoadSettings()
	if err != nil {
		settings = data.DefaultSettings()
	}
	settings.Cricket = s.Settings
	if len(settings.Cricket.SelectedFormats) == 0 {
		settings.Cricket.SelectedFormats = []string{"test", "odi", "t20"}
	}
	if err := data.SaveSettings(settings); err != nil {
		return err
	}
	s.Settings = settings.Cricket
	s.Saved = true
	return nil
}

func (s *CricketSettingsState) maxRow() int {
	return len(cricketFormatOptions) + 5
}

// RenderCricketSettingsView renders cricket-specific settings.
func RenderCricketSettingsView(width, height int, state *CricketSettingsState, apiKeyConfigured bool, bannerType constants.StatusBannerType) string {
	const contentWidth = 76

	if state == nil {
		state = NewCricketSettingsState()
	}

	statusBanner := renderStatusBanner(bannerType, contentWidth)
	if statusBanner != "" {
		statusBanner += "\n"
	}
	title := design.RenderHeader("Cricket Settings", contentWidth)
	rows := renderCricketSettingsRows(state, apiKeyConfigured, contentWidth)
	helpText := "j/k: navigate  Space: toggle  h/l: adjust  Enter: edit/save  s: save  Esc: back"
	if state.Editing {
		helpText = "type: edit  Backspace: delete  Enter: apply  Esc: cancel"
	}
	help := menuHelpStyle.Render(helpText)

	content := lipgloss.JoinVertical(lipgloss.Left, statusBanner, title, "", rows, "", help)
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content)
}

func renderCricketSettingsRows(state *CricketSettingsState, apiKeyConfigured bool, width int) string {
	rows := []string{
		neonDimStyle.Width(width).Render("Formats"),
	}
	for i, format := range cricketFormatOptions {
		checked := " "
		if slices.Contains(state.Settings.SelectedFormats, format) {
			checked = "x"
		}
		rows = append(rows, settingsRow(i == state.Cursor, fmt.Sprintf("[%s] %s", checked, strings.ToUpper(format)), width))
	}

	rows = append(rows, "", neonDimStyle.Width(width).Render("Refresh"))
	rows = append(rows, settingsRow(state.Cursor == 3, fmt.Sprintf("Live list refresh: %ds", state.Settings.LiveRefreshSeconds), width))
	rows = append(rows, settingsRow(state.Cursor == 4, fmt.Sprintf("Selected match refresh: %ds", state.Settings.DetailRefreshSeconds), width))
	rows = append(rows, settingsRow(state.Cursor == 5, fmt.Sprintf("Archive window: %d days", state.Settings.ArchiveRecentDays), width))

	rows = append(rows, "", neonDimStyle.Width(width).Render("Filters"))
	rows = append(rows, settingsRow(state.Cursor == 6, "Teams: "+editableValue(state, 6, state.Settings.SelectedTeams), width))
	rows = append(rows, settingsRow(state.Cursor == 7, "Competitions: "+editableValue(state, 7, state.Settings.SelectedCompetitions), width))

	apiStatus := "missing"
	if apiKeyConfigured {
		apiStatus = "configured"
	}
	rows = append(rows, "", neonDimStyle.Width(width).Render("CricketData"))
	apiEnv := displayAPIKeySetting(state.Settings.APIKeyEnv)
	if state.Editing && state.EditRow == 8 {
		apiEnv = "> " + state.EditText
	}
	rows = append(rows, settingsRow(state.Cursor == 8, fmt.Sprintf("API key/env: %s (%s)", apiEnv, apiStatus), width))
	if state.Saved {
		rows = append(rows, "", neonDimStyle.Width(width).Render("Saved"))
	}
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func editableValue(state *CricketSettingsState, row int, values []string) string {
	if state.Editing && state.EditRow == row {
		return "> " + state.EditText
	}
	if len(values) == 0 {
		return "all"
	}
	return strings.Join(values, ", ")
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			values = append(values, part)
		}
	}
	return values
}

func displayAPIKeySetting(value string) string {
	value = strings.TrimSpace(value)
	if value == "" || looksLikeEnvName(value) {
		return value
	}
	if len(value) <= 8 {
		return "********"
	}
	return value[:4] + "..." + value[len(value)-4:]
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

func settingsRow(selected bool, text string, width int) string {
	style := menuItemStyle.Width(width)
	if selected {
		style = menuItemSelectedStyle.Width(width)
	}
	return style.Render(Truncate(text, width))
}

func removeString(values []string, remove string) []string {
	next := values[:0]
	for _, value := range values {
		if value != remove {
			next = append(next, value)
		}
	}
	return next
}

func clamp(value, minValue, maxValue int) int {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}
