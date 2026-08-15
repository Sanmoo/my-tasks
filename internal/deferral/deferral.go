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

// Parse converts a `mt defer` time argument to the canonical
// deferred_until value (issue.NaiveLayout, YYYY-MM-DDTHH:MM naive local
// time). It accepts:
//
//   - an absolute "YY-MM-DD HH:MM"; the two-digit year is expanded to
//     20YY, so any year this century is reachable and the hour is kept;
//   - a relative "+<n><unit>" computed from now, where unit is d (days),
//     w (weeks) or h (hours) and n is a positive integer.
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
		return "", fmt.Errorf("invalid defer time %q: want YY-MM-DD HH:MM (e.g. 26-08-20 08:00) or a relative duration (+2d, +1w, +3h)", s)
	}
	if t.Year() < 2000 {
		t = t.AddDate(100, 0, 0)
	}
	return t.Format(issue.NaiveLayout), nil
}

// parseRelative parses "+<n><unit>" (n positive, unit d/w/h) and returns
// now plus that duration, formatted to the canonical stored form.
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
	per, ok := unitDuration(unit)
	if !ok {
		return "", fmt.Errorf("invalid defer duration %q: unit must be d (days), w (weeks) or h (hours)", s)
	}
	return now.Add(time.Duration(n) * per).Format(issue.NaiveLayout), nil
}

// invalidDuration renders the shared error for a malformed relative
// duration.
func invalidDuration(s string) error {
	return fmt.Errorf("invalid defer duration %q: want +<n><unit> with a positive count and unit d (days), w (weeks) or h (hours) (e.g. +2d)", s)
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
