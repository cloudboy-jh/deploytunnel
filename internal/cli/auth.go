package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/johnhorton/deploy-tunnel/internal/bridge"
	"github.com/johnhorton/deploy-tunnel/internal/keychain"
	"github.com/johnhorton/deploy-tunnel/ui"
)

type AuthCommand struct {
	bridge *bridge.Bridge
}

func NewAuthCommand(br *bridge.Bridge) *AuthCommand {
	return &AuthCommand{
		bridge: br,
	}
}

func (c *AuthCommand) Run(ctx context.Context, provider string) error {
	fmt.Println(ui.Header())
	fmt.Println()

	prov := bridge.Provider(provider)

	// Check capabilities
	fmt.Println(ui.Info(fmt.Sprintf("Checking %s adapter capabilities...", provider)))
	caps, err := c.bridge.Capabilities(ctx, prov)
	if err != nil {
		return fmt.Errorf("failed to get capabilities: %w", err)
	}

	fmt.Println(ui.Success(fmt.Sprintf("Adapter: %s v%s", caps.AdapterName, caps.AdapterVersion)))
	fmt.Println(ui.KeyValue("Auth Type", caps.AuthType))
	fmt.Println()

	// Start auth flow
	fmt.Println(ui.Info("Starting authentication..."))
	authData, err := c.bridge.AuthStart(ctx, bridge.AuthStartParams{
		Provider: prov,
	})
	if err != nil {
		return fmt.Errorf("failed to start auth: %w", err)
	}

	var token string

	if authData.AuthURL != "" {
		// OAuth flow
		fmt.Println()
		fmt.Println(ui.Info("Opening browser for authentication..."))
		fmt.Println(ui.KeyValue("URL", authData.AuthURL))
		fmt.Println()

		if err := openBrowser(authData.AuthURL); err != nil {
			fmt.Println(ui.Warning("Failed to open browser automatically"))
			fmt.Println(ui.Info("Please visit the URL above manually"))
		}

		fmt.Println()
		fmt.Print(ui.KeyStyle.Render("? ") + "Paste the token from your browser: ")

		reader := bufio.NewReader(os.Stdin)
		token, err = reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read token: %w", err)
		}
		token = strings.TrimSpace(token)
	} else {
		// Direct token input
		fmt.Println()
		fmt.Println(ui.Info("This provider requires a personal access token"))
		fmt.Print(ui.KeyStyle.Render("? ") + "Enter your token: ")

		reader := bufio.NewReader(os.Stdin)
		token, err = reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read token: %w", err)
		}
		token = strings.TrimSpace(token)
	}

	if token == "" {
		return fmt.Errorf("token cannot be empty")
	}

	// Store token in keychain
	fmt.Println()
	fmt.Println(ui.Info("Storing credentials securely..."))
	if err := keychain.Store(provider, token); err != nil {
		return fmt.Errorf("failed to store token: %w", err)
	}

	// Verify token by fetching capabilities with it
	fmt.Println(ui.Info("Verifying credentials..."))
	_, err = c.bridge.FetchConfig(ctx, bridge.FetchConfigParams{
		Provider: prov,
		Token:    token,
	})
	if err != nil {
		// If fetch fails due to missing project_id, that's OK - token is valid
		if bridgeErr, ok := err.(*bridge.BridgeError); ok && bridgeErr.Code == bridge.ErrInvalidParams {
			fmt.Println(ui.Success("Authentication successful!"))
			fmt.Println()
			return nil
		}
		// Otherwise, token might be invalid
		return fmt.Errorf("failed to verify token: %w", err)
	}

	fmt.Println(ui.Success("Authentication successful!"))
	fmt.Println()
	fmt.Println(ui.Info("Your credentials have been securely stored in the system keychain"))
	fmt.Println()

	return nil
}

func (c *AuthCommand) List() error {
	fmt.Println(ui.Header())
	fmt.Println()
	fmt.Println(ui.Info("Stored credentials:"))
	fmt.Println()

	providers, err := keychain.List()
	if err != nil {
		return fmt.Errorf("failed to list credentials: %w", err)
	}

	if len(providers) == 0 {
		fmt.Println(ui.Warning("No credentials stored"))
		fmt.Println()
		fmt.Println(ui.Info("Run: dt auth <provider>"))
		fmt.Println()
		return nil
	}

	for _, provider := range providers {
		fmt.Println(ui.Success(provider))
	}
	fmt.Println()

	return nil
}

func (c *AuthCommand) Revoke(provider string) error {
	fmt.Println(ui.Header())
	fmt.Println()

	if err := keychain.Delete(provider); err != nil {
		return fmt.Errorf("failed to delete credentials: %w", err)
	}

	fmt.Println(ui.Success(fmt.Sprintf("Credentials for %s have been removed", provider)))
	fmt.Println()

	return nil
}

// openBrowser opens a URL in the system's default browser
func openBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return fmt.Errorf("unsupported platform")
	}

	return cmd.Start()
}
