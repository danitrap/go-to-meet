package ui

import (
	"fmt"
	"log"
	"time"

	"github.com/caseymrm/menuet"
	"github.com/danitrap/go-to-meet/pkg/browser"
	"github.com/danitrap/go-to-meet/pkg/models"
)

type App struct {
	App      *menuet.Application
	meetings []models.Meeting
}

func NewApp() *App {
	app := menuet.App()

	a := &App{
		App: app,
	}

	app.SetMenuState(&menuet.MenuState{
		Title: "📅",
	})

	app.Label = "dev.trappi.go-to-meet"

	app.Children = func() []menuet.MenuItem {
		return menuItemsFromMeetings(a.meetings)
	}
	log.Println("Menu items set")

	return a
}

func (a *App) RunApplication() {
	a.App.RunApplication()
}

func (a *App) UpdateMeetings(meetings []models.Meeting) {
	a.meetings = meetings
	a.UpdateMenuDisplay()
}

func menuItemsFromMeetings(meetings []models.Meeting) []menuet.MenuItem {
	var items []menuet.MenuItem
	if len(meetings) == 0 {
		items = append(items, menuet.MenuItem{
			Text: "No upcoming meetings",
		})
		return items
	} else {
		for _, meet := range meetings {
			meet := meet // Create new variable for closure
			items = append(items, menuet.MenuItem{
				Text: fmt.Sprintf("%s (%s)",
					meet.Summary,
					meet.StartTime.Format("15:04")),
				Clicked: func() {
					browser.Open(meet.MeetLink)
				},
			})
		}
	}

	return items
}

func (a *App) UpdateMenuDisplay() {
	if len(a.meetings) == 0 {
		a.App.SetMenuState(&menuet.MenuState{
			Title: "📅",
		})
	} else {
		nextMeeting := a.meetings[0]
		timeUntil := time.Until(nextMeeting.StartTime)

		var displayTime string
		if timeUntil < 0 {
			displayTime = "Now"
		} else {
			displayTime = nextMeeting.StartTime.Format("15:04")
		}

		// Update menu bar title
		a.App.SetMenuState(&menuet.MenuState{
			Title: fmt.Sprintf("📅 %s", displayTime),
		})

	}
}
