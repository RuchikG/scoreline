// Package ui provides rendering functions for the terminal user interface.
package ui

import (
	"strings"

	"github.com/RuchikG/scoreline/internal/constants"
	"github.com/RuchikG/scoreline/internal/ui/design"
	"github.com/RuchikG/scoreline/internal/ui/logo"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
)

// logoWidth is the standard width for the logo container.
const logoWidth = 80

var (
	// Menu styles
	menuItemStyle = lipgloss.NewStyle().
			Foreground(textColor).
			Padding(0, 0)

	menuItemSelectedStyle = lipgloss.NewStyle().
				Foreground(highlightColor).
				Bold(true).
				Padding(0, 0)

	menuHelpStyle = lipgloss.NewStyle().
			Foreground(dimColor).
			Align(lipgloss.Center).
			Padding(0, 0)
)

// RenderSportSelector renders the first-run sport selection menu.
func RenderSportSelector(width, height, selected int, bannerType constants.StatusBannerType, animatedLogo *logo.AnimatedLogo) string {
	return renderMenu(width, height, selected, []string{
		constants.MenuSoccer,
		constants.MenuCricket,
	}, constants.HelpSportSelector, false, nil, bannerType, animatedLogo)
}

// RenderMainMenu renders the soccer main menu view with navigation options.
// width and height specify the terminal dimensions.
// selected indicates which menu item is currently selected (0-indexed).
// sp is the spinner model to display when loading (for other views).
// randomSpinner is the random character spinner for main view.
// loading indicates if the spinner should be shown.
// bannerType determines what status banner (if any) to display at the top.
// animatedLogo is the animated logo instance for the main view.
func RenderMainMenu(width, height, selected int, sp spinner.Model, randomSpinner *RandomCharSpinner, loading bool, bannerType constants.StatusBannerType, animatedLogo *logo.AnimatedLogo) string {
	return renderMenu(width, height, selected, []string{
		constants.MenuStats,
		constants.MenuLiveMatches,
		constants.MenuSettings,
	}, constants.HelpSportMainMenu, loading, randomSpinner, bannerType, animatedLogo)
}

// RenderCricketMenu renders the cricket main menu.
func RenderCricketMenu(width, height, selected int, bannerType constants.StatusBannerType, animatedLogo *logo.AnimatedLogo) string {
	return renderMenu(width, height, selected, []string{
		constants.MenuCricketLive,
		constants.MenuArchives,
		constants.MenuSettings,
	}, constants.HelpSportMainMenu, false, nil, bannerType, animatedLogo)
}

func renderMenu(width, height, selected int, menuItems []string, helpText string, loading bool, randomSpinner *RandomCharSpinner, bannerType constants.StatusBannerType, animatedLogo *logo.AnimatedLogo) string {
	items := make([]string, 0, len(menuItems))
	for i, item := range menuItems {
		if i == selected {
			items = append(items, menuItemSelectedStyle.Render(item))
		} else {
			items = append(items, menuItemStyle.Render(item))
		}
	}

	menuContent := strings.Join(items, "\n")

	// Get logo content from animated logo (handles animation state internally)
	logoContent := logo.Render("", false, logo.DefaultOpts())
	if animatedLogo != nil {
		logoContent = animatedLogo.View()
	}

	// Place logo in centered container
	title := lipgloss.NewStyle().
		Width(logoWidth).
		Align(lipgloss.Center).
		Render(logoContent)
	help := menuHelpStyle.Render(helpText)

	// Spinner with fixed spacing - always reserve space to prevent movement
	// Use multiple spinner instances for a longer, more prominent animation
	spinnerStyle := lipgloss.NewStyle().
		Align(lipgloss.Center).
		Height(1).
		Padding(0, 0)

	var spinnerContent string
	if loading && randomSpinner != nil {
		// Use random character spinner for main view
		spinnerContent = spinnerStyle.Render(randomSpinner.View())
	} else {
		// Reserve space even when not loading to prevent menu movement
		spinnerContent = spinnerStyle.Render("")
	}

	// Add status banner if needed
	statusBanner := renderStatusBanner(bannerType, width)
	if statusBanner != "" {
		statusBanner += "\n"
	}

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		statusBanner,
		spinnerContent,
		"\n",
		menuContent,
		"\n",
		help,
	)

	return lipgloss.Place(
		width,
		height,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
}

// RenderGradientText applies a gradient (cyan to red) to multi-line text.
// Exported wrapper for external use.
func RenderGradientText(text string) string {
	return design.ApplyGradientToMultilineText(text)
}
