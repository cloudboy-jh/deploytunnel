package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/johnhorton/deploy-tunnel/internal/bridge"
	"github.com/johnhorton/deploy-tunnel/internal/keychain"
	"github.com/johnhorton/deploy-tunnel/internal/state"
	"github.com/johnhorton/deploy-tunnel/ui"
)

type InitCommand struct {
	state  *state.DB
	bridge *bridge.Bridge
}

func NewInitCommand(stateDB *state.DB, br *bridge.Bridge) *InitCommand {
	return &InitCommand{
		state:  stateDB,
		bridge: br,
	}
}

func (c *InitCommand) Run(ctx context.Context) error {
	fmt.Println(ui.Header())
	fmt.Println()
	fmt.Println(ui.Info("Let's set up your migration"))
	fmt.Println()

	// Select source provider
	source, err := c.selectProvider("Source provider (where you're migrating FROM)")
	if err != nil {
		return fmt.Errorf("failed to select source provider: %w", err)
	}

	// Select target provider
	target, err := c.selectProvider("Target provider (where you're migrating TO)")
	if err != nil {
		return fmt.Errorf("failed to select target provider: %w", err)
	}

	if source == target {
		fmt.Println()
		fmt.Println(ui.Warning("Source and target providers are the same. This is unusual but allowed."))
		fmt.Println()
	}

	// Prompt for domain
	domain, err := c.promptString("Domain name to migrate")
	if err != nil {
		return fmt.Errorf("failed to get domain: %w", err)
	}

	fmt.Println()
	fmt.Println(ui.Info("Creating migration configuration..."))

	// Create migration record
	migrationID := uuid.New().String()
	if err := c.state.CreateMigration(migrationID, string(source), string(target), domain); err != nil {
		return fmt.Errorf("failed to create migration: %w", err)
	}

	fmt.Println(ui.Success("Migration initialized"))
	fmt.Println()
	fmt.Println(ui.KeyValue("Migration ID", migrationID))
	fmt.Println(ui.KeyValue("Source", string(source)))
	fmt.Println(ui.KeyValue("Target", string(target)))
	fmt.Println(ui.KeyValue("Domain", domain))
	fmt.Println()

	// Check authentication
	fmt.Println(ui.Info("Checking authentication status..."))
	fmt.Println()

	sourceAuth, _ := keychain.Get(string(source))
	targetAuth, _ := keychain.Get(string(target))

	if sourceAuth == "" {
		fmt.Println(ui.Warning(fmt.Sprintf("No credentials found for %s", source)))
		fmt.Println(ui.Info(fmt.Sprintf("Run: dt auth %s", source)))
	} else {
		fmt.Println(ui.Success(fmt.Sprintf("%s is authenticated", source)))
	}

	if targetAuth == "" {
		fmt.Println(ui.Warning(fmt.Sprintf("No credentials found for %s", target)))
		fmt.Println(ui.Info(fmt.Sprintf("Run: dt auth %s", target)))
	} else {
		fmt.Println(ui.Success(fmt.Sprintf("%s is authenticated", target)))
	}

	fmt.Println()
	fmt.Println(ui.Info("Next steps:"))
	fmt.Println(ui.List([]string{
		fmt.Sprintf("Authenticate providers: dt auth %s && dt auth %s", source, target),
		"Fetch source configuration: dt fetch:config",
		"Sync environment variables: dt sync env",
		"Create preview tunnel: dt tunnel create --preview",
		"Verify routes: dt verify",
		"Cutover when ready: dt cutover",
	}))
	fmt.Println()

	return nil
}

func (c *InitCommand) selectProvider(prompt string) (bridge.Provider, error) {
	providers := []bridge.Provider{
		bridge.ProviderVercel,
		bridge.ProviderCloudflare,
		bridge.ProviderRender,
		bridge.ProviderNetlify,
	}

	options := make([]string, len(providers))
	for i, p := range providers {
		options[i] = string(p)
	}

	fmt.Println(ui.Select(prompt, options))

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	input = strings.TrimSpace(input)
	choice, err := strconv.Atoi(input)
	if err != nil || choice < 1 || choice > len(providers) {
		return "", fmt.Errorf("invalid choice: must be 1-%d", len(providers))
	}

	return providers[choice-1], nil
}

func (c *InitCommand) promptString(prompt string) (string, error) {
	fmt.Printf("%s %s: ", ui.KeyStyle.Render("?"), prompt)

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(input), nil
}
