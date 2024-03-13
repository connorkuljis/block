package cmd

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/connorkuljis/block-cli/blocker"
	"github.com/connorkuljis/block-cli/interactive"
	"github.com/connorkuljis/block-cli/tasks"
	"github.com/urfave/cli/v2"
)

type App struct {
	hasBlocker        bool
	hasScreenRecorder bool
	hasDebug          bool

	startTime time.Time
	endTime   time.Time

	duration float64
	taskName string

	currentTask tasks.Task
	blocker     blocker.Blocker
}

var RootCmd = &cli.Command{
	Name:  "start",
	Usage: "start the blocker",
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
		var app = App{}
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

// Init contructs the app state from the cli context
func (app *App) Init(ctx *cli.Context) error {
	app.startTime = time.Now()

	// validate args length
	if ctx.NArg() < 1 {
		return errors.New("Error, no arguments provided")
	}

	// parse args into duration and taskname
	duration, taskName, err := parseArgs(ctx)
	if err != nil {
		return err
	}

	// set app options from flags
	app.hasBlocker = true
	if ctx.Bool("no-blocker") {
		app.hasBlocker = false
	}

	app.hasScreenRecorder = ctx.Bool("capture")

	if app.hasBlocker {
		app.blocker = blocker.NewBlocker()
	}

	app.currentTask = tasks.NewTask(taskName, duration, app.hasBlocker, app.hasScreenRecorder)

	return nil
}

func (app *App) Start() error {
	if app.hasBlocker {
		log.Println("starting blocker and resetting dns")

		err := app.blocker.Start()
		if err != nil {
			return err
		}

		log.Println("successfully enabled blocker and reset dns")
	}

	app.currentTask = tasks.InsertTask(app.currentTask)

	remote := interactive.NewRemote(app.currentTask, app.blocker)
	remote.RunTasks(app.hasScreenRecorder)

	log.Println("Finished waiting on goroutines")

	return nil
}

func (app *App) SaveAndExit() error {
	app.endTime = time.Now()
	elapsedTime := app.endTime.Sub(app.startTime)

	err := tasks.UpdateFinishTimeAndDuration(app.currentTask, app.endTime, elapsedTime)
	if err != nil {
		return err
	}

	log.Println("stopping blocker and reset dns")

	if app.hasBlocker {
		err = app.blocker.Stop()
		if err != nil {
			return err
		}
	}

	log.Println("successfully stopped blocker and reset dns")

	fmt.Println("Goodbye.")
	return nil

}

func parseArgs(ctx *cli.Context) (float64, string, error) {
	duration := 0.0
	name := ""

	var err error
	if ctx.NArg() == 1 {
		durationArg := ctx.Args().Get(0)

		duration, err = convertStringToFloat64(durationArg)
		if err != nil {
			return duration, name, err
		}
	}

	if ctx.NArg() >= 2 {
		durationArg := ctx.Args().Get(0)
		taskNameArg := ctx.Args().Get(1)

		duration, err = convertStringToFloat64(durationArg)
		if err != nil {
			return duration, name, err
		}
		name = taskNameArg

	}

	return duration, name, nil
}

func convertStringToFloat64(str string) (float64, error) {
	var floatVal float64
	floatVal, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return floatVal, err
	}

	return floatVal, nil
}
