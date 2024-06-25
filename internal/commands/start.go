package commands

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/connorkuljis/block-cli/internal/app"
	"github.com/connorkuljis/block-cli/internal/tasks"
	"github.com/connorkuljis/block-cli/internal/utils"
	"github.com/jmoiron/sqlx"
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
		&cli.Int64Flag{
			Name:    "bucket",
			Aliases: []string{"b"},
			Usage:   "Tag a task with bucket id",
		},
	},
	Action: func(ctx *cli.Context) error {
		db := ctx.Context.Value("db").(*sqlx.DB)
		// sqlx.DB

		if ctx.NArg() < 1 {
			return errors.New("Error, no arguments provided")
		}

		argDurationMinutes := ctx.Args().Get(0)
		argTaskName := ctx.Args().Get(1) // empty string is ok.

		var floatDurationMinutes float64
		floatDurationMinutes, err := strconv.ParseFloat(argDurationMinutes, 64)
		if err != nil {
			return err
		}

		durationSeconds := int64(floatDurationMinutes * 60)

		capture := ctx.Bool("capture")
		blocker := !ctx.Bool("no-blocker")
		bucketId := ctx.Int64("bucket")

		currentTask := tasks.NewTask(argTaskName, durationSeconds, blocker, capture, time.Now())

		if bucketId != 0 {
			currentTask.AddBucketTag(bucketId)
		}

		err = app.Start(os.Stdout, db, *currentTask)
		if err != nil {
			log.Fatal(err)
		}

		var totalSecondsToday int64
		today := currentTask.CreatedAt.Truncate(24 * time.Hour)
		tasks, _ := tasks.GetRecentTasks(db, today, 0)
		for _, task := range tasks {
			totalSecondsToday += task.ActualDurationSeconds.Int64
		}

		// take a break for 1/3 of time worked.
		var breakRatio float64
		var totalBreakSecondsToday int64
		breakRatio = 1.0 / 3
		totalBreakSecondsToday = int64(float64(totalSecondsToday) * breakRatio)

		fmt.Println("---")
		fmt.Println("Total focus time today ==>", utils.SecsToHHMMSS(totalSecondsToday))
		fmt.Println("Cumulative break time today ==>", utils.SecsToHHMMSS(totalBreakSecondsToday))
		fmt.Println("Goodbye.")

		return nil
	},
}
