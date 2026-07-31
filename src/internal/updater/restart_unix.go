//go:build !windows

package updater

import (
	"os"
	"syscall"
)

// Restart replaces the current process image with the new binary. The PID is
// kept and no second process appears, so the shell sees one continuous
// program. Must be called only after bubbletea has restored the terminal.
func Restart() error {
	exe, err := SwappedExecutablePath()

	if err != nil {
		return err
	}

	return syscall.Exec(exe, os.Args, os.Environ())
}
