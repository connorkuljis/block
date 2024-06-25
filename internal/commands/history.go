package commands

import (
	"log"
	"strings"
	"time"

	"github.com/connorkuljis/block-cli/internal/tasks"
	"github.com/jmoiron/sqlx"
	"github.com/urfave/cli/v2"
)

var HistoryCmd = &cli.Command{
	Name:  "history",
	Usage: "display task history.",
	Action: func(ctx *cli.Context) error {
		db := ctx.Context.Value("db").(*sqlx.DB)

		var all []tasks.Task

		if ctx.NArg() == 0 {
			var err error
			all, err = tasks.GetAllTasks(db)
			if err != nil {
				log.Fatal(err)
			}
		}

		if ctx.NArg() == 1 {
			switch strings.ToLower(ctx.Args().Get(0)) {
			case "today":
				var err error
				all, err = tasks.GetTasksByDate(db, time.Now())
				if err != nil {
					return err
				}
			default:
				indate, err := time.Parse("2006-01-02", ctx.Args().Get(0))
				if err != nil {
					log.Fatal("error parsing date: " + ctx.Args().Get(0))
				}

				all, err = tasks.GetTasksByDate(db, indate)
				if err != nil {
					return err
				}
			}
		}

		tasks.RenderTable(all)

		return nil
	},
}
