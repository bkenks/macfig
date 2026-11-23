package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// run executes a command, or prints it when --dry-run is set.
// stdin may be nil.
func run(stdin []byte, name string, args ...string) error {
	if opts.dryRun {
		line := shellJoin(append([]string{name}, args...))
		if stdin != nil {
			line += fmt.Sprintf("  # with %d bytes on stdin", len(stdin))
		}
		fmt.Println("would run:", line)
		return nil
	}
	cmd := exec.Command(name, args...)
	if stdin != nil {
		cmd.Stdin = bytes.NewReader(stdin)
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg != "" {
			return fmt.Errorf("%s: %w: %s", name, err, msg)
		}
		return fmt.Errorf("%s: %w", name, err)
	}
	return nil
}

// output runs a read-only command and returns its stdout. Read-only commands
// run even in --dry-run mode.
func output(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("%s %s: %w", name, strings.Join(args, " "), err)
	}
	return out, nil
}

func shellJoin(args []string) string {
	quoted := make([]string, len(args))
	for i, a := range args {
		if a != "" && !strings.ContainsAny(a, " \t\n\"'{}();=$`\\*?[]|&<>!#~") {
			quoted[i] = a
			continue
		}
		quoted[i] = "'" + strings.ReplaceAll(a, "'", `'\''`) + "'"
	}
	return strings.Join(quoted, " ")
}
