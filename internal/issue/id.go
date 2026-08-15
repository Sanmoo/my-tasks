package issue

import (
	"fmt"
	"io"
)

// suffixAlphabet is the symbol set of a random ID suffix: digits and
// lowercase letters. Short and URL-safe, it matches the spec examples
// (pkm-055, pkm-07r0).
const suffixAlphabet = "0123456789abcdefghijklmnopqrstuvwxyz"

// suffixLen is the length of a random ID suffix. Short enough to type,
// long enough that collisions are rare; NewID retries on collision.
const suffixLen = 4

// maxAttempts bounds the collision-retry loop of NextID.
const maxAttempts = 100

// NewID joins the vault prefix and a suffix into an issue ID (ex.:
// pkm-055). The file name (issues/<id>.md) is the authority; no id field
// lives in the frontmatter.
func NewID(prefix, suffix string) string {
	return prefix + "-" + suffix
}

// randomToken returns n chars drawn from alphabet, reading randomness from
// rng. what names the token in errors (e.g. "random suffix", "comment
// anchor").
func randomToken(rng io.Reader, alphabet string, n int, what string) (string, error) {
	buf := make([]byte, n)
	if _, err := io.ReadFull(rng, buf); err != nil {
		return "", fmt.Errorf("generating %s: %w", what, err)
	}
	out := make([]byte, n)
	for i, b := range buf {
		out[i] = alphabet[int(b)%len(alphabet)]
	}
	return string(out), nil
}

// RandomSuffix returns a random n-char suffix drawn from suffixAlphabet,
// reading randomness from rng.
func RandomSuffix(rng io.Reader, n int) (string, error) {
	return randomToken(rng, suffixAlphabet, n, "random suffix")
}

// NextID returns a fresh issue ID for prefix that is not in taken,
// retrying with new random suffixes until it finds a free one. taken is
// the set of IDs already on disk; it is not modified.
func NextID(prefix string, taken map[string]bool, rng io.Reader) (string, error) {
	for range maxAttempts {
		suffix, err := RandomSuffix(rng, suffixLen)
		if err != nil {
			return "", err
		}
		id := NewID(prefix, suffix)
		if !taken[id] {
			return id, nil
		}
	}
	return "", fmt.Errorf("could not allocate a unique issue ID for prefix %q after %d attempts", prefix, maxAttempts)
}
