// +build !windows

package flags

import (
	"fmt"
	"os"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"
)

func askPassword() (string, error) {
	var fd int
	if terminal.IsTerminal(syscall.Stdin) {
		fd = syscall.Stdin
	} else {
		tty, err := os.Open("/dev/tty")
		if err != nil {
			return "", fmt.Errorf("failed to allocate terminal: %w", err)
		}
		defer tty.Close()
		fd = int(tty.Fd())
	}

	fmt.Fprintf(os.Stderr, "Password: ")
	password, err := terminal.ReadPassword(fd)
	if err != nil {
		return "", fmt.Errorf("failed to read password from terminal: %w", err)
	}
	fmt.Fprintln(os.Stderr)
	return string(password), nil
}
