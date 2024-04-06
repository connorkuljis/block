package commands

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/connorkuljis/block-cli/src/app"
	"github.com/urfave/cli/v2"
)

var StartCmd = &cli.Command{
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
			Name:    "capture",
			Aliases: []string{"c"},
			Usage:   "Enables screen capture.",
		},
	},
	Action: func(ctx *cli.Context) error {
		// validate args length
		if ctx.NArg() < 1 {
			return errors.New("Error, no arguments provided")
		}

		durationArg := ctx.Args().Get(0)
		taskNameArg := ctx.Args().Get(1) // empty string is ok.

		var durationFloat float64
		durationFloat, err := strconv.ParseFloat(durationArg, 64)
		if err != nil {
			return err
		}

		// TODO: I want to read the bool flag value of 'capture' and assign it to a variable
		capture := ctx.Bool("capture")
		blocker := !ctx.Bool("no-blocker")

		fmt.Println("## capture (bool):", capture)
		fmt.Println("## blocker (bool):", blocker)

		app.Start(os.Stdout, durationFloat, taskNameArg, blocker, capture, true)

		return nil
	},
}
