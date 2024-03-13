package commands

import (
	"github.com/connorkuljis/block-cli/internal/server"
	"github.com/urfave/cli/v2"
)

var ServeCmd = &cli.Command{
	Name:  "serve",
	Usage: "Serves http server.",
	Action: func(ctx *cli.Context) error {
		err := server.Serve()
		if err != nil {
			return err
		}
		return nil
	},
}
