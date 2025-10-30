package tui

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/johnhorton/deploy-tunnel/internal/bridge"
	"github.com/johnhorton/deploy-tunnel/internal/keychain"
	"github.com/johnhorton/deploy-tunnel/internal/state"
)

type authStep int

const (
	authStepMenu authStep = iota
	authStepSelectProvider
	authStepFetchingCapabilities
	authStepEnterToken
	authStepVerifying
	authStepComplete
	authStepError
)

type AuthModel struct {
	step               authStep
	menuList           list.Model
	providerList       list.Model
	tokenInput         textinput.Model
	spinner            spinner.Model
	selectedAction     string
	selectedProvider   bridge.Provider
	capabilities       *bridge.CapabilitiesData
	authData           *bridge.AuthStartData
	token              string
	err                error
	successMessage     string
	width              int
	height             int
	stateDB            *state.DB
	bridge             *bridge.Bridge
	ctx                context.Context
	authenticatedProvs []string
}

type authMenuItem struct {
	title string
	desc  string
	key   string
}

func (i authMenuItem) Title() string       { return i.title }
func (i authMenuItem) Description() string { return i.desc }
func (i authMenuItem) FilterValue() string { return i.title }

type providerItem struct {
	title  string
	desc   string
	value  bridge.Provider
	authed bool
}

func (i providerItem) Title() string {
	if i.authed {
		return GreenStyle.Render("✓ ") + i.title
	}
	return "  " + i.title
}
func (i providerItem) Description() string { return i.desc }
func (i providerItem) FilterValue() string { return i.title }

func NewAuthModel(stateDB *state.DB, br *bridge.Bridge) AuthModel {
	// Get authenticated providers
	authedProviders, _ := keychain.List()
	authedMap := make(map[string]bool)
	for _, p := range authedProviders {
		authedMap[p] = true
	}

	// Menu items
	menuItems := []list.Item{
		authMenuItem{
			title: "Authenticate Provider",
			desc:  "Add credentials for a new provider",
			key:   "auth",
		},
		authMenuItem{
			title: "List Authenticated",
			desc:  "View currently authenticated providers",
			key:   "list",
		},
		authMenuItem{
			title: "Revoke Credentials",
			desc:  "Remove stored credentials",
			key:   "revoke",
		},
		authMenuItem{
			title: "Back to Dashboard",
			desc:  "Return to main menu",
			key:   "back",
		},
	}

	menuList := list.New(menuItems, list.NewDefaultDelegate(), 0, 0)
	menuList.Title = "Authentication Menu"
	menuList.SetShowStatusBar(false)
	menuList.SetFilteringEnabled(false)
	menuList.Styles.Title = TitleStyle

	// Provider items
	providerItems := []list.Item{
		providerItem{
			title:  "Vercel",
			desc:   "Deploy in seconds with Vercel",
			value:  bridge.ProviderVercel,
			authed: authedMap["vercel"],
		},
		providerItem{
			title:  "Cloudflare",
			desc:   "Pages & Workers at the edge",
			value:  bridge.ProviderCloudflare,
			authed: authedMap["cloudflare"],
		},
		providerItem{
			title:  "Render",
			desc:   "Unified cloud for web services",
			value:  bridge.ProviderRender,
			authed: authedMap["render"],
		},
		providerItem{
			title:  "Netlify",
			desc:   "All-in-one platform for web projects",
			value:  bridge.ProviderNetlify,
			authed: authedMap["netlify"],
		},
	}

	providerList := list.New(providerItems, list.NewDefaultDelegate(), 0, 0)
	providerList.Title = "Select Provider"
	providerList.SetShowStatusBar(false)
	providerList.SetFilteringEnabled(false)
	providerList.Styles.Title = TitleStyle

	// Token input
	tokenInput := textinput.New()
	tokenInput.Placeholder = "Paste your token here"
	tokenInput.EchoMode = textinput.EchoPassword
	tokenInput.EchoCharacter = '•'
	tokenInput.Focus()
	tokenInput.CharLimit = 500
	tokenInput.Width = 60
	tokenInput.Prompt = PromptStyle.Render("► ")

	// Spinner
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(Coral)

	return AuthModel{
		step:               authStepMenu,
		menuList:           menuList,
		providerList:       providerList,
		tokenInput:         tokenInput,
		spinner:            s,
		stateDB:            stateDB,
		bridge:             br,
		ctx:                context.Background(),
		authenticatedProvs: authedProviders,
	}
}

func (m AuthModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m AuthModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "q":
			if m.step == authStepMenu || m.step == authStepComplete || m.step == authStepError {
				return m, tea.Quit
			}

		case "enter":
			return m.handleEnter()
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.menuList.SetSize(msg.Width-4, msg.Height-10)
		m.providerList.SetSize(msg.Width-4, msg.Height-10)
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case capabilitiesMsg:
		m.capabilities = msg.caps
		m.authData = msg.authData
		if msg.err != nil {
			m.err = msg.err
			m.step = authStepError
		} else {
			m.step = authStepEnterToken
		}
		return m, nil

	case verifyMsg:
		if msg.err != nil {
			m.err = msg.err
			m.step = authStepError
		} else {
			m.successMessage = fmt.Sprintf("✓ Successfully authenticated with %s!", m.selectedProvider)
			m.step = authStepComplete
		}
		return m, nil
	}

	// Update current component for other keys (arrows, typing, etc)
	var cmd tea.Cmd
	switch m.step {
	case authStepMenu:
		m.menuList, cmd = m.menuList.Update(msg)
	case authStepSelectProvider:
		m.providerList, cmd = m.providerList.Update(msg)
	case authStepEnterToken:
		m.tokenInput, cmd = m.tokenInput.Update(msg)
	}

	return m, cmd
}

func (m AuthModel) handleEnter() (tea.Model, tea.Cmd) {
	switch m.step {
	case authStepMenu:
		if i, ok := m.menuList.SelectedItem().(authMenuItem); ok {
			m.selectedAction = i.key

			switch i.key {
			case "back":
				return m, tea.Quit
			case "auth":
				m.step = authStepSelectProvider
			case "list":
				m.step = authStepComplete
				if len(m.authenticatedProvs) == 0 {
					m.successMessage = "No providers authenticated yet."
				} else {
					m.successMessage = "Authenticated providers:\n\n"
					for _, p := range m.authenticatedProvs {
						m.successMessage += GreenStyle.Render("✓ ") + p + "\n"
					}
				}
			}
		}

	case authStepSelectProvider:
		if i, ok := m.providerList.SelectedItem().(providerItem); ok {
			m.selectedProvider = i.value
			m.step = authStepFetchingCapabilities
			return m, fetchCapabilitiesCmd(m.bridge, m.ctx, m.selectedProvider)
		}

	case authStepEnterToken:
		m.token = m.tokenInput.Value()
		if m.token != "" {
			m.step = authStepVerifying
			return m, verifyTokenCmd(m.bridge, m.ctx, m.selectedProvider, m.token)
		}

	case authStepComplete, authStepError:
		return m, tea.Quit
	}

	return m, nil
}

func (m AuthModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	header := Header()
	var content string

	switch m.step {
	case authStepMenu:
		content = m.menuList.View()

	case authStepSelectProvider:
		content = lipgloss.JoinVertical(
			lipgloss.Left,
			PromptStyle.Render("Select provider to authenticate:"),
			"",
			m.providerList.View(),
		)

	case authStepFetchingCapabilities:
		content = lipgloss.JoinVertical(
			lipgloss.Left,
			m.spinner.View()+" Fetching provider capabilities...",
		)

	case authStepEnterToken:
		var instructions string
		if m.authData != nil && m.authData.AuthURL != "" {
			instructions = lipgloss.JoinVertical(
				lipgloss.Left,
				PromptStyle.Render("Get your token:"),
				InputStyle.Render(m.authData.AuthURL),
				"",
				HelpStyle.Render("Opening in browser..."),
				"",
			)
			// Open browser
			openBrowser(m.authData.AuthURL)
		} else {
			instructions = lipgloss.JoinVertical(
				lipgloss.Left,
				HelpStyle.Render(fmt.Sprintf("Get your %s token from your account settings", m.selectedProvider)),
				"",
			)
		}

		content = lipgloss.JoinVertical(
			lipgloss.Left,
			SuccessStyle.Render(fmt.Sprintf("✓ Adapter: %s v%s", m.capabilities.AdapterName, m.capabilities.AdapterVersion)),
			PromptStyle.Render(fmt.Sprintf("Auth Type: %s", m.capabilities.AuthType)),
			"",
			instructions,
			PromptStyle.Render("Paste your token:"),
			m.tokenInput.View(),
			"",
			HelpStyle.Render("Press Enter to continue • Token will be stored securely in your system keychain"),
		)

	case authStepVerifying:
		content = lipgloss.JoinVertical(
			lipgloss.Left,
			m.spinner.View()+" Verifying credentials...",
		)

	case authStepComplete:
		content = lipgloss.JoinVertical(
			lipgloss.Left,
			SuccessStyle.Render(m.successMessage),
			"",
			HelpStyle.Render("Press q to return to dashboard"),
		)

	case authStepError:
		content = lipgloss.JoinVertical(
			lipgloss.Left,
			ErrorStyle.Render(fmt.Sprintf("✗ Error: %s", m.err)),
			"",
			HelpStyle.Render("Press q to return"),
		)
	}

	footer := StatusBarStyle.Render(" Deploy Tunnel Auth | q: back ")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		content,
		"",
		footer,
	)
}

// Messages
type capabilitiesMsg struct {
	caps     *bridge.CapabilitiesData
	authData *bridge.AuthStartData
	err      error
}

type verifyMsg struct {
	err error
}

// Commands
func fetchCapabilitiesCmd(br *bridge.Bridge, ctx context.Context, provider bridge.Provider) tea.Cmd {
	return func() tea.Msg {
		caps, err := br.Capabilities(ctx, provider)
		if err != nil {
			return capabilitiesMsg{err: err}
		}

		authData, err := br.AuthStart(ctx, bridge.AuthStartParams{
			Provider: provider,
		})
		if err != nil {
			return capabilitiesMsg{err: err}
		}

		return capabilitiesMsg{caps: caps, authData: authData}
	}
}

func verifyTokenCmd(br *bridge.Bridge, ctx context.Context, provider bridge.Provider, token string) tea.Cmd {
	return func() tea.Msg {
		// Store in keychain
		if err := keychain.Store(string(provider), token); err != nil {
			return verifyMsg{err: err}
		}

		// Verify by fetching config (will fail with INVALID_PARAMS if no project, but token is valid)
		_, err := br.FetchConfig(ctx, bridge.FetchConfigParams{
			Provider: provider,
			Token:    token,
		})

		// INVALID_PARAMS means token works, just no project specified
		if err != nil {
			if bridgeErr, ok := err.(*bridge.BridgeError); ok {
				if bridgeErr.Code == bridge.ErrInvalidParams {
					return verifyMsg{err: nil}
				}
			}
			// Delete token if verification failed
			keychain.Delete(string(provider))
			return verifyMsg{err: err}
		}

		return verifyMsg{err: nil}
	}
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return
	}
	cmd.Start()
}

// RunAuthTUI runs the interactive auth TUI
func RunAuthTUI(stateDB *state.DB, br *bridge.Bridge) error {
	p := tea.NewProgram(
		NewAuthModel(stateDB, br),
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		return err
	}

	// Return to dashboard
	return RunDashboardTUI(stateDB, br)
}
