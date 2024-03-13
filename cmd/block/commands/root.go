package commands

import (
	"github.com/connorkuljis/block-cli/internal/app"
	"github.com/urfave/cli/v2"
)

var RootCmd = &cli.Command{
	Name:      "start",
	Usage:     "start the blocker.",
	Args:      true,
	ArgsUsage: "[duration] [taskname]",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "no-blocker",
			Usage: "Disables the blocker.",
		},
		&cli.BoolFlag{
			Name:  "capture",
			Usage: "Enables screen capture.",
		},
	},
	Action: func(ctx *cli.Context) error {
		var app = app.App{}
		err := app.Init(ctx)
		if err != nil {
			return err
		}

		err = app.Start()
		if err != nil {
			return err
		}

		err = app.SaveAndExit()
		if err != nil {
			return err
		}

		return nil
	},
}
