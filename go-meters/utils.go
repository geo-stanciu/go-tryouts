package main

import "time"

func inTimeSpan(start, end, check time.Time) bool {
	return check.After(start) && check.Before(end)
}
