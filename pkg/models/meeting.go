package models

import "time"

type Meeting struct {
	Summary   string
	StartTime time.Time
	MeetLink  string
}
