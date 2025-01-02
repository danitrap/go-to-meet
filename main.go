package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"sort"
	"time"

	"github.com/caseymrm/menuet"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

// Credentials holds the application's OAuth credentials
type Credentials struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

type Meeting struct {
	Summary   string
	StartTime time.Time
	MeetLink  string
}

const (
	tokenFile      = "token.json"
	credentialFile = "credentials.json"
	maxRetries     = 3
	retryDelay     = 5 * time.Second
)

var (
	calendarService *calendar.Service
	currentMeetings []Meeting
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Load credentials from config file
	config, err := setupOAuthConfig()
	if err != nil {
		log.Fatalf("Failed to setup OAuth config: %v", err)
	}

	// Create a token source with automatic refresh
	tokenSource, err := getTokenSource(config)
	if err != nil {
		log.Fatalf("Failed to get token source: %v", err)
	}

	// Create Calendar service with auto-refreshing client
	calendarService, err = createCalendarService(tokenSource)
	if err != nil {
		log.Fatalf("Failed to create calendar service: %v", err)
	}

	// Set up menuet app
	app := menuet.App()
	app.SetMenuState(&menuet.MenuState{
		Title: "📅",
	})

	// Update meetings every minute
	go func() {
		for {
			meetings := checkUpcomingMeetings(calendarService)
			if isAuthError(err) {
				log.Println("Attempting to recreate calendar service...")
				if calendarService, err = recreateService(config); err != nil {
					log.Printf("Failed to recreate service: %v", err)
					time.Sleep(retryDelay)
					continue
				}
			}
			currentMeetings = meetings
			updateMenuDisplay(app)
			time.Sleep(30 * time.Second)
		}
	}()

	app.Label = "dev.trappi.meetbar"

	app.Children = menuItemsFromMeetings

	app.RunApplication()
}

func setupOAuthConfig() (*oauth2.Config, error) {
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

func getTokenSource(config *oauth2.Config) (oauth2.TokenSource, error) {
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

func createCalendarService(tokenSource oauth2.TokenSource) (*calendar.Service, error) {
	client := oauth2.NewClient(context.Background(), tokenSource)

	var srv *calendar.Service
	var err error

	// Retry logic for service creation
	for i := 0; i < maxRetries; i++ {
		srv, err = calendar.NewService(context.Background(), option.WithHTTPClient(client))
		if err == nil {
			return srv, nil
		}
		log.Printf("Attempt %d: Failed to create calendar service: %v", i+1, err)
		time.Sleep(retryDelay)
	}

	return nil, fmt.Errorf("failed to create calendar service after %d attempts: %w", maxRetries, err)
}

func recreateService(config *oauth2.Config) (*calendar.Service, error) {
	// Force new token creation
	if err := os.Remove(tokenFile); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to remove old token: %w", err)
	}

	tokenSource, err := getTokenSource(config)
	if err != nil {
		return nil, fmt.Errorf("failed to get new token source: %w", err)
	}

	return createCalendarService(tokenSource)
}

func isAuthError(err error) bool {
	return err != nil && (
	// Add specific error type checks here
	err.Error() == "oauth2: token expired and refresh token is not available" ||
		err.Error() == "oauth2: cannot fetch token: 401 Unauthorized" ||
		err.Error() == "oauth2: token expired and refresh failed")
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

func checkUpcomingMeetings(srv *calendar.Service) []Meeting {
	var meets []Meeting
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	now := time.Now()
	thirtyMinutesFromNow := now.Add(30 * time.Minute)

	events, err := srv.Events.List("primary").
		TimeMin(now.Format(time.RFC3339)).
		TimeMax(thirtyMinutesFromNow.Format(time.RFC3339)).
		SingleEvents(true).
		OrderBy("startTime").
		Context(ctx).
		Do()
	if err != nil {
		log.Printf("Failed to retrieve events: %v", err)
		return meets
	}

	for _, event := range events.Items {
		if event.HangoutLink != "" {
			startTime, err := time.Parse(time.RFC3339, event.Start.DateTime)
			if err != nil {
				log.Printf("Error parsing start time for event %s: %v", event.Summary, err)
				continue
			}
			meets = append(meets, Meeting{
				Summary:   event.Summary,
				StartTime: startTime,
				MeetLink:  event.HangoutLink,
			})
		}
	}

	sort.Slice(meets, func(i, j int) bool {
		return meets[i].StartTime.Before(meets[j].StartTime)
	})

	return meets
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

func menuItemsFromMeetings() []menuet.MenuItem {
	var items []menuet.MenuItem
	if len(currentMeetings) == 0 {
		items = append(items, menuet.MenuItem{
			Text: "No upcoming meetings",
		})
		return items
	} else {
		for _, meet := range currentMeetings {
			meet := meet // Create new variable for closure
			items = append(items, menuet.MenuItem{
				Text: fmt.Sprintf("%s (%s)",
					meet.Summary,
					meet.StartTime.Format("15:04")),
				Clicked: func() {
					openMeetLink(meet.MeetLink)
				},
			})
		}
	}

	// Add separator and quit option
	items = append(items,
		menuet.MenuItem{Type: menuet.Separator},
		menuet.MenuItem{
			Text: "Quit",
			Clicked: func() {
				os.Exit(0)
			},
		},
	)

	return items
}

func updateMenuDisplay(app *menuet.Application) {
	if len(currentMeetings) == 0 {
		app.SetMenuState(&menuet.MenuState{
			Title: "📅",
		})
	} else {
		nextMeeting := currentMeetings[0]
		timeUntil := time.Until(nextMeeting.StartTime)

		var displayTime string
		if timeUntil < 0 {
			displayTime = "Now"
		} else {
			displayTime = nextMeeting.StartTime.Format("15:04")
		}

		// Update menu bar title
		app.SetMenuState(&menuet.MenuState{
			Title: fmt.Sprintf("📅 %s", displayTime),
		})

	}
}

func openMeetLink(link string) {
	cmd := fmt.Sprintf("open '%s'", link)
	err := exec.Command("bash", "-c", cmd).Run()
	if err != nil {
		log.Printf("Error opening meet link: %v", err)
	}
}
