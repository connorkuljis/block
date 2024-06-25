package commands

import (
	"fmt"

	"github.com/connorkuljis/block-cli/internal/blocker"
	"github.com/urfave/cli/v2"
)

var ResetDNSCmd = &cli.Command{
	Name:  "reset",
	Usage: "Reset DNS cache.",
	Action: func(ctx *cli.Context) error {
		err := blocker.ResetDNS()
		if err != nil {
			return err
		}
		fmt.Println("Successfully reset dns.")
		return nil
	},
}
