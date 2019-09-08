package helpers

import (
	"testing"
	"time"
)

func TestBeginOfDay(t *testing.T) {
	datetime := time.Now()
	date := BeginOfDay(datetime)

	if date.Hour() != 0 {
		t.Error("hour must be 0")
	}

	if date.Minute() != 0 {
		t.Error("minute must be 0")
	}

	if date.Second() != 0 {
		t.Error("seconds must be 0")
	}
}

func TestEndOfDay(t *testing.T) {
	datetime := time.Now()
	date := EndOfDay(datetime)

	if date.Hour() != 23 {
		t.Error("hour must be 23")
	}

	if date.Minute() != 59 {
		t.Error("minute must be 59")
	}

	if date.Second() != 59 {
		t.Error("seconds must be 59")
	}
}
