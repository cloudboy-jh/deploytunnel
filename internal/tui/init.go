package tui

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
	"github.com/johnhorton/deploy-tunnel/internal/bridge"
	"github.com/johnhorton/deploy-tunnel/internal/keychain"
	"github.com/johnhorton/deploy-tunnel/internal/state"
)

type initStep int

const (
	stepSelectSource initStep = iota
	stepSelectTarget
	stepEnterDomain
	stepConfirm
	stepComplete
)

type InitModel struct {
	step           initStep
	sourceList     list.Model
	targetList     list.Model
	domainInput    textinput.Model
	selectedSource bridge.Provider
	selectedTarget bridge.Provider
	domain         string
	migrationID    string
	err            error
	width          int
	height         int
	stateDB        *state.DB
	bridge         *bridge.Bridge
	ctx            context.Context
}

type item struct {
	title string
	desc  string
	value bridge.Provider
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

func NewInitModel(stateDB *state.DB, br *bridge.Bridge) InitModel {
	// Provider items
	items := []list.Item{
		item{title: "Vercel", desc: "Deploy in seconds with Vercel", value: bridge.ProviderVercel},
		item{title: "Cloudflare", desc: "Pages & Workers at the edge", value: bridge.ProviderCloudflare},
		item{title: "Render", desc: "Unified cloud for web services", value: bridge.ProviderRender},
		item{title: "Netlify", desc: "All-in-one platform for web projects", value: bridge.ProviderNetlify},
	}

	// Source list
	sourceList := list.New(items, list.NewDefaultDelegate(), 0, 0)
	sourceList.Title = "Select Source Provider"
	sourceList.SetShowStatusBar(false)
	sourceList.SetFilteringEnabled(false)
	sourceList.Styles.Title = TitleStyle
	sourceList.Styles.HelpStyle = HelpStyle

	// Target list
	targetList := list.New(items, list.NewDefaultDelegate(), 0, 0)
	targetList.Title = "Select Target Provider"
	targetList.SetShowStatusBar(false)
	targetList.SetFilteringEnabled(false)
	targetList.Styles.Title = TitleStyle
	targetList.Styles.HelpStyle = HelpStyle

	// Domain input
	domainInput := textinput.New()
	domainInput.Placeholder = "example.com"
	domainInput.Focus()
	domainInput.CharLimit = 255
	domainInput.Width = 50
	domainInput.Prompt = PromptStyle.Render("► ")
	domainInput.TextStyle = InputStyle

	return InitModel{
		step:        stepSelectSource,
		sourceList:  sourceList,
		targetList:  targetList,
		domainInput: domainInput,
		stateDB:     stateDB,
		bridge:      br,
		ctx:         context.Background(),
	}
}

func (m InitModel) Init() tea.Cmd {
	return nil
}

func (m InitModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "q":
			return m, tea.Quit

		case "enter":
			return m.handleEnter()
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.sourceList.SetSize(msg.Width-4, msg.Height-10)
		m.targetList.SetSize(msg.Width-4, msg.Height-10)
		return m, nil
	}

	// Update current component for other keys (arrows, etc)
	var cmd tea.Cmd
	switch m.step {
	case stepSelectSource:
		m.sourceList, cmd = m.sourceList.Update(msg)
	case stepSelectTarget:
		m.targetList, cmd = m.targetList.Update(msg)
	case stepEnterDomain:
		m.domainInput, cmd = m.domainInput.Update(msg)
	}

	return m, cmd
}

func (m InitModel) handleEnter() (tea.Model, tea.Cmd) {
	switch m.step {
	case stepSelectSource:
		if i, ok := m.sourceList.SelectedItem().(item); ok {
			m.selectedSource = i.value
			m.step = stepSelectTarget
		}

	case stepSelectTarget:
		if i, ok := m.targetList.SelectedItem().(item); ok {
			m.selectedTarget = i.value
			m.step = stepEnterDomain
		}

	case stepEnterDomain:
		m.domain = m.domainInput.Value()
		if m.domain != "" {
			m.step = stepConfirm
		}

	case stepConfirm:
		// Create migration
		m.migrationID = uuid.New().String()
		if err := m.stateDB.CreateMigration(
			m.migrationID,
			string(m.selectedSource),
			string(m.selectedTarget),
			m.domain,
		); err != nil {
			m.err = err
			return m, nil
		}
		m.step = stepComplete
		return m, tea.Quit
	}

	return m, nil
}

func (m InitModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	header := Header()

	var content string

	switch m.step {
	case stepSelectSource:
		content = lipgloss.JoinVertical(
			lipgloss.Left,
			StepIndicator(1, 4, "Where are you migrating FROM?"),
			"",
			m.sourceList.View(),
		)

	case stepSelectTarget:
		content = lipgloss.JoinVertical(
			lipgloss.Left,
			StepIndicator(2, 4, "Where are you migrating TO?"),
			"",
			SuccessStyle.Render(fmt.Sprintf("✓ Source: %s", m.selectedSource)),
			"",
			m.targetList.View(),
		)

	case stepEnterDomain:
		content = lipgloss.JoinVertical(
			lipgloss.Left,
			StepIndicator(3, 4, "What domain are you migrating?"),
			"",
			SuccessStyle.Render(fmt.Sprintf("✓ Source: %s", m.selectedSource)),
			SuccessStyle.Render(fmt.Sprintf("✓ Target: %s", m.selectedTarget)),
			"",
			PromptStyle.Render("Domain name:"),
			m.domainInput.View(),
			"",
			HelpStyle.Render("Press Enter to continue"),
		)

	case stepConfirm:
		// Check auth status
		sourceAuth, _ := keychain.Get(string(m.selectedSource))
		targetAuth, _ := keychain.Get(string(m.selectedTarget))

		sourceStatus := RedStyle.Render("✗ Not authenticated")
		if sourceAuth != "" {
			sourceStatus = GreenStyle.Render("✓ Authenticated")
		}

		targetStatus := RedStyle.Render("✗ Not authenticated")
		if targetAuth != "" {
			targetStatus = GreenStyle.Render("✓ Authenticated")
		}

		confirmBox := BoxStyle.Render(lipgloss.JoinVertical(
			lipgloss.Left,
			TitleStyle.Render("Migration Summary"),
			"",
			fmt.Sprintf("Source:     %s", SelectedItemStyle.Render(string(m.selectedSource))),
			fmt.Sprintf("            %s", sourceStatus),
			"",
			fmt.Sprintf("Target:     %s", SelectedItemStyle.Render(string(m.selectedTarget))),
			fmt.Sprintf("            %s", targetStatus),
			"",
			fmt.Sprintf("Domain:     %s", SelectedItemStyle.Render(m.domain)),
		))

		content = lipgloss.JoinVertical(
			lipgloss.Left,
			StepIndicator(4, 4, "Confirm Migration Setup"),
			"",
			confirmBox,
			"",
			HelpStyle.Render("Press Enter to create migration • q to cancel"),
		)

	case stepComplete:
		if m.err != nil {
			content = ErrorStyle.Render(fmt.Sprintf("Error: %s", m.err))
		} else {
			content = lipgloss.JoinVertical(
				lipgloss.Left,
				SuccessStyle.Render("✓ Migration initialized successfully!"),
				"",
				BoxStyle.Render(lipgloss.JoinVertical(
					lipgloss.Left,
					PromptStyle.Render("Migration ID:"),
					InputStyle.Render(m.migrationID),
					"",
					PromptStyle.Render("Next Steps:"),
					UnselectedItemStyle.Render("1. Run 'dt' to open the dashboard"),
					UnselectedItemStyle.Render("2. Authenticate with your providers"),
					UnselectedItemStyle.Render("3. Start the migration workflow"),
				)),
			)
		}
	}

	footer := StatusBarStyle.Render(
		fmt.Sprintf(" %s | Press 'q' to quit ", "Deploy Tunnel v1.0"),
	)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		content,
		"",
		footer,
	)
}

// RunInitTUI runs the interactive init TUI
func RunInitTUI(stateDB *state.DB, br *bridge.Bridge) error {
	p := tea.NewProgram(
		NewInitModel(stateDB, br),
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		return err
	}

	return nil
}
