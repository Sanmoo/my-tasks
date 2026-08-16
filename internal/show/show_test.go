package show_test

import (
	"strings"
	"testing"

	"github.com/Sanmoo/my-tasks2/internal/issue"
	"github.com/Sanmoo/my-tasks2/internal/show"
)

// full is an Issue with every optional field set, mirroring the fields
// a migrated vault can carry.
func full() issue.Issue {
	rank := 3
	return issue.Issue{
		Frontmatter: issue.Frontmatter{
			Title:         "Comprar material",
			Status:        "in_progress",
			Labels:        []string{"compras", "familia"},
			CreatedAt:     "2026-06-26T18:00",
			Rank:          &rank,
			DeferredUntil: "2026-08-23T00:00",
			Deadline:      "2026-08-30T00:00",
			StartedAt:     "2026-06-27T09:00",
			CompletedAt:   "2026-06-30T10:51",
			BlockedBy:     []string{"bjd-001", "bjd-002"},
		},
		Body: "## Description\ncorpo\n",
	}
}

// minimal is an Issue with only the always-present fields.
func minimal() issue.Issue {
	return issue.Issue{
		Frontmatter: issue.Frontmatter{
			Title:     "Ideia",
			Status:    "open",
			Labels:    []string{},
			CreatedAt: "2026-06-26T18:00",
		},
		Body: "## Description\n",
	}
}

func TestRenderPlainFull(t *testing.T) {
	got := show.Render(full(), "pkm-0b4", show.Options{Color: false})
	want := `◐ pkm-0b4 . Comprar material [in_progress]
Created: 2026-06-26 18:00
Labels: compras, familia
Rank: 3
Deferred until: 2026-08-23 00:00
Deadline: 2026-08-30 00:00
Started: 2026-06-27 09:00
Completed: 2026-06-30 10:51
Blocked by: bjd-001, bjd-002

## Description
corpo
`
	if got != want {
		t.Errorf("Render(plain full) = %q, want %q", got, want)
	}
	if strings.Contains(got, "\x1b[") {
		t.Errorf("plain render must not contain ANSI codes:\n%q", got)
	}
}

// TestRenderPlainBodyVerbatimLeadingNewline guards the verbatim body:
// a body that starts with a blank line keeps it (the DefaultBody
// shape), on top of the view's own separator blank line.
func TestRenderPlainBodyVerbatimLeadingNewline(t *testing.T) {
	i := minimal()
	i.Body = "\n## Description\n"
	got := show.Render(i, "pkm-x", show.Options{Color: false})
	want := "○ pkm-x . Ideia [open]\nCreated: 2026-06-26 18:00\n\n\n## Description\n"
	if got != want {
		t.Errorf("Render(verbatim leading newline) = %q, want %q", got, want)
	}
}

func TestRenderPlainOmitsEmptyFields(t *testing.T) {
	got := show.Render(minimal(), "pkm-0b4", show.Options{Color: false})
	want := `○ pkm-0b4 . Ideia [open]
Created: 2026-06-26 18:00

## Description
`
	if got != want {
		t.Errorf("Render(plain minimal) = %q, want %q", got, want)
	}
	for _, absent := range []string{"Labels:", "Rank:", "Deferred until:", "Deadline:", "Started:", "Completed:", "Blocked by:"} {
		if strings.Contains(got, absent) {
			t.Errorf("plain minimal render must omit %q:\n%q", absent, got)
		}
	}
}

func TestRenderPlainEmptyBody(t *testing.T) {
	i := minimal()
	i.Body = ""
	got := show.Render(i, "pkm-0b4", show.Options{Color: false})
	want := "○ pkm-0b4 . Ideia [open]\nCreated: 2026-06-26 18:00\n"
	if got != want {
		t.Errorf("Render(empty body) = %q, want %q", got, want)
	}
}

func TestRenderColor(t *testing.T) {
	got := show.Render(full(), "pkm-0b4", show.Options{Color: true})
	// The header and metadata block render with exact ANSI codes; the
	// glamour body is asserted loosely below (its padding spans are an
	// implementation detail of the library).
	want := "\x1b[38;2;255;180;84m◐\x1b[0m pkm-0b4 \x1b[38;2;108;118;128m.\x1b[0m \x1b[1mComprar material\x1b[0m [\x1b[38;2;255;180;84min_progress\x1b[0m]\n" +
		"\x1b[38;2;89;194;255mCreated:\x1b[0m 2026-06-26 18:00\n" +
		"\x1b[38;2;89;194;255mLabels:\x1b[0m compras, familia\n" +
		"\x1b[38;2;89;194;255mRank:\x1b[0m 3\n" +
		"\x1b[38;2;89;194;255mDeferred until:\x1b[0m 2026-08-23 00:00\n" +
		"\x1b[38;2;89;194;255mDeadline:\x1b[0m 2026-08-30 00:00\n" +
		"\x1b[38;2;89;194;255mStarted:\x1b[0m 2026-06-27 09:00\n" +
		"\x1b[38;2;89;194;255mCompleted:\x1b[0m 2026-06-30 10:51\n" +
		"\x1b[38;2;89;194;255mBlocked by:\x1b[0m bjd-001, bjd-002\n\n"
	if !strings.HasPrefix(got, want) {
		t.Errorf("Render(color) = %q, want prefix %q", got, want)
	}
	body := strings.TrimPrefix(got, want)
	if !strings.Contains(body, "\x1b[") {
		t.Errorf("color render must contain ANSI codes, got %q", body)
	}
	if !strings.Contains(body, "Description") {
		t.Errorf("color render must contain the heading text, got %q", body)
	}
	// The glamour heading re-styles the line: the raw markdown marker
	// no longer appears as a plain substring.
	if strings.Contains(body, "## Description") {
		t.Errorf("color render must restyle the heading, got %q", body)
	}
	// The body paragraph is rendered in the dark style's text color.
	if !strings.Contains(body, "\x1b[38;5;252mcorpo\x1b[0m") {
		t.Errorf("color render must style the paragraph, got %q", body)
	}
}

func TestRenderColorStatuses(t *testing.T) {
	cases := []struct {
		status string
		glyph  string
		hex    string // "" = plain
	}{
		{"open", "○", ""},
		{"in_progress", "◐", "\x1b[38;2;255;180;84m"},
		{"done", "●", "\x1b[38;2;128;144;160m"},
		{"waiting", "?", ""},
	}
	for _, tc := range cases {
		i := minimal()
		i.Frontmatter.Status = tc.status
		got := show.Render(i, "pkm-x", show.Options{Color: true})
		header := strings.SplitN(got, "\n", 2)[0]
		wantGlyph := tc.glyph
		if tc.hex != "" {
			wantGlyph = tc.hex + tc.glyph + "\x1b[0m"
		}
		wantStatus := tc.status
		if tc.hex != "" {
			wantStatus = tc.hex + tc.status + "\x1b[0m"
		}
		wantHeader := wantGlyph + " pkm-x " + "\x1b[38;2;108;118;128m.\x1b[0m \x1b[1mIdeia\x1b[0m [" + wantStatus + "]"
		if header != wantHeader {
			t.Errorf("header for status %q = %q, want %q", tc.status, header, wantHeader)
		}
	}
}

func TestRenderColorLightTheme(t *testing.T) {
	t.Setenv("MT_THEME", "light")
	got := show.Render(full(), "pkm-0b4", show.Options{Color: true})
	// Light variants: accent #399ee6, in_progress #f2ae49, muted #828c99.
	want := "\x1b[38;2;242;174;73m◐\x1b[0m pkm-0b4 \x1b[38;2;130;140;153m.\x1b[0m \x1b[1mComprar material\x1b[0m [\x1b[38;2;242;174;73min_progress\x1b[0m]\n" +
		"\x1b[38;2;57;158;230mCreated:\x1b[0m 2026-06-26 18:00\n"
	if !strings.HasPrefix(got, want) {
		t.Errorf("Render(light) = %q, want prefix %q", got, want)
	}
}

func TestRenderBodyVerbatimWithoutColor(t *testing.T) {
	body := "\n## Description\n- a\n- b\n\n**bold**\n"
	i := minimal()
	i.Body = body
	got := show.Render(i, "pkm-x", show.Options{Color: false})
	if !strings.HasSuffix(got, body) {
		t.Errorf("plain render must pass the body verbatim, got %q", got)
	}
}

func TestRenderBodyGlamourWithColor(t *testing.T) {
	i := minimal()
	i.Body = "\n## Description\n- a\n- b\n"
	got := show.Render(i, "pkm-x", show.Options{Color: true, Width: 40})
	if !strings.Contains(got, "\x1b[") {
		t.Errorf("color render must contain ANSI codes, got %q", got)
	}
	if !strings.Contains(got, "Description") {
		t.Errorf("color render must contain the heading text, got %q", got)
	}
	// The glamour heading re-styles the line: the raw markdown marker
	// no longer appears as a plain substring.
	if strings.Contains(got, "## Description") {
		t.Errorf("color render must restyle the heading, got %q", got)
	}
}

// TestRenderBodyWrap exercises the wrap rules: the default 80 when the
// width is unknown, and the cap at 100. The phrases are sized so that
// the line counts differ between the widths under test (glamour wraps
// at spaces only).
func TestRenderBodyWrap(t *testing.T) {
	// 18 words = 90 columns: one line at 100, wrapped at 80.
	line90 := strings.Repeat("word ", 18)
	// 35 words = 175 columns: one line uncapped at 200, wrapped at 100.
	line175 := strings.Repeat("word ", 35)
	tests := []struct {
		name      string
		opts      show.Options
		line      string
		wantLines int
	}{
		{"unknown width wraps at 80", show.Options{Color: true}, line90, 2},
		{"wide width caps at 100", show.Options{Color: true, Width: 200}, line90, 1},
		{"line beyond the cap still wraps", show.Options{Color: true, Width: 200}, line175, 2},
		{"explicit narrow width is honored", show.Options{Color: true, Width: 40}, line90, 3},
	}
	for _, tc := range tests {
		i := minimal()
		i.Body = tc.line + "\n"
		got := show.Render(i, "pkm-x", tc.opts)
		textLines := 0
		for _, l := range strings.Split(got, "\n") {
			if strings.Contains(l, "word") {
				textLines++
			}
		}
		if textLines != tc.wantLines {
			t.Errorf("%s: got %d text lines, want %d", tc.name, textLines, tc.wantLines)
		}
	}
}

func TestShouldUseColor(t *testing.T) {
	t.Run("tty only by default", func(t *testing.T) {
		if show.ShouldUseColor(true) != true {
			t.Error("tty without env must use color")
		}
		if show.ShouldUseColor(false) != false {
			t.Error("pipe without env must not use color")
		}
	})
	t.Run("NO_COLOR wins over everything", func(t *testing.T) {
		t.Setenv("NO_COLOR", "1")
		t.Setenv("CLICOLOR_FORCE", "1")
		if show.ShouldUseColor(true) != false {
			t.Error("NO_COLOR must disable color even with CLICOLOR_FORCE")
		}
	})
	t.Run("CLICOLOR zero disables", func(t *testing.T) {
		t.Setenv("CLICOLOR", "0")
		if show.ShouldUseColor(true) != false {
			t.Error("CLICOLOR=0 must disable color on a tty")
		}
	})
	t.Run("CLICOLOR_FORCE enables without tty", func(t *testing.T) {
		t.Setenv("CLICOLOR_FORCE", "1")
		if show.ShouldUseColor(false) != true {
			t.Error("CLICOLOR_FORCE must enable color on a pipe")
		}
	})
	t.Run("NO_COLOR disables without tty", func(t *testing.T) {
		t.Setenv("NO_COLOR", "1")
		if show.ShouldUseColor(false) != false {
			t.Error("NO_COLOR must disable color on a pipe")
		}
	})
}

func TestBackgroundIsDark(t *testing.T) {
	t.Run("default is dark", func(t *testing.T) {
		if !show.BackgroundIsDark() {
			t.Error("unset env must default to dark")
		}
	})
	t.Run("MT_THEME overrides", func(t *testing.T) {
		t.Setenv("MT_THEME", "light")
		if show.BackgroundIsDark() {
			t.Error("MT_THEME=light must be light")
		}
		t.Setenv("MT_THEME", "dark")
		if !show.BackgroundIsDark() {
			t.Error("MT_THEME=dark must be dark")
		}
		t.Setenv("MT_THEME", "LIGHT")
		if show.BackgroundIsDark() {
			t.Error("MT_THEME must be case-insensitive")
		}
	})
	t.Run("COLORFGBG honored", func(t *testing.T) {
		t.Setenv("COLORFGBG", "15;0")
		if !show.BackgroundIsDark() {
			t.Error("COLORFGBG=15;0 must be dark")
		}
		t.Setenv("COLORFGBG", "0;15")
		if show.BackgroundIsDark() {
			t.Error("COLORFGBG=0;15 must be light")
		}
		t.Setenv("COLORFGBG", "15;default;7")
		if show.BackgroundIsDark() {
			t.Error("three-field COLORFGBG must read the last field")
		}
	})
	t.Run("malformed COLORFGBG falls back to dark", func(t *testing.T) {
		t.Setenv("COLORFGBG", "abc")
		if !show.BackgroundIsDark() {
			t.Error("malformed COLORFGBG must fall back to dark")
		}
	})
}
