package main

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/connorkuljis/block-cli/internal/app"
	"github.com/connorkuljis/block-cli/internal/blocker"
	"github.com/connorkuljis/block-cli/internal/config"
	"github.com/connorkuljis/block-cli/internal/db"
	"github.com/connorkuljis/block-cli/internal/interactive"
	"github.com/connorkuljis/block-cli/internal/server"
	"github.com/connorkuljis/block-cli/internal/tasks"
	"github.com/connorkuljis/block-cli/internal/utils"
	"github.com/jmoiron/sqlx"

	"github.com/urfave/cli/v2"
)

//go:embed www
var embedWebContent embed.FS

func main() {
	err := config.InitConfig()
	if err != nil {
		log.Fatal(err)
	}
	slog.Info("Loaded config.")

	db, err := db.InitDB()
	if err != nil {
		log.Fatal(err)
	}
	slog.Info("Loaded db.")

	app := &cli.App{
		Name:  "block",
		Usage: "block-cli blocks distractions from the command line. track tasks and capture your screen.",
		Before: func(c *cli.Context) error {
			c.Context = context.WithValue(c.Context, "db", db)
			return nil
		},
		Commands: []*cli.Command{
			StartCmd,
			HistoryCmd,
			DeleteTaskCmd,
			ServeCmd,
			GenerateCmd,
			ResetDNSCmd,
			UpCmd,
			DownCmd,
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

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

		fmt.Println("---")
		fmt.Println("Total focus time today ==>", utils.SecsToHHMMSS(totalSecondsToday))
		fmt.Println("Goodbye.")

		return nil
	},
}

var DeleteTaskCmd = &cli.Command{
	Name:  "delete",
	Usage: "Deletes a task by given ID.",
	Action: func(ctx *cli.Context) error {
		db := ctx.Context.Value("db").(*sqlx.DB)

		if ctx.NArg() < 1 {
			return errors.New("Empty arguments")
		}

		id := ctx.Args().Get(0)
		rowsAffected, err := tasks.DeleteTaskByID(db, id)
		if err != nil {
			return err
		}

		if rowsAffected > 0 {
			fmt.Printf("Successfully deleted element by id: %s, (%d rows affected).\n", id, rowsAffected)
		} else {
			return errors.New("Unable to delete element with id: " + id)
		}

		return nil
	},
}

var GenerateCmd = &cli.Command{
	Name:  "generate",
	Usage: "Concatenate capture recording files into a seperate file.",
	Action: func(ctx *cli.Context) error {
		db := ctx.Context.Value("db").(*sqlx.DB)

		if ctx.NArg() < 1 {
			log.Fatal("Invalid arguments, expected either 'today' or [timestamp] in yyyy-mm-dd")
		}

		arg1 := ctx.Args().First()
		var t time.Time
		if strings.ToLower(arg1) == "today" {
			t = time.Now()
		} else {
			var err error
			t, err = time.Parse("2006-01-02", arg1)
			if err != nil {
				return err
			}
		}

		tasks, err := tasks.GetCapturedTasksByDate(db, t)
		if err != nil {
			return err
		}

		var screenCaptureFiles []string
		for _, task := range tasks {
			screenCaptureFile := task.ScreenURL.String
			if screenCaptureFile != "" {
				screenCaptureFiles = append(screenCaptureFiles, screenCaptureFile)
			}
		}

		outfile, err := interactive.FfmpegConcatenateScreenRecordings(t, screenCaptureFiles)
		if err != nil {
			fmt.Println("Unable to concatenate recordings")
			return err
		}

		fmt.Println("Generated concatenated recording: " + outfile)
		return nil
	},
}

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
				inDate, err := time.Parse("2006-01-02", ctx.Args().Get(0))
				if err != nil {
					log.Fatal("Error parsing date: " + ctx.Args().Get(0))
				}

				all, err = tasks.GetTasksByDate(db, inDate)
				if err != nil {
					return err
				}
			}
		}

		tasks.RenderTable(all)

		return nil
	},
}

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

var ServeCmd = &cli.Command{
	Name:  "serve",
	Usage: "Serves http server.",
	Action: func(ctx *cli.Context) error {
		db := ctx.Context.Value("db").(*sqlx.DB)

		templatesPath := "www/templates"
		staticPath := "www/static"
		s, err := server.NewServer(embedWebContent, db, "8080", templatesPath, staticPath)

		err = s.Routes()
		if err != nil {
			return err
		}

		err = s.ListenAndServe()
		if err != nil {
			return err
		}

		return nil
	},
}
