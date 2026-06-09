package jobs

import "time"

type Result struct {
	JobName    string
	StartedAt  time.Time
	FinishedAt time.Time
	Scanned    int
	Matched    int
	Published  int
	Deleted    int
	Skipped    int
	Errors     int
}
