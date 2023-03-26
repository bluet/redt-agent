package utils

import (
	"time"
)

// CurrentTimestamp returns the current time as a formatted string.
func CurrentTimestamp() string {
	return time.Now().Format("2006-01-02 15:04:05")
}
