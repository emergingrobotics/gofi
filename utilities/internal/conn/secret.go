// secret.go obtains a secret without putting it on the command line: no CLI
// flag ever accepts a password or API key value directly (C-GLOBAL-005).
package conn

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/term"
)

// ReadSecret resolves a secret: the named command if given, otherwise a
// prompt on the controlling terminal with echo disabled.
func ReadSecret(prompt, command string) (string, error) {
	if command != "" {
		return runSecretCommand(command)
	}
	return promptSecret(prompt)
}

// promptSecret reads from /dev/tty rather than stdin, so a secret can still
// be typed while stdin carries piped data (e.g. `gofi ips import -`).
func promptSecret(prompt string) (string, error) {
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return "", fmt.Errorf("no terminal to prompt on: %w", err)
	}
	defer tty.Close()

	fd := int(tty.Fd())
	if !term.IsTerminal(fd) {
		return "", errors.New("not a terminal; use the --*-command form instead")
	}

	fmt.Fprint(tty, prompt)
	raw, err := term.ReadPassword(fd)
	fmt.Fprintln(tty)
	if err != nil {
		return "", fmt.Errorf("reading the secret: %w", err)
	}

	secret := strings.TrimRight(string(raw), "\r\n")
	if secret == "" {
		return "", errors.New("empty secret")
	}
	return secret, nil
}

func runSecretCommand(command string) (string, error) {
	cmd := exec.Command("sh", "-c", command)
	cmd.Stderr = os.Stderr

	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("secret command failed (%s): %w", command, err)
	}

	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	if !scanner.Scan() {
		return "", fmt.Errorf("secret command produced no output: %s", command)
	}
	secret := strings.TrimRight(scanner.Text(), "\r")
	if secret == "" {
		return "", fmt.Errorf("secret command produced an empty first line: %s", command)
	}
	return secret, nil
}
