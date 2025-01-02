package auth

import (
	"encoding/json"
	"fmt"
	"os"
)

// Credentials holds the application's OAuth credentials
type Credentials struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

func loadCredentials(filename string) (*Credentials, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading credentials file: %w", err)
	}

	var creds Credentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, fmt.Errorf("error parsing credentials: %w", err)
	}

	return &creds, nil
}
