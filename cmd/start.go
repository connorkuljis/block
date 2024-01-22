package cmd

import (
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

type Remote struct {
	Task    Task
	Blocker Blocker

	wg     *sync.WaitGroup
	Pause  chan bool
	Cancel chan bool
	Finish chan bool
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Expects a duration in minutes, followed by a task name. Eg: block start [duration] [task name]",
	Run: func(cmd *cobra.Command, args []string) {
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

		// start blocker
		blocker := NewBlocker(flags.DisableBlocker)
		if err = blocker.BlockAndReset(); err != nil {
			log.Print(err)
		}

		// print to standard out
		color.Red("To exit press Control-C. Any key to pause.")
		currentTask := InsertTask(NewTask(name, duration))
		fmt.Printf("Starting a timer for %0.1f minutes\n", currentTask.PlannedDuration)

		// initialise remote
		r := Remote{
			Task:    currentTask,
			Blocker: blocker,
			wg:      &sync.WaitGroup{},
			Pause:   make(chan bool, 1),
			Cancel:  make(chan bool, 1),
			Finish:  make(chan bool, 1),
		}

		// run the configured goroutines
		r.wg.Add(2)
		go RenderProgressBar(r)
		go PollInput(r)

		if flags.ScreenRecorder {
			r.wg.Add(1)
			go FfmpegCaptureScreen(r)
		}

		// wait for the goroutines to finish
		r.wg.Wait()

		// stop blocker
		if err = blocker.UnblockAndReset(); err != nil {
			log.Print(err)
		}

		// calculation
		finishedAt := time.Now()
		actualDuration := finishedAt.Sub(createdAt)

		// persist calculations
		if err = UpdateFinishTimeAndDuration(currentTask, finishedAt, actualDuration); err != nil {
			log.Fatal(err)
		}

		// print to standard out
		if flags.Verbose {
			RenderTable([]Task{currentTask})
		}

		fmt.Println("Goodbye.")
	},
}
