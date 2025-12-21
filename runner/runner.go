package runner

import (
	"context"
	"os"
	"os/exec"
)

// Esto nos permite usar un MockCommander en los tests.
type Commander interface {
	Run(ctx context.Context, command string, args []string) error
}

type OSCommander struct {
	// Aquí podríamos guardar referencias a stdout/stderr si fuera necesario
}

func NewOSCommander() *OSCommander {
	return &OSCommander{}
}

func (o *OSCommander) Run(ctx context.Context, command string, args []string) error {
	cmd := exec.CommandContext(ctx, command, args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}
