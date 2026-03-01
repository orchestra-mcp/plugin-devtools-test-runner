package runner

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// Run executes a command in the given directory and returns combined stdout+stderr output.
// args[0] is the executable; the rest are arguments.
func Run(ctx context.Context, dir string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s\n%s", err, out)
	}
	return strings.TrimSpace(string(out)), nil
}
