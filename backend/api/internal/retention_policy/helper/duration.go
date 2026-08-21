package helper

import (
	"fmt"
	"strings"
	"time"
)

const (
	UnitDay   = "DAY"
	UnitWeek  = "WEEK"
	UnitMonth = "MONTH"
	UnitYear  = "YEAR"
)

// NormalizeUnit accepts the singular API contract and the plural legacy
// values. Calendar units are deliberately not converted to a day count.
func NormalizeUnit(value string) (string, error) {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case "DAY", "DAYS", "DAILY":
		return UnitDay, nil
	case "WEEK", "WEEKS", "WEEKLY":
		return UnitWeek, nil
	case "MONTH", "MONTHS", "MONTHLY":
		return UnitMonth, nil
	case "YEAR", "YEARS", "YEARLY":
		return UnitYear, nil
	default:
		return "", fmt.Errorf("invalid retention unit %q: allowed values are DAY, WEEK, MONTH, YEAR", value)
	}
}

func Validate(value int, unit string, allowZero bool) error {
	if value < 0 || (!allowZero && value == 0) {
		return fmt.Errorf("retention value must be greater than zero")
	}
	_, err := NormalizeUnit(unit)
	return err
}

// Add applies a retention duration using calendar arithmetic. For example,
// Jan 31 + 1 MONTH follows time.AddDate's calendar semantics and is never
// treated as 30 days.
func Add(value time.Time, amount int, unit string) (time.Time, error) {
	normalized, err := NormalizeUnit(unit)
	if err != nil {
		return time.Time{}, err
	}
	switch normalized {
	case UnitDay:
		return value.AddDate(0, 0, amount), nil
	case UnitWeek:
		return value.AddDate(0, 0, amount*7), nil
	case UnitMonth:
		return value.AddDate(0, amount, 0), nil
	case UnitYear:
		return value.AddDate(amount, 0, 0), nil
	default:
		return time.Time{}, fmt.Errorf("unsupported retention unit %q", unit)
	}
}

func Subtract(value time.Time, amount int, unit string) (time.Time, error) {
	return Add(value, -amount, unit)
}
