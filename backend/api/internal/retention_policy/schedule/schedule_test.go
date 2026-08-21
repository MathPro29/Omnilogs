package schedule

import (
	"testing"
	"time"
)

func TestNextUsesPolicyTimezone(t *testing.T) {
	after := time.Date(2026, 8, 20, 16, 0, 0, 0, time.UTC)
	got, err := Next("0 1 * * *", "Asia/Bangkok", after)
	if err != nil {
		t.Fatal(err)
	}
	want := time.Date(2026, 8, 20, 18, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Fatalf("next run = %s, want %s", got, want)
	}
}

func TestNextRejectsInvalidValues(t *testing.T) {
	if _, err := Next("not-a-cron", "Asia/Bangkok", time.Now()); err == nil {
		t.Fatal("expected invalid cron error")
	}
	if _, err := Next("0 1 * * *", "Mars/Olympus", time.Now()); err == nil {
		t.Fatal("expected invalid timezone error")
	}
}
