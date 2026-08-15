// Package deferral_test holds the black-box unit tests of the mt defer
// datetime parsing (Seam 2): absolute YY-MM-DD HH:MM (the year expanded
// to 20YY, the hour preserved), relative +<n><unit> durations computed
// from now, and the error edges. Only the exported Parse is exercised.
package deferral_test

import (
	"strings"
	"testing"
	"time"

	"github.com/Sanmoo/my-tasks2/internal/deferral"
)

// now is a fixed instant for the relative-form tests, in UTC so that
// Add across day/week boundaries cannot drift under a DST local zone.
var now = time.Date(2026, 8, 15, 14, 30, 0, 0, time.UTC)

func TestParseAbsolute(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"basic with hour", "26-08-20 08:00", "2026-08-20T08:00"},
		{"year 00 expands to 2000", "00-01-01 00:00", "2000-01-01T00:00"},
		{"year 68 stays in this century", "68-12-31 23:59", "2068-12-31T23:59"},
		{"year 69 expands to 2069", "69-01-01 00:00", "2069-01-01T00:00"},
		{"year 99 expands to 2099", "99-01-01 00:00", "2099-01-01T00:00"},
		{"zero-padded month and hour preserved", "26-01-01 00:00", "2026-01-01T00:00"},
		{"leap day", "24-02-29 12:00", "2024-02-29T12:00"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := deferral.Parse(c.in, now)
			if err != nil {
				t.Fatalf("Parse(%q) error: %v", c.in, err)
			}
			if got != c.want {
				t.Errorf("Parse(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}

func TestParseRelative(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"days", "+2d", "2026-08-17T14:30"},
		{"weeks", "+1w", "2026-08-22T14:30"},
		{"hours", "+3h", "2026-08-15T17:30"},
		{"one day", "+1d", "2026-08-16T14:30"},
		{"one hour", "+1h", "2026-08-15T15:30"},
		{"hour crossing midnight", "+10h", "2026-08-16T00:30"},
		{"multi-digit count", "+15d", "2026-08-30T14:30"},
		{"week crossing a month", "+3w", "2026-09-05T14:30"},
		{"nine days", "+9d", "2026-08-24T14:30"},
		{"nineteen hours", "+19h", "2026-08-16T09:30"},
		{"largest safe day count", "+106751d", "2318-11-24T14:30"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := deferral.Parse(c.in, now)
			if err != nil {
				t.Fatalf("Parse(%q) error: %v", c.in, err)
			}
			if got != c.want {
				t.Errorf("Parse(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}

func TestParseRelativeTruncatesSubMinute(t *testing.T) {
	// The stored value has minute granularity; seconds and nanos of now
	// are dropped, not rounded.
	withSeconds := time.Date(2026, 8, 15, 14, 30, 45, 123456789, time.UTC)
	got, err := deferral.Parse("+3h", withSeconds)
	if err != nil {
		t.Fatal(err)
	}
	if want := "2026-08-15T17:30"; got != want {
		t.Errorf("Parse(+3h, now with seconds) = %q, want %q (seconds truncated)", got, want)
	}
}

func TestParseRejectsMalformed(t *testing.T) {
	cases := []struct {
		name string
		in   string
	}{
		{"empty", ""},
		{"garbage", "garbage"},
		{"missing time", "26-08-20"},
		{"missing date", "08:00"},
		{"four-digit year", "2026-08-20 08:00"},
		{"single-digit month", "26-8-20 08:00"},
		{"month 13", "26-13-20 08:00"},
		{"day 32", "26-08-32 08:00"},
		{"hour 25", "26-08-20 25:00"},
		{"minute 60", "26-08-20 08:60"},
		{"wrong separators", "26/08/20 08:00"},
		{"zero duration", "+0d"},
		{"zero hours", "+0h"},
		{"negative duration", "+-2d"},
		{"unknown unit", "+2x"},
		{"uppercase unit", "+2D"},
		{"missing unit", "+2"},
		{"missing number", "+d"},
		{"only plus", "+"},
		{"no plus", "2d"},
		{"double plus", "++2d"},
		{"overflowing number", "+99999999999999999999d"},
		{"duration multiplication overflow", "+106752d"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got, err := deferral.Parse(c.in, now); err == nil {
				t.Errorf("Parse(%q) = %q, want an error", c.in, got)
			}
		})
	}
}

func TestParseErrorMessagesHintAtTheFormat(t *testing.T) {
	if _, err := deferral.Parse("bogus", now); err == nil || !strings.Contains(err.Error(), "YY-MM-DD HH:MM") {
		t.Errorf("absolute error = %v, want it to hint at YY-MM-DD HH:MM", err)
	}
	if _, err := deferral.Parse("+2x", now); err == nil || !strings.Contains(err.Error(), "d (days), w (weeks) or h (hours)") {
		t.Errorf("relative error = %v, want it to hint at d/w/h", err)
	}
}
