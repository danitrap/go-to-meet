package models

import "time"

type Meeting struct {
	Summary   string
	StartTime time.Time
	EndTime   time.Time
	MeetLink  string
}
