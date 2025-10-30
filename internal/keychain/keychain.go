package keychain

import (
	"fmt"

	"github.com/zalando/go-keyring"
)

const (
	serviceName = "deploy-tunnel"
)

// Store stores a credential in the system keychain
func Store(provider, token string) error {
	key := fmt.Sprintf("%s-token", provider)
	return keyring.Set(serviceName, key, token)
}

// Get retrieves a credential from the system keychain
func Get(provider string) (string, error) {
	key := fmt.Sprintf("%s-token", provider)
	token, err := keyring.Get(serviceName, key)
	if err == keyring.ErrNotFound {
		return "", fmt.Errorf("no credentials found for %s", provider)
	}
	return token, err
}

// Delete removes a credential from the system keychain
func Delete(provider string) error {
	key := fmt.Sprintf("%s-token", provider)
	return keyring.Delete(serviceName, key)
}

// List returns all stored provider keys
func List() ([]string, error) {
	// Note: keyring doesn't provide a list function, so we'll try common providers
	providers := []string{"vercel", "cloudflare", "render", "netlify"}
	var found []string

	for _, provider := range providers {
		if _, err := Get(provider); err == nil {
			found = append(found, provider)
		}
	}

	return found, nil
}

// StoreRefreshToken stores a refresh token
func StoreRefreshToken(provider, token string) error {
	key := fmt.Sprintf("%s-refresh-token", provider)
	return keyring.Set(serviceName, key, token)
}

// GetRefreshToken retrieves a refresh token
func GetRefreshToken(provider string) (string, error) {
	key := fmt.Sprintf("%s-refresh-token", provider)
	token, err := keyring.Get(serviceName, key)
	if err == keyring.ErrNotFound {
		return "", fmt.Errorf("no refresh token found for %s", provider)
	}
	return token, err
}
