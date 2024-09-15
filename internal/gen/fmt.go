package gen

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
)

func goFmt(ctx context.Context, formatFile string) error {
	var stdErr bytes.Buffer
	cmd := exec.CommandContext(ctx, "gofmt", "-w", formatFile)
	cmd.Stderr = &stdErr
	cmd.Stdout = io.Discard
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("gofmt %w failed: %s", err, stdErr.String())
	}
	return cmd.Run()
}
