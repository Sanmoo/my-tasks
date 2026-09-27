// Package deferral holds the pure logic of `mt defer`: parsing the time
// argument — an absolute "YY-MM-DD HH:MM" or a relative "+<n><unit>" —
// into the canonical stored deferred_until value (issue.NaiveLayout). It
// is decision-dense, so it lives at Seam 2: black-box unit tested, with
// the coverage and mutation gates.
package deferral

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Sanmoo/my-tasks2/internal/issue"
)

// absoluteLayout is the input layout of `mt defer`'s absolute form: a
// two-digit year with zero-padded month, day, hour and minute.
const absoluteLayout = "06-01-02 15:04"

// maxDuration is the largest duration time.Time.Add can represent without
// overflowing its nanosecond count. Relative counts are checked against it
// before multiplying by their unit duration.
const maxDuration = time.Duration(1<<63 - 1)

// maxYear is the largest year NaiveLayout can round-trip: its "2006"
// token parses exactly four digits, so a calendar target must stay within
// 0000-9999.
const maxYear = 9999

// unitList is the catalogue of relative units, shared by the parse error
// messages so the accepted set is documented in one place.
const unitList = "d (days), w (weeks), h (hours), m (months) or y (years)"

// Parse converts a `mt defer` time argument to the canonical
// deferred_until value (issue.NaiveLayout, YYYY-MM-DDTHH:MM naive local
// time). It accepts:
//
//   - an absolute "YY-MM-DD HH:MM"; the two-digit year is expanded to
//     20YY, so any year this century is reachable and the hour is kept;
//   - a relative "+<n><unit>" computed from now, where unit is d (days),
//     w (weeks), h (hours), m (months) or y (years) and n is a positive
//     integer. Duration units (d/w/h) add an exact duration; calendar
//     units (m/y) add calendar months/years, clamping the day to the
//     target month's last day (Jan 31 + 1m is Feb 28/29).
//
// now is used only for relative forms. Anything else returns an error.
func Parse(s string, now time.Time) (string, error) {
	if strings.HasPrefix(s, "+") {
		return parseRelative(s, now)
	}
	return parseAbsolute(s)
}

// parseAbsolute parses "YY-MM-DD HH:MM" into the canonical stored form.
// time.Parse validates the ranges (month, day, hour, minute); the
// century is then forced to 20YY, because Go's "06" maps years 69-99
// into the 1900s — useless for a defer target, which is always in this
// century.
func parseAbsolute(s string) (string, error) {
	t, err := time.ParseInLocation(absoluteLayout, s, time.Local)
	if err != nil {
		return "", fmt.Errorf("invalid defer time %q: want YY-MM-DD HH:MM (e.g. 26-08-20 08:00) or a relative duration (+2d, +1w, +3h, +1m, +1y)", s)
	}
	if t.Year() < 2000 {
		t = t.AddDate(100, 0, 0)
	}
	return t.Format(issue.NaiveLayout), nil
}

// parseRelative parses "+<n><unit>" (n positive, unit d/w/h/m/y) and
// returns now plus that amount, formatted to the canonical stored form.
// Duration units (d/w/h) add exact durations; calendar units (m/y) add
// whole months/years with end-of-month clamping.
func parseRelative(s string, now time.Time) (string, error) {
	body := s[1:] // drop the "+" (Parse guarantees the prefix)
	if len(body) < 2 {
		return "", invalidDuration(s)
	}
	unit := body[len(body)-1]
	numStr := body[:len(body)-1]
	// The count must be a plain positive integer: reject signs and stray
	// characters that Atoi would silently accept (e.g. "+2", "-2").
	for i := 0; i < len(numStr); i++ {
		if numStr[i] < '0' || numStr[i] > '9' {
			return "", invalidDuration(s)
		}
	}
	n, err := strconv.Atoi(numStr)
	if err != nil || n <= 0 {
		return "", invalidDuration(s)
	}
	if unit == 'y' {
		return calendarUntil(s, now, n, 0)
	}
	if unit == 'm' {
		return calendarUntil(s, now, n/12, n%12)
	}
	per, ok := unitDuration(unit)
	if !ok {
		return "", fmt.Errorf("invalid defer duration %q: unit must be %s", s, unitList)
	}
	count := time.Duration(n)
	if count > maxDuration/per {
		return "", invalidDuration(s)
	}
	return now.Add(count * per).Format(issue.NaiveLayout), nil
}

// calendarUntil returns the canonical stored form of now shifted by the
// given calendar years and months, or invalidDuration(s) when the target
// falls outside the four-digit year range NaiveLayout can represent.
func calendarUntil(s string, now time.Time, years, months int) (string, error) {
	t, ok := addCalendar(now, years, months)
	if !ok {
		return "", invalidDuration(s)
	}
	return t.Format(issue.NaiveLayout), nil
}

// addCalendar returns now shifted by years and months with calendar
// semantics: the day-of-month is kept when it exists in the target
// month, and clamped to that month's last day otherwise (Jan 31 + 1
// month is Feb 28/29; Feb 29 + 1 year is Feb 28). The clock time is
// preserved. Go's time.AddDate normalizes an overflowing day into the
// next month (Jan 31 + 1 month is Mar 2/3), which is the wrong target
// for a deferral, so the day is clamped explicitly. ok is false when the
// target year exceeds maxYear.
//
// months is expected to be under 12 (the caller splits a month count into
// years and a remainder), so the month arithmetic cannot overflow.
func addCalendar(now time.Time, years, months int) (time.Time, bool) {
	// Check the year before adding so a count near the Atoi limit cannot
	// overflow the int year arithmetic.
	if years > maxYear-now.Year() {
		return time.Time{}, false
	}
	total := int(now.Month()) - 1 + months
	year := now.Year() + years + total/12
	if year > maxYear {
		return time.Time{}, false
	}
	month := time.Month(total%12 + 1)
	day := now.Day()
	if last := lastDayOfMonth(year, month); day > last {
		day = last
	}
	return time.Date(year, month, day, now.Hour(), now.Minute(), 0, 0, now.Location()), true
}

// lastDayOfMonth returns the last day of year/month: day 0 of the next
// month normalizes to it (month 13 rolls into January of year+1).
func lastDayOfMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

// invalidDuration renders the shared error for a malformed relative
// duration.
func invalidDuration(s string) error {
	return fmt.Errorf("invalid defer duration %q: want +<n><unit> with a positive count and unit %s (e.g. +2d)", s, unitList)
}

// unitDuration returns the duration of one unit of the relative form,
// and whether u is a known unit (d = days, w = weeks, h = hours).
func unitDuration(u byte) (time.Duration, bool) {
	switch u {
	case 'd':
		return 24 * time.Hour, true
	case 'w':
		return 7 * 24 * time.Hour, true
	case 'h':
		return time.Hour, true
	default:
		return 0, false
	}
}
