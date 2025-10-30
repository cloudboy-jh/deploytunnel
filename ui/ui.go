package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Colors
	coral     = lipgloss.Color("#ef9f76")
	gray      = lipgloss.Color("#6c6f85")
	lightGray = lipgloss.Color("#a5adce")
	green     = lipgloss.Color("#a6d189")
	red       = lipgloss.Color("#e78284")
	yellow    = lipgloss.Color("#e5c890")

	// Styles
	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(coral).
			MarginTop(1).
			MarginBottom(1)

	SubheaderStyle = lipgloss.NewStyle().
			Foreground(lightGray).
			Italic(true)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(green).
			Bold(true)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(red).
			Bold(true)

	WarningStyle = lipgloss.NewStyle().
			Foreground(yellow)

	InfoStyle = lipgloss.NewStyle().
			Foreground(lightGray)

	KeyStyle = lipgloss.NewStyle().
			Foreground(coral).
			Bold(true)

	ValueStyle = lipgloss.NewStyle().
			Foreground(lightGray)

	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(coral).
			Padding(1, 2).
			MarginTop(1)

	SpinnerStyle = lipgloss.NewStyle().
			Foreground(coral)
)

// Header renders the Deploy Tunnel header
func Header() string {
	title := HeaderStyle.Render("DEPLOY ▸ TUNNEL")
	subtitle := SubheaderStyle.Render("migrate safely between hosts")

	return BoxStyle.Render(
		fmt.Sprintf("%s\n%s", title, subtitle),
	)
}

// Success renders a success message
func Success(message string) string {
	return SuccessStyle.Render("✓ " + message)
}

// Error renders an error message
func Error(message string) string {
	return ErrorStyle.Render("✗ " + message)
}

// Warning renders a warning message
func Warning(message string) string {
	return WarningStyle.Render("⚠ " + message)
}

// Info renders an info message
func Info(message string) string {
	return InfoStyle.Render("ℹ " + message)
}

// KeyValue renders a key-value pair
func KeyValue(key, value string) string {
	return fmt.Sprintf("%s %s",
		KeyStyle.Render(key+":"),
		ValueStyle.Render(value),
	)
}

// List renders a bulleted list
func List(items []string) string {
	var lines []string
	for _, item := range items {
		lines = append(lines, InfoStyle.Render("  • "+item))
	}
	return strings.Join(lines, "\n")
}

// Step renders a step in a process
func Step(current, total int, message string) string {
	prefix := InfoStyle.Render(fmt.Sprintf("[%d/%d]", current, total))
	return fmt.Sprintf("%s %s", prefix, message)
}

// Spinner frames for CLI animations
var SpinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// Table renders a simple table
func Table(headers []string, rows [][]string) string {
	if len(rows) == 0 {
		return InfoStyle.Render("No data")
	}

	// Calculate column widths
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	// Build table
	var sb strings.Builder

	// Header row
	for i, h := range headers {
		sb.WriteString(KeyStyle.Render(padRight(h, widths[i])))
		if i < len(headers)-1 {
			sb.WriteString("  ")
		}
	}
	sb.WriteString("\n")

	// Separator
	for i, w := range widths {
		sb.WriteString(strings.Repeat("─", w))
		if i < len(widths)-1 {
			sb.WriteString("  ")
		}
	}
	sb.WriteString("\n")

	// Data rows
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) {
				sb.WriteString(ValueStyle.Render(padRight(cell, widths[i])))
				if i < len(row)-1 {
					sb.WriteString("  ")
				}
			}
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func padRight(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}

// ProgressBar renders a simple progress bar
func ProgressBar(current, total int, width int) string {
	if width <= 0 {
		width = 40
	}
	if total <= 0 {
		return ""
	}

	percent := float64(current) / float64(total)
	filled := int(float64(width) * percent)

	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	percentText := fmt.Sprintf(" %3.0f%%", percent*100)

	return SpinnerStyle.Render(bar) + InfoStyle.Render(percentText)
}

// Confirm formats a confirmation prompt
func Confirm(message string) string {
	return fmt.Sprintf("%s %s",
		KeyStyle.Render("?"),
		message+" (y/n)",
	)
}

// Select formats a selection prompt
func Select(message string, options []string) string {
	lines := []string{
		KeyStyle.Render("?") + " " + message,
		"",
	}

	for i, opt := range options {
		lines = append(lines, InfoStyle.Render(fmt.Sprintf("  %d) %s", i+1, opt)))
	}

	lines = append(lines, "")
	lines = append(lines, InfoStyle.Render("Enter number: "))

	return strings.Join(lines, "\n")
}
