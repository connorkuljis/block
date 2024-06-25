package commands

import (
	"fmt"
	"log/slog"

	"github.com/connorkuljis/block-cli/internal/blocker"
	"github.com/urfave/cli/v2"
)

var UpCmd = &cli.Command{
	Name:  "up",
	Usage: "enable the blocker",
	Action: func(ctx *cli.Context) error {
		slog.Info("Blocker up.")
		blocker := blocker.NewBlocker()
		n, err := blocker.Start()
		if err != nil {
			return fmt.Errorf("Error running up command: %w", err)
		}
		slog.Info(fmt.Sprintf("%d bytes written", n))
		return nil
	},
}
