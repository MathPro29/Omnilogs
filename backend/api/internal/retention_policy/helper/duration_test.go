package helper

import (
	"testing"
	"time"
)

func TestAddUsesCalendarMonthsAndYears(t *testing.T) {
	loc := time.UTC
	date := time.Date(2024, 1, 31, 12, 0, 0, 0, loc)
	got, err := Add(date, 1, "MONTH")
	if err != nil {
		t.Fatal(err)
	}
	if !got.Equal(time.Date(2024, 3, 2, 12, 0, 0, 0, loc)) {
		t.Fatalf("calendar month = %s", got)
	}
	year, err := Add(time.Date(2024, 2, 29, 0, 0, 0, 0, loc), 1, "YEAR")
	if err != nil {
		t.Fatal(err)
	}
	if !year.Equal(time.Date(2025, 3, 1, 0, 0, 0, 0, loc)) {
		t.Fatalf("calendar year = %s", year)
	}
}

func TestNormalizeUnitAcceptsLegacyPluralValues(t *testing.T) {
	for input, expected := range map[string]string{"DAYS": UnitDay, "WEEKS": UnitWeek, "MONTHS": UnitMonth, "YEARS": UnitYear} {
		got, err := NormalizeUnit(input)
		if err != nil || got != expected {
			t.Fatalf("NormalizeUnit(%q) = %q, %v", input, got, err)
		}
	}
}
