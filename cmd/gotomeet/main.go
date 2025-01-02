package main

import (
	"log"

	"github.com/danitrap/go-to-meet/internal/auth"
	"github.com/danitrap/go-to-meet/internal/calendar"
	"github.com/danitrap/go-to-meet/internal/ui"
	"github.com/danitrap/go-to-meet/pkg/models"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	app := ui.NewApp()
	updateCh := make(chan []models.Meeting)

	go func() {
		config, err := auth.Setup()
		if err != nil {
			log.Fatalf("Failed to setup OAuth config: %v", err)
		}
		tokenSource, err := auth.GetTokenSource(config)
		if err != nil {
			log.Fatalf("Failed to get token source: %v", err)
		}
		calendarService, err := calendar.NewService(tokenSource)
		if err != nil {
			log.Fatalf("Failed to create calendar service: %v", err)
		}

		go calendarService.StartMeetingChecker(updateCh)

		for meetings := range updateCh {
			log.Println("Updating meetings")
			app.UpdateMeetings(meetings)
		}
	}()

	app.RunApplication()
}
