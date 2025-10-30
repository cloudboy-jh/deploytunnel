package bridge

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

const (
	defaultTimeout = 30 * time.Second
	maxRetries     = 3
)

// Bridge manages communication with Bun adapters
type Bridge struct {
	adaptersPath string
	timeout      time.Duration
}

// NewBridge creates a new Bridge instance
func NewBridge(adaptersPath string) *Bridge {
	if adaptersPath == "" {
		// Default to ./adapters relative to binary
		execPath, _ := os.Executable()
		adaptersPath = filepath.Join(filepath.Dir(execPath), "..", "adapters")
	}

	return &Bridge{
		adaptersPath: adaptersPath,
		timeout:      defaultTimeout,
	}
}

// SetTimeout configures the command timeout
func (b *Bridge) SetTimeout(timeout time.Duration) {
	b.timeout = timeout
}

// Execute runs an adapter command and returns the parsed response
func (b *Bridge) Execute(ctx context.Context, provider Provider, verb string, params interface{}) (*Response, error) {
	adapterPath := filepath.Join(b.adaptersPath, string(provider), "index.ts")

	// Check if adapter exists
	if _, err := os.Stat(adapterPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("adapter not found: %s", provider)
	}

	// Marshal params to JSON
	var stdinData []byte
	var err error
	if params != nil {
		stdinData, err = json.Marshal(params)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal params: %w", err)
		}
	}

	// Create command with timeout context
	timeoutCtx, cancel := context.WithTimeout(ctx, b.timeout)
	defer cancel()

	cmd := exec.CommandContext(timeoutCtx, "bun", "run", adapterPath, verb)
	cmd.Stdin = bytes.NewReader(stdinData)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Execute command
	err = cmd.Run()
	if err != nil {
		if timeoutCtx.Err() == context.DeadlineExceeded {
			return nil, &BridgeError{
				Code:        ErrTimeout,
				Message:     fmt.Sprintf("adapter command timed out after %s", b.timeout),
				Recoverable: true,
			}
		}
		return nil, fmt.Errorf("adapter execution failed: %w (stderr: %s)", err, stderr.String())
	}

	// Parse response
	var response Response
	if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
		return nil, fmt.Errorf("failed to parse adapter response: %w (output: %s)", err, stdout.String())
	}

	// Check for error in response
	if !response.OK && response.Error != nil {
		return &response, response.Error
	}

	return &response, nil
}

// Capabilities fetches adapter capabilities
func (b *Bridge) Capabilities(ctx context.Context, provider Provider) (*CapabilitiesData, error) {
	resp, err := b.Execute(ctx, provider, "capabilities", nil)
	if err != nil {
		return nil, err
	}

	var caps CapabilitiesData
	if err := mapToStruct(resp.Data, &caps); err != nil {
		return nil, fmt.Errorf("failed to parse capabilities: %w", err)
	}

	return &caps, nil
}

// AuthStart initiates authentication flow
func (b *Bridge) AuthStart(ctx context.Context, params AuthStartParams) (*AuthStartData, error) {
	resp, err := b.Execute(ctx, params.Provider, "auth:start", params)
	if err != nil {
		return nil, err
	}

	var data AuthStartData
	if err := mapToStruct(resp.Data, &data); err != nil {
		return nil, fmt.Errorf("failed to parse auth data: %w", err)
	}

	return &data, nil
}

// FetchConfig retrieves project configuration
func (b *Bridge) FetchConfig(ctx context.Context, params FetchConfigParams) (*FetchConfigData, error) {
	resp, err := b.Execute(ctx, params.Provider, "fetch:config", params)
	if err != nil {
		return nil, err
	}

	var data FetchConfigData
	if err := mapToStruct(resp.Data, &data); err != nil {
		return nil, fmt.Errorf("failed to parse config data: %w", err)
	}

	return &data, nil
}

// SyncEnv synchronizes environment variables
func (b *Bridge) SyncEnv(ctx context.Context, params SyncEnvParams) (*SyncEnvData, error) {
	resp, err := b.Execute(ctx, params.Provider, "sync:env", params)
	if err != nil {
		return nil, err
	}

	var data SyncEnvData
	if err := mapToStruct(resp.Data, &data); err != nil {
		return nil, fmt.Errorf("failed to parse sync data: %w", err)
	}

	return &data, nil
}

// DeployPreview creates a preview deployment
func (b *Bridge) DeployPreview(ctx context.Context, params DeployPreviewParams) (*DeployPreviewData, error) {
	resp, err := b.Execute(ctx, params.Provider, "deploy:preview", params)
	if err != nil {
		return nil, err
	}

	var data DeployPreviewData
	if err := mapToStruct(resp.Data, &data); err != nil {
		return nil, fmt.Errorf("failed to parse deploy data: %w", err)
	}

	return &data, nil
}

// DnsUpdate updates a DNS record
func (b *Bridge) DnsUpdate(ctx context.Context, params DnsUpdateParams) (*DnsUpdateData, error) {
	resp, err := b.Execute(ctx, params.Provider, "dns:update", params)
	if err != nil {
		return nil, err
	}

	var data DnsUpdateData
	if err := mapToStruct(resp.Data, &data); err != nil {
		return nil, fmt.Errorf("failed to parse DNS update data: %w", err)
	}

	return &data, nil
}

// DnsRollback rolls back a DNS record
func (b *Bridge) DnsRollback(ctx context.Context, params DnsRollbackParams) (*DnsRollbackData, error) {
	resp, err := b.Execute(ctx, params.Provider, "dns:rollback", params)
	if err != nil {
		return nil, err
	}

	var data DnsRollbackData
	if err := mapToStruct(resp.Data, &data); err != nil {
		return nil, fmt.Errorf("failed to parse DNS rollback data: %w", err)
	}

	return &data, nil
}

// mapToStruct converts a map to a struct using JSON marshaling
func mapToStruct(m map[string]interface{}, v interface{}) error {
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}
