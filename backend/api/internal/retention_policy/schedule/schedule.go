package schedule

import (
	"fmt"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
)

const DefaultTimezone = "Asia/Bangkok"

var parser = cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)

// Next returns the next execution instant in UTC. The cron expression is
// evaluated in the policy timezone so schedules such as "0 1 * * *" continue
// to mean 01:00 local time regardless of where the worker is running.
func Next(expression, timezone string, after time.Time) (time.Time, error) {
	expression = strings.TrimSpace(expression)
	if expression == "" {
		expression = "0 1 * * *"
	}
	timezone = strings.TrimSpace(timezone)
	if timezone == "" {
		timezone = DefaultTimezone
	}
	location, err := time.LoadLocation(timezone)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid schedule_timezone %q: %w", timezone, err)
	}
	value, err := parser.Parse(expression)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid cron_schedule: %w", err)
	}
	return value.Next(after.In(location)).UTC(), nil
}

func Location(timezone string) (*time.Location, error) {
	if strings.TrimSpace(timezone) == "" {
		timezone = DefaultTimezone
	}
	location, err := time.LoadLocation(timezone)
	if err != nil {
		return nil, fmt.Errorf("invalid schedule_timezone %q: %w", timezone, err)
	}
	return location, nil
}
