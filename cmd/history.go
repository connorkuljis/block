package cmd

import (
	"log"
	"strings"
	"time"

	"github.com/connorkuljis/block-cli/tasks"
	"github.com/urfave/cli/v2"
)

var HistoryCmd = &cli.Command{
	Name:  "history",
	Usage: "display task history.",
	Action: func(ctx *cli.Context) error {
		var all []tasks.Task

		if ctx.NArg() == 0 {
			var err error
			all, err = tasks.GetAllTasks()
			if err != nil {
				log.Fatal(err)
			}
		}

		if ctx.NArg() == 1 {
			switch strings.ToLower(ctx.Args().Get(0)) {
			case "today":
				var err error
				all, err = tasks.GetTasksByDate(time.Now())
				if err != nil {
					return err
				}
			default:
				inDate, err := time.Parse("2006-01-02", ctx.Args().Get(0))
				if err != nil {
					log.Fatal("Error parsing date: " + ctx.Args().Get(0))
				}

				all, err = tasks.GetTasksByDate(inDate)
				if err != nil {
					return err
				}
			}
		}

		tasks.RenderTable(all)

		return nil
	},
}
