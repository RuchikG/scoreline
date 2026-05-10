package ui

import (
	"github.com/RuchikG/scoreline/internal/constants"
	"github.com/RuchikG/scoreline/internal/ui/design"
	"github.com/charmbracelet/lipgloss"
)

// RenderPlaceholderView renders a simple sport-native placeholder screen.
func RenderPlaceholderView(width, height int, title, message, help string, bannerType constants.StatusBannerType) string {
	const contentWidth = 72

	statusBanner := renderStatusBanner(bannerType, contentWidth)
	if statusBanner != "" {
		statusBanner += "\n"
	}

	titleView := design.RenderHeader(title, contentWidth)
	messageView := lipgloss.NewStyle().
		Width(contentWidth).
		Align(lipgloss.Center).
		Foreground(textColor).
		Render(message)
	helpView := menuHelpStyle.Render(help)

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		statusBanner,
		titleView,
		"",
		messageView,
		"",
		helpView,
	)

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content)
}
