// Package issue_test holds the black-box unit tests of issue ID
// generation (Seam 2): prefix+suffix shape, collision retry, and the
// deterministic randomness source.
package issue_test

import (
	"bytes"
	"io"
	"regexp"
	"strings"
	"testing"

	"github.com/Sanmoo/my-tasks2/internal/issue"
)

// seqReader is a deterministic io.Reader that yields its bytes, then an
// error, so RandomSuffix/NextID are fully controlled in tests.
type seqReader struct {
	data []byte
	err  error
}

func (r *seqReader) Read(p []byte) (int, error) {
	if len(r.data) == 0 {
		return 0, r.err
	}
	n := copy(p, r.data)
	r.data = r.data[n:]
	return n, nil
}

func TestNewID(t *testing.T) {
	if got := issue.NewID("pkm", "055"); got != "pkm-055" {
		t.Errorf("NewID = %q, want %q", got, "pkm-055")
	}
}

func TestRandomSuffixLengthAndAlphabet(t *testing.T) {
	// 0x00 maps to '0', 0xff maps to (255 % 36) -> index 3 -> '3'.
	suffix, err := issue.RandomSuffix(&seqReader{data: []byte{0x00, 0xff, 0x01, 0x02}}, 4)
	if err != nil {
		t.Fatal(err)
	}
	if suffix != "0312" {
		t.Errorf("RandomSuffix = %q, want %q", suffix, "0312")
	}
}

func TestRandomSuffixShortReadFails(t *testing.T) {
	// The reader yields one byte then io.EOF: io.ReadFull cannot fill 4.
	if _, err := issue.RandomSuffix(&seqReader{data: []byte{0x01}, err: io.EOF}, 4); err == nil {
		t.Fatal("RandomSuffix(short read) = nil error, want failure")
	}
}

var idRe = regexp.MustCompile(`^pkm-[0-9a-z]{4}$`)

func TestNextIDReturnsFreeID(t *testing.T) {
	// Suffix bytes 0x00,0x01,0x02,0x03 -> "0123".
	got, err := issue.NextID("pkm", map[string]bool{"pkm-aaaa": true}, &seqReader{data: []byte{0x00, 0x01, 0x02, 0x03}})
	if err != nil {
		t.Fatal(err)
	}
	if !idRe.MatchString(got) {
		t.Errorf("NextID = %q, want prefix + 4 lowercase alnum", got)
	}
	if got == "pkm-aaaa" {
		t.Errorf("NextID = %q, collided with taken", got)
	}
}

func TestNextIDRetriesOnCollision(t *testing.T) {
	// First candidate "0123" is taken; the second attempt "4567" is free.
	rng := &seqReader{data: []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07}}
	got, err := issue.NextID("pkm", map[string]bool{"pkm-0123": true}, rng)
	if err != nil {
		t.Fatal(err)
	}
	if got != "pkm-4567" {
		t.Errorf("NextID = %q, want %q (the retry)", got, "pkm-4567")
	}
}

func TestNextIDExhaustsAttempts(t *testing.T) {
	// The only candidate keeps colliding; after the cap it fails. A
	// reader of 4×100 zero bytes yields "0000" every time.
	_, err := issue.NextID("pkm", map[string]bool{"pkm-0000": true}, bytes.NewReader(make([]byte, 4*100)))
	if err == nil {
		t.Fatal("NextID(all taken) = nil error, want failure")
	}
	if !strings.Contains(err.Error(), "unique") {
		t.Errorf("error %q does not mention uniqueness", err)
	}
}
