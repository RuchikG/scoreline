package ui

import (
	"fmt"
	"strings"

	"github.com/RuchikG/scoreline/internal/constants"
	"github.com/RuchikG/scoreline/internal/cricket"
	"github.com/RuchikG/scoreline/internal/ui/design"
	"github.com/charmbracelet/lipgloss"
)

// RenderCricketArchivesView renders locally indexed Cricsheet archive matches.
func RenderCricketArchivesView(width, height int, matches []cricket.Match, selected int, details *cricket.MatchDetails, loading bool, lastError string, bannerType constants.StatusBannerType) string {
	const panelWidth = 46

	statusBanner := renderStatusBanner(bannerType, width)
	if statusBanner != "" {
		statusBanner += "\n"
	}

	title := design.RenderHeader("Cricket Archives", width)
	list := renderCricketArchiveList(matches, selected, panelWidth, loading, lastError)
	detail := renderCricketDetails(details, max(width-panelWidth-6, 40))
	body := lipgloss.JoinHorizontal(lipgloss.Top, list, "  ", detail)
	help := menuHelpStyle.Render("j/k: navigate  r: refresh cache  Esc: back  q: quit")

	content := lipgloss.JoinVertical(lipgloss.Center, statusBanner, title, "", body, "", help)
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content)
}

func renderCricketArchiveList(matches []cricket.Match, selected int, width int, loading bool, lastError string) string {
	header := design.RenderHeader("Completed Matches", width)
	if loading {
		return lipgloss.JoinVertical(lipgloss.Left, header, "", neonDimStyle.Width(width).Render("Refreshing Cricsheet archive cache..."))
	}
	if lastError != "" {
		return lipgloss.JoinVertical(lipgloss.Left, header, "", neonEmptyStyle.Width(width).Render(lastError))
	}
	if len(matches) == 0 {
		return lipgloss.JoinVertical(lipgloss.Left, header, "", neonEmptyStyle.Width(width).Render("No archived matches cached"))
	}

	rows := make([]string, 0, len(matches)*3)
	for i, match := range matches {
		date := ""
		if !match.StartTime.IsZero() {
			date = match.StartTime.Format("2006-01-02")
		}
		line := fmt.Sprintf("%s  %s", cricketTeams(match), match.CurrentScoreSummary)
		meta := strings.Join(nonEmptyStrings(date, match.MatchType, match.Competition), " · ")
		style := menuItemStyle.Width(width)
		if i == selected {
			style = menuItemSelectedStyle.Width(width)
		}
		rows = append(rows, style.Render(Truncate(line, width)))
		rows = append(rows, lipgloss.NewStyle().Foreground(dimColor).Width(width).Render(Truncate(meta, width)))
		rows = append(rows, lipgloss.NewStyle().Foreground(dimColor).Width(width).Render(Truncate(match.Venue, width)))
	}

	return lipgloss.JoinVertical(lipgloss.Left, append([]string{header, ""}, rows...)...)
}

func nonEmptyStrings(values ...string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			result = append(result, value)
		}
	}
	return result
}
