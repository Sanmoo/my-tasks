package issue

import (
	"fmt"
	"io"
	"strings"
)

// The comment feature of an Issue: a comment is a ### heading carrying a
// naive timestamp, the comment text, and a stable HTML-comment anchor
// (<!-- comment: <token> -->). Comments are append-only — the existing
// body is never rewritten, only extended.

// anchorAlphabet is the symbol set of a comment anchor: the hex digits
// (the spec's example anchor is 4f2b9c1a). Its length divides 256 evenly,
// so the modulo mapping in NewAnchor is unbiased.
const anchorAlphabet = "0123456789abcdef"

// anchorLen is the length of a comment anchor token (8 hex chars, 32
// bits), matching the spec's example (4f2b9c1a). An anchor only needs to
// be unique within its own Issue; at personal scale (tens of comments) the
// collision odds are negligible.
const anchorLen = 8

// maxAnchorAttempts bounds collision retries when allocating an anchor. A
// collision is exceptionally unlikely, but a bound keeps a broken or
// deterministic randomness source from hanging the command forever.
const maxAnchorAttempts = 100

// NewAnchor returns a fresh random comment anchor of anchorLen hex chars,
// reading randomness from rng. The anchor is a stable marker: once written
// it is never rewritten (comments are append-only), so later commands can
// address a comment by it.
func NewAnchor(rng io.Reader) (string, error) {
	return randomToken(rng, anchorAlphabet, anchorLen, "comment anchor")
}

// NextAnchor returns a generated anchor that is not already present in body.
// Existing markers are compared in their complete HTML-comment form, so an
// anchor is unique within an Issue while the existing body remains opaque to
// the allocator. It retries a bounded number of times before returning an
// error.
func NextAnchor(rng io.Reader, body string) (string, error) {
	for range maxAnchorAttempts {
		anchor, err := NewAnchor(rng)
		if err != nil {
			return "", err
		}
		marker := fmt.Sprintf("<!-- comment: %s -->", anchor)
		if !strings.Contains(body, marker) {
			return anchor, nil
		}
	}
	return "", fmt.Errorf("could not allocate a unique comment anchor after %d attempts", maxAnchorAttempts)
}

// AppendComment returns body with one comment appended at the end: a ###
// heading carrying the timestamp, the comment text verbatim, then a stable
// <!-- comment: anchor --> marker. The existing body is preserved
// byte-for-byte — only the new block is added — and a newline is inserted
// before the heading only when the body does not already end with one.
// Comments is always the last body section, so appending at the end is
// appending to Comments.
func AppendComment(body, timestamp, text, anchor string) string {
	var b strings.Builder
	b.WriteString(body)
	if !strings.HasSuffix(body, "\n") {
		b.WriteString("\n")
	}
	fmt.Fprintf(&b, "### %s\n%s\n<!-- comment: %s -->\n", timestamp, text, anchor)
	return b.String()
}
