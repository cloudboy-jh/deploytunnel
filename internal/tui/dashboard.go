package tui

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/johnhorton/deploy-tunnel/internal/bridge"
	"github.com/johnhorton/deploy-tunnel/internal/state"
)

type menuItem struct {
	title string
	desc  string
	key   string
}

func (i menuItem) Title() string       { return i.title }
func (i menuItem) Description() string { return i.desc }
func (i menuItem) FilterValue() string { return i.title }

type DashboardModel struct {
	list      list.Model
	stateDB   *state.DB
	bridge    *bridge.Bridge
	ctx       context.Context
	width     int
	height    int
	selected  string
	quitting  bool
	migration *state.Migration
}

func NewDashboardModel(stateDB *state.DB, br *bridge.Bridge) DashboardModel {
	items := []list.Item{
		menuItem{
			title: "Start New Migration",
			desc:  "Initialize a new migration between providers",
			key:   "init",
		},
		menuItem{
			title: "View Migrations",
			desc:  "See your migration history and status",
			key:   "list",
		},
		menuItem{
			title: "Manage Auth",
			desc:  "Authenticate with providers",
			key:   "auth",
		},
		menuItem{
			title: "Current Migration",
			desc:  "Continue working on your active migration",
			key:   "current",
		},
		menuItem{
			title: "Exit",
			desc:  "Quit Deploy Tunnel",
			key:   "quit",
		},
	}

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = SelectedItemStyle
	delegate.Styles.SelectedDesc = lipgloss.NewStyle().Foreground(LightGray)

	l := list.New(items, delegate, 0, 0)
	l.Title = "Main Menu"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = TitleStyle
	l.Styles.HelpStyle = HelpStyle

	// Try to load the most recent migration
	migrations, _ := stateDB.ListMigrations("")
	var currentMigration *state.Migration
	if len(migrations) > 0 {
		currentMigration = &migrations[0]
	}

	return DashboardModel{
		list:      l,
		stateDB:   stateDB,
		bridge:    br,
		ctx:       context.Background(),
		migration: currentMigration,
	}
}

func (m DashboardModel) Init() tea.Cmd {
	return nil
}

func (m DashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "q":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			if i, ok := m.list.SelectedItem().(menuItem); ok {
				m.selected = i.key

				switch i.key {
				case "quit":
					m.quitting = true
					return m, tea.Quit

				case "init":
					// Launch init TUI
					return m, func() tea.Msg {
						return switchToInitMsg{}
					}

				case "auth":
					// Launch auth TUI
					return m, func() tea.Msg {
						return switchToAuthMsg{}
					}

				case "current":
					if m.migration != nil {
						// Launch migration workflow TUI
						return m, func() tea.Msg {
							return switchToMigrationMsg{migration: m.migration}
						}
					}

				case "list":
					// Launch migration list TUI
					return m, func() tea.Msg {
						return switchToListMsg{}
					}
				}
			}
			// Don't propagate enter to list if we handled it
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width-4, msg.Height-15)
		return m, nil
	}

	// Update list for other keys (arrow keys, etc)
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m DashboardModel) View() string {
	if m.quitting {
		return SuccessStyle.Render("Thanks for using Deploy Tunnel!\n")
	}

	if m.width == 0 {
		return "Loading..."
	}

	header := Header()

	// Show current migration info if exists
	var migrationInfo string
	if m.migration != nil {
		statusStyle := YellowStyle
		if m.migration.Status == "completed" {
			statusStyle = GreenStyle
		} else if m.migration.Status == "failed" {
			statusStyle = RedStyle
		}

		migrationInfo = BoxStyle.Render(lipgloss.JoinVertical(
			lipgloss.Left,
			PromptStyle.Render("Active Migration"),
			"",
			fmt.Sprintf("Domain:  %s", InputStyle.Render(m.migration.Domain)),
			fmt.Sprintf("Source:  %s", InputStyle.Render(m.migration.Source)),
			fmt.Sprintf("Target:  %s", InputStyle.Render(m.migration.Target)),
			fmt.Sprintf("Status:  %s", statusStyle.Render(m.migration.Status)),
		))
	} else {
		migrationInfo = BoxStyle.Render(
			HelpStyle.Render("No active migrations. Start a new one!"),
		)
	}

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		migrationInfo,
		"",
		m.list.View(),
	)

	footer := StatusBarStyle.Render(
		fmt.Sprintf(" Deploy Tunnel v1.0 | ↑↓ navigate • enter select • q quit "),
	)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		content,
		"",
		footer,
	)
}

// Messages for switching between TUIs
type switchToInitMsg struct{}
type switchToAuthMsg struct{}
type switchToListMsg struct{}
type switchToMigrationMsg struct {
	migration *state.Migration
}

// RunDashboardTUI runs the main dashboard TUI
func RunDashboardTUI(stateDB *state.DB, br *bridge.Bridge) error {
	p := tea.NewProgram(
		NewDashboardModel(stateDB, br),
		tea.WithAltScreen(),
	)

	model, err := p.Run()
	if err != nil {
		return err
	}

	// Check if we need to switch to another TUI
	if m, ok := model.(DashboardModel); ok {
		switch m.selected {
		case "init":
			return RunInitTUI(stateDB, br)
		case "auth":
			return RunAuthTUI(stateDB, br)
			// Add more cases as we build more TUIs
		}
	}

	return nil
}
