package commands

import (
	"fmt"
	"log/slog"

	"github.com/connorkuljis/block-cli/internal/blocker"
	"github.com/urfave/cli/v2"
)

var DownCmd = &cli.Command{
	Name:  "down",
	Usage: "disable the blocker",
	Action: func(ctx *cli.Context) error {
		slog.Info("Blocker down.")
		blocker := blocker.NewBlocker()
		n, err := blocker.Stop()
		if err != nil {
			return fmt.Errorf("Error running down command: %w", err)
		}
		slog.Info(fmt.Sprintf("%d bytes written", n))
		return nil
	},
}
