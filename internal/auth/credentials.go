package auth

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

// Credentials holds the application's OAuth credentials
type Credentials struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

//go:embed assets/credentials.json
var credentialsFile []byte

func loadCredentials() (*Credentials, error) {
	var creds Credentials
	if err := json.Unmarshal(credentialsFile, &creds); err != nil {
		return nil, fmt.Errorf("error parsing credentials: %w", err)
	}

	return &creds, nil
}
