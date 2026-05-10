package ui

import (
	"fmt"
	"strings"

	"github.com/RuchikG/scoreline/internal/constants"
	"github.com/RuchikG/scoreline/internal/cricket"
	"github.com/RuchikG/scoreline/internal/ui/design"
	"github.com/charmbracelet/lipgloss"
)

// RenderCricketLiveView renders a cricket-native live dashboard.
func RenderCricketLiveView(width, height int, matches []cricket.Match, selected int, details *cricket.MatchDetails, bannerType constants.StatusBannerType) string {
	const panelWidth = 42

	statusBanner := renderStatusBanner(bannerType, width)
	if statusBanner != "" {
		statusBanner += "\n"
	}

	title := design.RenderHeader("Cricket Live", width)
	list := renderCricketMatchList(matches, selected, panelWidth)
	detail := renderCricketDetails(details, max(width-panelWidth-6, 40))

	body := lipgloss.JoinHorizontal(lipgloss.Top, list, "  ", detail)
	help := menuHelpStyle.Render("j/k: navigate  Esc: back  q: quit")

	content := lipgloss.JoinVertical(lipgloss.Center, statusBanner, title, "", body, "", help)
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content)
}

func renderCricketMatchList(matches []cricket.Match, selected int, width int) string {
	header := design.RenderHeader("Matches", width)
	if len(matches) == 0 {
		return lipgloss.JoinVertical(lipgloss.Left, header, "", neonEmptyStyle.Width(width).Render("No cricket matches available"))
	}

	rows := make([]string, 0, len(matches))
	for i, match := range matches {
		teams := cricketTeams(match)
		line := fmt.Sprintf("%s  %s", teams, match.CurrentScoreSummary)
		style := menuItemStyle.Width(width)
		if i == selected {
			style = menuItemSelectedStyle.Width(width)
		}
		rows = append(rows, style.Render(Truncate(line, width)))
		rows = append(rows, lipgloss.NewStyle().Foreground(dimColor).Width(width).Render(Truncate(match.Competition+" · "+match.Venue, width)))
	}

	return lipgloss.JoinVertical(lipgloss.Left, append([]string{header, ""}, rows...)...)
}

func renderCricketDetails(details *cricket.MatchDetails, width int) string {
	header := design.RenderHeader("Scorecard", width)
	if details == nil {
		return lipgloss.JoinVertical(lipgloss.Left, header, "", neonEmptyStyle.Width(width).Render("Select a cricket match"))
	}

	lines := []string{
		menuItemSelectedStyle.Render(cricketTeams(details.Match)),
		neonDimStyle.Render(details.Match.Competition+" · "+details.Match.MatchType),
		neonDimStyle.Render(details.Match.Venue),
		"",
		fmt.Sprintf("Status: %s", details.Match.Status),
		fmt.Sprintf("Score: %s", details.Match.CurrentScoreSummary),
	}
	if details.Toss != "" {
		lines = append(lines, "Toss: "+details.Toss)
	}
	if details.Result != "" {
		lines = append(lines, "Result: "+details.Result)
	}

	if len(details.CurrentBatters) > 0 {
		lines = append(lines, "", menuItemSelectedStyle.Render("Batters"))
		for _, batter := range details.CurrentBatters {
			lines = append(lines, fmt.Sprintf("%s  %d (%d)  SR %s", batter.Player, batter.Runs, batter.Balls, batter.StrikeRate))
		}
	}
	if details.CurrentBowler != nil {
		bowler := details.CurrentBowler
		lines = append(lines, "", menuItemSelectedStyle.Render("Bowler"))
		lines = append(lines, fmt.Sprintf("%s  %s-%d-%d-%d  Econ %s", bowler.Player, bowler.Overs, bowler.Maidens, bowler.Runs, bowler.Wickets, bowler.Economy))
	}
	if len(details.RecentOvers) > 0 {
		lines = append(lines, "", menuItemSelectedStyle.Render("Recent Overs"))
		lines = append(lines, details.RecentOvers...)
	}
	if len(details.Innings) > 0 {
		lines = append(lines, "", menuItemSelectedStyle.Render("Innings"))
		for _, innings := range details.Innings {
			lines = append(lines, fmt.Sprintf("%s %d/%d (%s ov)", innings.BattingTeam.ShortName, innings.Runs, innings.Wickets, innings.Overs))
			for _, batter := range innings.BattingCard {
				lines = append(lines, fmt.Sprintf("  %s  %d (%d)", batter.Player, batter.Runs, batter.Balls))
			}
			for _, bowler := range innings.BowlingCard {
				lines = append(lines, fmt.Sprintf("  %s  %s-%d-%d-%d", bowler.Player, bowler.Overs, bowler.Maidens, bowler.Runs, bowler.Wickets))
			}
		}
	}

	return lipgloss.JoinVertical(lipgloss.Left, header, "", lipgloss.NewStyle().Width(width).Render(strings.Join(lines, "\n")))
}

func cricketTeams(match cricket.Match) string {
	if len(match.Teams) < 2 {
		return "TBD"
	}
	return match.Teams[0].ShortName + " v " + match.Teams[1].ShortName
}
