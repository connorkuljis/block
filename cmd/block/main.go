package main

import (
	"embed"
	"errors"
	"fmt"
	"github.com/connorkuljis/block-cli/internal/app"
	"github.com/connorkuljis/block-cli/internal/blocker"
	"github.com/connorkuljis/block-cli/internal/config"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/connorkuljis/block-cli/interactive"
	"github.com/connorkuljis/block-cli/server"
	"github.com/connorkuljis/block-cli/tasks"
	"github.com/urfave/cli/v2"
)

//go:embed www
var embedWebContent embed.FS

func main() {
	err := config.InitConfig()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Loaded config.")

	err = tasks.InitDB()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Loaded db.")

	app := &cli.App{
		Name:  "block",
		Usage: "block-cli blocks distractions from the command line. track tasks and capture your screen.",
		Commands: []*cli.Command{
			StartCmd,
			HistoryCmd,
			DeleteTaskCmd,
			ServeCmd,
			GenerateCmd,
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

var DeleteTaskCmd = &cli.Command{
	Name:  "delete",
	Usage: "Deletes a task by given ID.",
	Action: func(ctx *cli.Context) error {
		if ctx.NArg() < 1 {
			return errors.New("Empty arguments")
		}

		id := ctx.Args().Get(0)
		rowsAffected, err := tasks.DeleteTaskByID(id)
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

		tasks, err := tasks.GetCapturedTasksByDate(t)
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
		s, err := server.NewServer(embedWebContent, "8080")

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

// var timerCmd = &cobra.Command{
// 	Use: "timer",
// 	Run: func(cmd *cobra.Command, args []string) {
// 		fmt.Println("timer")
// 		// no args
// 		createdAt := time.Now()

// 		currentTask := tasks.InsertTask(tasks.NewTask("test", -1, true, false))

// 		timer(currentTask)

// 		finishedAt := time.Now()
// 		actualDuration := finishedAt.Sub(createdAt)

// 		tasks.UpdateCompletionPercent(currentTask, -1)

// 		// persist calculations
// 		if err := tasks.UpdateFinishTimeAndDuration(currentTask, finishedAt, actualDuration); err != nil {
// 			log.Fatal(err)
// 		}
// 	},
// }

// func timer(currentTask tasks.Task) {
// 	blocker := blocker.NewHostsBlocker()
// 	err := blocker.Start()
// 	if err != nil {
// 		log.Println(err)
// 	}
// 	// initialise remote
// 	r := interactive.Remote{
// 		Task:    currentTask,
// 		Blocker: blocker,
// 		Wg:      &sync.WaitGroup{},
// 		Pause:   make(chan bool, 1),
// 		Cancel:  make(chan bool, 1),
// 		Finish:  make(chan bool, 1),
// 	}

// 	r.Wg.Add(2)
// 	go interactive.PollInput(r)
// 	go incrementer(r)
// 	r.Wg.Wait()

// 	err = blocker.Stop()
// 	if err != nil {
// 		log.Println(err)
// 	}
// }

// func incrementer(r interactive.Remote) {
// 	ticker := time.NewTicker(time.Second * 1)

// 	i := 0
// 	paused := false
// 	for {
// 		select {
// 		case <-r.Finish:
// 			r.Wg.Done()
// 			return
// 		case <-r.Cancel:
// 			r.Wg.Done()
// 			return
// 		case <-r.Pause:
// 			paused = !paused
// 		case <-ticker.C:
// 			if !paused {
// 				fmt.Printf("%d seconds\n", i)
// 				i++
// 			}
// 		}
// 	}
// }
