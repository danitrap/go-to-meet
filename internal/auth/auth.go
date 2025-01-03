package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/danitrap/go-to-meet/pkg/browser"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
)

const (
	appName   = "go-to-meet"
	tokenFile = "token.json"
)

func Setup() (*oauth2.Config, error) {
	creds, err := loadCredentials()
	if err != nil {
		return nil, fmt.Errorf("unable to load credentials: %w", err)
	}

	return &oauth2.Config{
		ClientID:     creds.ClientID,
		ClientSecret: creds.ClientSecret,
		Endpoint:     google.Endpoint,
		RedirectURL:  "http://localhost:8080/callback",
		Scopes: []string{
			calendar.CalendarReadonlyScope,
		},
	}, nil
}

func GetTokenSource(config *oauth2.Config) (oauth2.TokenSource, error) {
	token, err := loadToken()
	if err != nil {
		// If no token exists, start OAuth flow
		token = getTokenFromWeb(config)
		if err := saveToken(token); err != nil {
			return nil, fmt.Errorf("failed to save token: %w", err)
		}
	}

	// Create a token source that automatically refreshes
	tokenSource := config.TokenSource(context.Background(), token)

	// Verify token is valid or can be refreshed
	if _, err := tokenSource.Token(); err != nil {
		log.Println("Stored token is invalid or cannot be refreshed, getting new token...")
		token = getTokenFromWeb(config)
		if err := saveToken(token); err != nil {
			return nil, fmt.Errorf("failed to save new token: %w", err)
		}
		tokenSource = config.TokenSource(context.Background(), token)
	}

	return tokenSource, nil
}

func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	codeChan := make(chan string)
	server := &http.Server{Addr: ":8080"}

	// Create context with timeout for server shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		codeChan <- code
		fmt.Fprintf(w, "Authorization successful! You can close this window.")
		go func() {
			time.Sleep(time.Second)
			if err := server.Shutdown(ctx); err != nil {
				log.Printf("Error shutting down server: %v", err)
			}
		}()
	})

	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	authURL := config.AuthCodeURL("state")
	fmt.Printf("Visit this URL to authorize the application:\n%v\n", authURL)
	browser.Open(authURL)

	code := <-codeChan

	token, err := config.Exchange(ctx, code)
	if err != nil {
		log.Fatalf("Unable to exchange authorization code: %v", err)
	}

	return token
}

func getAppDataDir() (string, error) {
	userConfigDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	appDir := filepath.Join(userConfigDir, "Library", "Application Support", appName)
	if err := os.MkdirAll(appDir, 0700); err != nil {
		return "", fmt.Errorf("failed to create app directory: %w", err)
	}

	return appDir, nil
}

func getTokenPath() (string, error) {
	appDir, err := getAppDataDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(appDir, tokenFile), nil
}

func loadToken() (*oauth2.Token, error) {
	tokenPath, err := getTokenPath()
	if err != nil {
		return nil, err
	}

	f, err := os.Open(tokenPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	token := &oauth2.Token{}
	if err := json.NewDecoder(f).Decode(token); err != nil {
		return nil, fmt.Errorf("failed to decode token: %w", err)
	}
	return token, nil
}

func saveToken(token *oauth2.Token) error {
	tokenPath, err := getTokenPath()
	if err != nil {
		return err
	}

	f, err := os.Create(tokenPath)
	if err != nil {
		return fmt.Errorf("unable to create token file: %w", err)
	}
	defer f.Close()

	if err := json.NewEncoder(f).Encode(token); err != nil {
		return fmt.Errorf("failed to encode token: %w", err)
	}
	return nil
}
