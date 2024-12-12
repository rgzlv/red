//go:build unix && !darwin

package term

import (
	"golang.org/x/sys/unix"
)

// List of constants are in ioctl_tty(2) and descriptions for each constant are
// under the 2const section.

const ioctlGetTermios = unix.TCGETS
const ioctlSetTermios = unix.TCSETS
