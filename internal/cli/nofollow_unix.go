//go:build aix || darwin || dragonfly || freebsd || illumos || linux || netbsd || openbsd || solaris

package cli

import "syscall"

// issueOpenNoFollow prevents opening an Issue symlink on Unix-like systems.
const issueOpenNoFollow = syscall.O_NOFOLLOW
