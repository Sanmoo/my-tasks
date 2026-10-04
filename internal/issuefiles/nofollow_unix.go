//go:build aix || darwin || dragonfly || freebsd || illumos || linux || netbsd || openbsd || solaris

package issuefiles

import "syscall"

// issueOpenNoFollow prevents opening an Issue symlink on Unix-like
// systems: even if the path changes between the Lstat validation and
// the open, the open itself refuses to follow the link.
const issueOpenNoFollow = syscall.O_NOFOLLOW
