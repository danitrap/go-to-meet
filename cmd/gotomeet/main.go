package main

import (
	"log"
	"time"

	"github.com/danitrap/go-to-meet/internal/auth"
	"github.com/danitrap/go-to-meet/internal/calendar"
	"github.com/danitrap/go-to-meet/internal/ui"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	app := ui.NewApp()
	tickCh := time.Tick(1 * time.Second)

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

		go calendarService.StartMeetingChecker()

		for range tickCh {
			app.UpdateMeetings(calendarService.GetMeetings())
		}
	}()

	app.RunApplication()
}
