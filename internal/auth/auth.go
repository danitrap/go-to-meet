package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
)

const (
	credentialFile = "credentials.json"
	tokenFile      = "token.json"
)

func Setup() (*oauth2.Config, error) {
	creds, err := loadCredentials(credentialFile)
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
	token, err := loadToken(tokenFile)
	if err != nil {
		// If no token exists, start OAuth flow
		token = getTokenFromWeb(config)
		if err := saveToken(tokenFile, token); err != nil {
			return nil, fmt.Errorf("failed to save token: %w", err)
		}
	}

	// Create a token source that automatically refreshes
	tokenSource := config.TokenSource(context.Background(), token)

	// Verify token is valid or can be refreshed
	if _, err := tokenSource.Token(); err != nil {
		log.Println("Stored token is invalid or cannot be refreshed, getting new token...")
		token = getTokenFromWeb(config)
		if err := saveToken(tokenFile, token); err != nil {
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

	code := <-codeChan

	token, err := config.Exchange(ctx, code)
	if err != nil {
		log.Fatalf("Unable to exchange authorization code: %v", err)
	}

	return token
}

func loadToken(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
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

func saveToken(file string, token *oauth2.Token) error {
	f, err := os.Create(file)
	if err != nil {
		return fmt.Errorf("unable to create token file: %w", err)
	}
	defer f.Close()

	if err := json.NewEncoder(f).Encode(token); err != nil {
		return fmt.Errorf("failed to encode token: %w", err)
	}
	return nil
}
