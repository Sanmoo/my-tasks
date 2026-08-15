// Package issue_test holds the black-box unit tests of the comment
// feature (Seam 2): the stable anchor token and the append-only
// AppendComment that preserves the existing body byte-for-byte.
package issue_test

import (
	"bytes"
	"io"
	"regexp"
	"strings"
	"testing"

	"github.com/Sanmoo/my-tasks2/internal/issue"
)

func TestNewAnchorIsEightHexChars(t *testing.T) {
	got, err := issue.NewAnchor(bytes.NewReader(make([]byte, 8)))
	if err != nil {
		t.Fatal(err)
	}
	if !regexp.MustCompile(`^[0-9a-f]{8}$`).MatchString(got) {
		t.Errorf("NewAnchor = %q, want 8 hex chars", got)
	}
}

func TestNewAnchorDeterministicMapping(t *testing.T) {
	// 0x00 -> '0', 0x0f -> 'f', 0x10 -> '0', 0x1a -> 'a', 0xff -> 'f',
	// 0x01 -> '1', 0x02 -> '2', 0x03 -> '3'.
	got, err := issue.NewAnchor(&seqReader{data: []byte{0x00, 0x0f, 0x10, 0x1a, 0xff, 0x01, 0x02, 0x03}})
	if err != nil {
		t.Fatal(err)
	}
	if got != "0f0af123" {
		t.Errorf("NewAnchor = %q, want %q", got, "0f0af123")
	}
}

func TestNewAnchorShortReadFails(t *testing.T) {
	// The reader yields one byte then io.EOF: io.ReadFull cannot fill 8.
	if _, err := issue.NewAnchor(&seqReader{data: []byte{0x01}, err: io.EOF}); err == nil {
		t.Fatal("NewAnchor(short read) = nil error, want failure")
	}
}

func TestNextAnchorRetriesOnCollision(t *testing.T) {
	// The first candidate is already present; the second candidate is free.
	rng := &seqReader{data: []byte{
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01,
	}}
	body := "<!-- comment: 00000000 -->"
	got, err := issue.NextAnchor(rng, body)
	if err != nil {
		t.Fatal(err)
	}
	if got != "11111111" {
		t.Errorf("NextAnchor = %q, want %q", got, "11111111")
	}
}

func TestNextAnchorExhaustsAttempts(t *testing.T) {
	body := "<!-- comment: 00000000 -->"
	_, err := issue.NextAnchor(bytes.NewReader(make([]byte, 8*100)), body)
	if err == nil {
		t.Fatal("NextAnchor(all taken) = nil error, want failure")
	}
	if !strings.Contains(err.Error(), "unique") {
		t.Errorf("error %q does not mention uniqueness", err)
	}
}

func TestAppendCommentToFreshBody(t *testing.T) {
	got := issue.AppendComment(issue.DefaultBody, "2026-08-16T14:05", "Comprei metade da lista.", "4f2b9c1a")
	want := "\n## Description\n## Notes\n## Comments\n### 2026-08-16T14:05\nComprei metade da lista.\n<!-- comment: 4f2b9c1a -->\n"
	if got != want {
		t.Errorf("AppendComment = %q, want %q", got, want)
	}
}

func TestAppendCommentPreservesExistingBody(t *testing.T) {
	body := "\n## Description\n## Notes\n## Comments\n### 2026-08-16T14:05\nfirst.\n<!-- comment: 4f2b9c1a -->\n"
	got := issue.AppendComment(body, "2026-08-16T15:00", "second.", "9a1b2c3d")
	if !strings.HasPrefix(got, body) {
		t.Errorf("AppendComment did not preserve the existing body byte-for-byte:\ngot:  %q\nwant prefix: %q", got, body)
	}
	wantSuffix := "### 2026-08-16T15:00\nsecond.\n<!-- comment: 9a1b2c3d -->\n"
	if !strings.HasSuffix(got, wantSuffix) {
		t.Errorf("AppendComment = %q, want suffix %q", got, wantSuffix)
	}
}

func TestAppendCommentInsertsNewlineWhenBodyLacksOne(t *testing.T) {
	got := issue.AppendComment("## Comments", "2026-08-16T14:05", "x", "abcd1234")
	want := "## Comments\n### 2026-08-16T14:05\nx\n<!-- comment: abcd1234 -->\n"
	if got != want {
		t.Errorf("AppendComment = %q, want %q", got, want)
	}
}

func TestAppendCommentDoesNotDoubleNewline(t *testing.T) {
	// A body already ending in newline gets the heading directly, with no
	// blank line between them.
	got := issue.AppendComment("## Comments\n", "2026-08-16T14:05", "x", "abcd1234")
	want := "## Comments\n### 2026-08-16T14:05\nx\n<!-- comment: abcd1234 -->\n"
	if got != want {
		t.Errorf("AppendComment = %q, want %q", got, want)
	}
}

func TestAppendCommentTextIsVerbatim(t *testing.T) {
	got := issue.AppendComment("## Comments\n", "2026-08-16T14:05", "line one\nline two", "abcd1234")
	if !strings.Contains(got, "line one\nline two") {
		t.Errorf("AppendComment did not keep text verbatim: %q", got)
	}
}

func TestAppendCommentLeavesOriginalUntouched(t *testing.T) {
	body := "## Comments\n"
	_ = issue.AppendComment(body, "2026-08-16T14:05", "x", "abcd1234")
	if body != "## Comments\n" {
		t.Errorf("AppendComment mutated its input: %q", body)
	}
}
