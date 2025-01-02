package calendar

import (
	"context"
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/danitrap/go-to-meet/pkg/models"
	"golang.org/x/oauth2"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

const (
	tokenFile  = "token.json"
	maxRetries = 3
	retryDelay = 5 * time.Second
)

type CalendarService struct {
	service calendar.Service
}

func NewService(tokenSource oauth2.TokenSource) (*CalendarService, error) {
	srv, err := createCalendarService(tokenSource)
	if err != nil {
		return nil, fmt.Errorf("failed to create calendar service: %w", err)
	}

	return &CalendarService{
		service: *srv,
	}, nil
}

func (s *CalendarService) StartMeetingChecker(updateCh chan []models.Meeting) {
	for {
		meetings := checkUpcomingMeetings(&s.service)
		log.Printf("Found %d upcoming meetings", len(meetings))
		updateCh <- meetings
		time.Sleep(30 * time.Second)
	}
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

func checkUpcomingMeetings(srv *calendar.Service) []models.Meeting {
	var meets []models.Meeting
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	now := time.Now()
	endOfDay := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, now.Location())

	events, err := srv.Events.List("primary").
		TimeMin(now.Format(time.RFC3339)).
		TimeMax(endOfDay.Format(time.RFC3339)).
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
			meets = append(meets, models.Meeting{
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
