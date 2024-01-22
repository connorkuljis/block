package cmd

import (
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/connorkuljis/task-tracker-cli/blocker"
	"github.com/connorkuljis/task-tracker-cli/interactive"
	"github.com/connorkuljis/task-tracker-cli/tasks"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Expects a duration in minutes, followed by a task name. Eg: block start [duration] [task name]",
	Run: func(cmd *cobra.Command, args []string) {
		disableBlocker := false
		screenRecorder := false
		verbose := false
		duration := 0.0
		name := ""

		// check args
		if len(args) < 1 || len(args) > 2 {
			log.Fatal("Invalid number of arguments. Usage: block start <float> [name]")
		}

		duration, err := strconv.ParseFloat(args[0], 64)
		if err != nil {
			log.Fatal(fmt.Errorf("Invalid argument, error converting %s to float. Please provide a valid float.", args[0]))
		}

		if len(args) == 2 {
			name = args[1]
		}

		// get current time
		createdAt := time.Now()

		disableBlocker, _ = cmd.Root().Flags().GetBool("noBlock")
		screenRecorder, _ = cmd.Root().Flags().GetBool("screenRecorder")
		verbose, _ = cmd.Root().Flags().GetBool("verbose")

		if verbose {
			log.Println("Disable blocker: ", disableBlocker)
			log.Println("Screen recorder: ", screenRecorder)
		}

		// start blocker
		blocker := blocker.NewBlocker(disableBlocker)
		if err = blocker.BlockAndReset(); err != nil {
			log.Print(err)
		}

		// print to standard out
		color.Red("To exit press Control-C. Any key to pause.")
		currentTask := tasks.InsertTask(tasks.NewTask(name, duration, !disableBlocker, screenRecorder))
		fmt.Printf("Starting a timer for %0.1f minutes\n", currentTask.PlannedDuration)

		// initialise remote
		r := interactive.Remote{
			Task:    currentTask,
			Blocker: blocker,
			Wg:      &sync.WaitGroup{},
			Pause:   make(chan bool, 1),
			Cancel:  make(chan bool, 1),
			Finish:  make(chan bool, 1),
		}

		// run the configured goroutines
		r.Wg.Add(2)
		go interactive.RenderProgressBar(r)
		go interactive.PollInput(r)

		if screenRecorder {
			r.Wg.Add(1)
			go interactive.FfmpegCaptureScreen(r)
		}

		// wait for the goroutines to finish
		r.Wg.Wait()

		// stop blocker
		if err = blocker.UnblockAndReset(); err != nil {
			log.Print(err)
		}

		// calculation
		finishedAt := time.Now()
		actualDuration := finishedAt.Sub(createdAt)

		// persist calculations
		if err = tasks.UpdateFinishTimeAndDuration(currentTask, finishedAt, actualDuration); err != nil {
			log.Fatal(err)
		}

		fmt.Println("Goodbye.")
	},
}
