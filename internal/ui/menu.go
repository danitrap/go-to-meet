package ui

import (
	"fmt"
	"log"
	"time"

	"github.com/caseymrm/menuet"
	"github.com/danitrap/go-to-meet/pkg/browser"
	"github.com/danitrap/go-to-meet/pkg/models"
)

var icons = map[string]string{
	"empty":   "🧘",
	"default": "📅",
	"soon":    "⏰",
	"now":     "🗣️",
}

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

func getIconForMeeting(meet models.Meeting) string {
	timeUntil := time.Until(meet.StartTime)

	if timeUntil <= 0 {
		return icons["now"]
	}

	if timeUntil < 2*time.Minute {
		return icons["soon"]
	}
	return icons["default"]
}

func formatDuration(d time.Duration) string {
	if d < 0 {
		return "0m"
	}

	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}

	return fmt.Sprintf("%dm", minutes)
}

func formatTime(t time.Time) string {
	d := time.Until(t)
	if d < 0 {
		return "now"
	}
	return "in " + formatDuration(d)
}

func (a *App) UpdateMenuDisplay() {
	if len(a.meetings) == 0 {
		a.App.SetMenuState(&menuet.MenuState{
			Title: icons["empty"],
		})
	} else {
		nextMeeting := a.meetings[0]

		icon := getIconForMeeting(nextMeeting)

		displayTime := formatTime(nextMeeting.StartTime)

		title := fmt.Sprintf("%s %s", icon, displayTime)

		// Update menu bar title
		a.App.SetMenuState(&menuet.MenuState{
			Title: title,
		})

	}
}
