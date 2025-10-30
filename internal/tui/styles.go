package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	Coral     = lipgloss.Color("#ef9f76")
	Gray      = lipgloss.Color("#6c6f85")
	LightGray = lipgloss.Color("#a5adce")
	Green     = lipgloss.Color("#a6d189")
	Red       = lipgloss.Color("#e78284")
	Yellow    = lipgloss.Color("#e5c890")
	Blue      = lipgloss.Color("#8caaee")

	// Color render styles
	GreenStyle  = lipgloss.NewStyle().Foreground(Green)
	RedStyle    = lipgloss.NewStyle().Foreground(Red)
	YellowStyle = lipgloss.NewStyle().Foreground(Yellow)

	// Base styles
	BaseStyle = lipgloss.NewStyle().
			Padding(1, 2)

	TitleStyle = lipgloss.NewStyle().
			Foreground(Coral).
			Bold(true).
			Padding(0, 1)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(LightGray).
			Italic(true)

	SelectedItemStyle = lipgloss.NewStyle().
				Foreground(Coral).
				Bold(true).
				PaddingLeft(2)

	UnselectedItemStyle = lipgloss.NewStyle().
				Foreground(LightGray).
				PaddingLeft(2)

	PromptStyle = lipgloss.NewStyle().
			Foreground(Coral).
			Bold(true)

	InputStyle = lipgloss.NewStyle().
			Foreground(LightGray)

	HelpStyle = lipgloss.NewStyle().
			Foreground(Gray).
			Italic(true)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(Red).
			Bold(true)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(Green).
			Bold(true)

	StatusBarStyle = lipgloss.NewStyle().
			Foreground(LightGray).
			Background(Gray).
			Padding(0, 1)

	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Coral).
			Padding(1, 2)

	ProgressBarStyle = lipgloss.NewStyle().
				Foreground(Coral)

	ProgressEmptyStyle = lipgloss.NewStyle().
				Foreground(Gray)
)

// Renders the Deploy Tunnel header with optional image
func Header() string {
	// Try to display the logo image
	imageDisplay := DisplayImage()

	title := TitleStyle.Render("DEPLOY ▸ TUNNEL")
	subtitle := SubtitleStyle.Render("migrate safely between hosts")

	// If we have an image, show it above the text
	if imageDisplay != "" {
		return lipgloss.JoinVertical(
			lipgloss.Left,
			"",
			imageDisplay,
			title,
			subtitle,
			"",
		)
	}

	// Otherwise just show the text header
	return lipgloss.JoinVertical(
		lipgloss.Left,
		"",
		title,
		subtitle,
		"",
	)
}

// Renders a step indicator
func StepIndicator(current, total int, description string) string {
	steps := ""
	for i := 1; i <= total; i++ {
		if i == current {
			steps += SelectedItemStyle.Render("●")
		} else if i < current {
			steps += SuccessStyle.Render("●")
		} else {
			steps += UnselectedItemStyle.Render("○")
		}
		if i < total {
			steps += " "
		}
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		steps,
		"",
		PromptStyle.Render(description),
	)
}
