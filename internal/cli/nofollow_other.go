//go:build !aix && !darwin && !dragonfly && !freebsd && !illumos && !linux && !netbsd && !openbsd && !solaris

package cli

// issueOpenNoFollow is unavailable on this platform. The standard library
// has no portable no-follow open flag here: regular-file checks reject
// symlinks discovered in the Vault, while a concurrent replacement requires
// an OS-specific API.
const issueOpenNoFollow = 0
