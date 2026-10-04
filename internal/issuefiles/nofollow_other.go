//go:build !aix && !darwin && !dragonfly && !freebsd && !illumos && !linux && !netbsd && !openbsd && !solaris

package issuefiles

// issueOpenNoFollow is unavailable on this platform. The standard
// library has no portable no-follow open flag here: the regular-file
// checks reject symlinks discovered in the Vault, while a concurrent
// replacement requires an OS-specific API.
const issueOpenNoFollow = 0
