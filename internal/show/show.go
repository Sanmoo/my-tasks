// Package show holds the pure logic of `mt show`: the structured Issue
// view — header line, metadata lines, Markdown body — rendered in the
// visual language of `nd show` (ayu palette, glamour body). Color is a
// parameter: the same view renders without ANSI codes when Color is
// off, so piped output stays clean. It is decision-dense, so it lives
// at Seam 2: black-box unit tested, with the coverage and mutation
// gates.
package show

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/glamour"

	"github.com/Sanmoo/my-tasks2/internal/issue"
	"github.com/Sanmoo/my-tasks2/internal/list"
)

// Options control the `mt show` view.
type Options struct {
	// Color enables ANSI codes (bold title, accent labels, status
	// colors). The view structure is identical without it.
	Color bool
	// Width is the terminal width in columns; 0 means unknown and the
	// body wraps at the default width.
	Width int
}

// The ayu palette (the nd show palette), one hex per role per
// background: light first, dark second.
var (
	hexMuted      = [2]string{"#828c99", "#6c7680"}
	hexAccent     = [2]string{"#399ee6", "#59c2ff"}
	hexInProgress = [2]string{"#f2ae49", "#ffb454"}
	hexDone       = [2]string{"#9099a1", "#8090a0"}
)

// Render returns the show view of Issue i with the file-name id:
//
//	◐ pkm-0b4 . Título [in_progress]
//	Created: 2026-06-26 18:00
//	Labels: compras, familia
//	Rank: 3
//	Deferred until: 2026-08-23 00:00
//	Deadline: 2026-08-30 00:00
//	Started: 2026-06-27 09:00
//	Completed: 2026-06-30 10:51
//	Blocked by: bjd-001, bjd-002
//
//	## Description
//	...
//
// Metadata lines appear only when their field is set (labels and
// blocked_by only when non-empty). The body is glamour-rendered when
// Color is on, verbatim otherwise. The returned string ends with a
// trailing newline (the body's own).
func Render(i issue.Issue, id string, opts Options) string {
	dark := BackgroundIsDark()
	fm := i.Frontmatter
	statusColor := statusColor(fm.Status)

	var b strings.Builder

	// Header: GLYPH ID . TITLE [STATUS].
	b.WriteString(fg(list.Glyph(fm.Status), pick(statusColor, dark), opts.Color))
	b.WriteString(" " + id + " ")
	b.WriteString(fg(".", pick(hexMuted, dark), opts.Color))
	b.WriteString(" ")
	b.WriteString(bold(fm.Title, opts.Color))
	b.WriteString(" [")
	b.WriteString(fg(fm.Status, pick(statusColor, dark), opts.Color))
	b.WriteString("]\n")

	meta(&b, "Created", displayTime(fm.CreatedAt), opts.Color, dark)
	if len(fm.Labels) > 0 {
		meta(&b, "Labels", strings.Join(fm.Labels, ", "), opts.Color, dark)
	}
	if fm.Rank != nil {
		meta(&b, "Rank", strconv.Itoa(*fm.Rank), opts.Color, dark)
	}
	if fm.DeferredUntil != "" {
		meta(&b, "Deferred until", displayTime(fm.DeferredUntil), opts.Color, dark)
	}
	if fm.Deadline != "" {
		meta(&b, "Deadline", displayTime(fm.Deadline), opts.Color, dark)
	}
	if fm.StartedAt != "" {
		meta(&b, "Started", displayTime(fm.StartedAt), opts.Color, dark)
	}
	if fm.CompletedAt != "" {
		meta(&b, "Completed", displayTime(fm.CompletedAt), opts.Color, dark)
	}
	if len(fm.BlockedBy) > 0 {
		meta(&b, "Blocked by", strings.Join(fm.BlockedBy, ", "), opts.Color, dark)
	}

	if i.Body != "" {
		b.WriteString("\n")
		b.WriteString(renderBody(i.Body, opts, dark))
	}
	return b.String()
}

// meta writes one "Label: value" line; the label is accented when
// color is on.
func meta(b *strings.Builder, label, value string, color, dark bool) {
	b.WriteString(fg(label+":", pick(hexAccent, dark), color))
	b.WriteString(" " + value + "\n")
}

// displayTime renders a stored naive datetime for display: the
// canonical YYYY-MM-DDTHH:MM becomes YYYY-MM-DD HH:MM. Only the first
// T is touched, so a value that does not follow the canonical layout
// passes through almost unchanged (mt check owns format validation).
func displayTime(v string) string {
	return strings.Replace(v, "T", " ", 1)
}

// renderBody renders the Markdown body with glamour when color is on —
// theme from the resolved background, wrapped at the terminal width
// capped at 100 (default 80) — and verbatim otherwise. A rendering
// failure falls back to the verbatim body.
func renderBody(body string, opts Options, dark bool) string {
	if !opts.Color || body == "" {
		return body
	}
	wrap := opts.Width
	if wrap <= 0 {
		wrap = 80
	}
	if wrap > 100 {
		wrap = 100
	}
	style := "dark"
	if !dark {
		style = "light"
	}
	renderer, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle(style),
		glamour.WithWordWrap(wrap),
	)
	if err != nil {
		return body
	}
	rendered, err := renderer.Render(body)
	if err != nil {
		return body
	}
	return rendered
}

// statusColor returns the hex pair of a status, or the zero pair when
// the status renders plain (open and any custom status).
func statusColor(status string) [2]string {
	switch status {
	case "in_progress":
		return hexInProgress
	case "done":
		return hexDone
	default:
		return [2]string{}
	}
}

// pick returns the light or dark variant of a hex pair.
func pick(pair [2]string, dark bool) string {
	if dark {
		return pair[1]
	}
	return pair[0]
}

// fg wraps s in a 24-bit foreground color escape when color is on and
// hex is non-empty; plain s otherwise.
func fg(s, hex string, color bool) string {
	if !color || hex == "" {
		return s
	}
	return "\x1b[38;2;" + rgb(hex) + "m" + s + "\x1b[0m"
}

// rgb converts "#rrggbb" to the "r;g;b" ANSI components.
func rgb(hex string) string {
	h := strings.TrimPrefix(hex, "#")
	r, _ := strconv.ParseUint(h[0:2], 16, 8)
	g, _ := strconv.ParseUint(h[2:4], 16, 8)
	b, _ := strconv.ParseUint(h[4:6], 16, 8)
	return fmt.Sprintf("%d;%d;%d", r, g, b)
}

// bold wraps s in a bold escape when color is on; plain s otherwise.
func bold(s string, color bool) string {
	if !color {
		return s
	}
	return "\x1b[1m" + s + "\x1b[0m"
}

// ShouldUseColor reports whether mt show should emit ANSI colors under
// the standard convention (no-color.org):
//
//   - NO_COLOR set (any value): no color
//   - CLICOLOR=0: no color
//   - CLICOLOR_FORCE set (any value): color, even when tty is false
//   - otherwise: color only when tty (stdout is a terminal)
func ShouldUseColor(tty bool) bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	if os.Getenv("CLICOLOR") == "0" {
		return false
	}
	if os.Getenv("CLICOLOR_FORCE") != "" {
		return true
	}
	return tty
}

// BackgroundIsDark reports whether the view should render for a dark
// terminal background. It never probes the terminal (an OSC 11 query
// can leak its raw reply onto the screen); it resolves deterministically
// from the environment, defaulting to dark — the common case.
//
// Resolution order:
//   - MT_THEME=light|dark: explicit override, wins over everything
//   - COLORFGBG="fg;bg": honored when the background field is an ANSI
//     index (7 and 15 are white = light; every other index is dark)
//   - default: dark
func BackgroundIsDark() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("MT_THEME"))) {
	case "light":
		return false
	case "dark":
		return true
	}
	if bg, ok := colorFGBGBackground(os.Getenv("COLORFGBG")); ok {
		return bg != 7 && bg != 15
	}
	return true
}

// colorFGBGBackground extracts the background ANSI color index from a
// COLORFGBG value of the form "fg;bg" (e.g. "15;0"). Some terminals
// emit a three-field "fg;default;bg" form, so the background is always
// the last field. ok is false when the value is missing or malformed.
func colorFGBGBackground(v string) (int, bool) {
	if !strings.Contains(v, ";") {
		return 0, false
	}
	parts := strings.Split(v, ";")
	bg, err := strconv.Atoi(strings.TrimSpace(parts[len(parts)-1]))
	if err != nil {
		return 0, false
	}
	return bg, true
}
