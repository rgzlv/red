//go:build darwin

package term

import (
	"golang.org/x/sys/unix"
)

// List of constants and descriptions are in tty(4).

const ioctlGetTermios = unix.TIOCGETA
const ioctlSetTermios = unix.TIOCSETAF
