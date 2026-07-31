//go:build windows

package updater

import (
	"os"
	"os/exec"
)

// Restart launches the new binary and lets this process exit. Windows has no
// exec that replaces a process image, and the running executable has already
// been renamed aside by the swap.
func Restart() error {
	exe, err := SwappedExecutablePath()

	if err != nil {
		return err
	}

	cmd := exec.Command(exe, os.Args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Start()
}
