package handlers

import (
	"testing"
	"time"
)

func FuzzParseHabitLogDate(f *testing.F) {
	f.Add("")
	f.Add("2026-06-22")
	f.Add("22.06.2026")
	f.Add("invalid-date")

	f.Fuzz(func(t *testing.T, date string) {
		parsed, err := parseHabitLogDate(date)
		if err != nil {
			return
		}

		if _, err := time.Parse("2006-01-02", parsed); err != nil {
			t.Fatalf("accepted invalid date %q: %v", parsed, err)
		}
	})
}

func FuzzValidateHabitLogsQuery(f *testing.F) {
	f.Add(1, 1, 2026)
	f.Add(0, 1, 2026)
	f.Add(1, 13, 2026)
	f.Add(1, 1, 0)

	f.Fuzz(func(t *testing.T, habitID int, month int, year int) {
		valid := isValidHabitLogsQuery(habitID, month, year)
		expected := habitID > 0 && month >= 1 && month <= 12 && year > 0
		if valid != expected {
			t.Fatalf("unexpected validation result for habitID=%d month=%d year=%d", habitID, month, year)
		}
	})
}

func FuzzValidateHexColor(f *testing.F) {
	f.Add("#4caf50")
	f.Add("#FFFFFF")
	f.Add("4caf50")
	f.Add("#xyzxyz")

	f.Fuzz(func(t *testing.T, color string) {
		valid := isValidHexColor(color)
		if valid && len(color) != 7 {
			t.Fatalf("accepted color with invalid length: %q", color)
		}
		if valid && color[0] != '#' {
			t.Fatalf("accepted color without #: %q", color)
		}
	})
}
