package auth

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// AuthConfig represents the authentication configuration stored in file
type AuthConfig struct {
	JWT      string `json:"jwt"`
	Username string `json:"username,omitempty"`
}

// getConfigPath returns the path to the config file
func getConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	configDir := filepath.Join(homeDir, ".data-vault")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return "", err
	}

	return filepath.Join(configDir, "auth.json"), nil
}

// SaveJWT saves JWT token to config file
func SaveJWT(jwt, username string) error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	config := AuthConfig{
		JWT:      jwt,
		Username: username,
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0600) // 0600 = read/write for owner only
}

// LoadJWT loads JWT token from config file
func LoadJWT() (string, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil // No saved JWT
		}
		return "", err
	}

	var config AuthConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return "", err
	}

	return config.JWT, nil
}

// ClearJWT removes saved JWT (for logout)
func ClearJWT() error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	if err := os.Remove(configPath); err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}
