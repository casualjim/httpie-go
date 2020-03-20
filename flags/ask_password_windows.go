// +build windows

package flags

import (
	"fmt"
	"os"

	"golang.org/x/crypto/ssh/terminal"
)

func askPassword() (string, error) {
	fmt.Fprintf(os.Stderr, "Password: ")
	fd := int(os.Stdin.Fd())
	password, err := terminal.ReadPassword(fd)
	if err != nil {
		return "", fmt.Errorf("failed to read password from terminal: %w", err)
	}
	fmt.Fprintln(os.Stderr)
	return string(password), nil
}
