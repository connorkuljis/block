package cmd

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/connorkuljis/task-tracker-cli/blocker"
	"github.com/connorkuljis/task-tracker-cli/interactive"
	"github.com/connorkuljis/task-tracker-cli/tasks"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "block",
	Short: "Expects a duration in minutes, followed by a task name. Eg: block start [duration] [task name]",
	Run: func(cmd *cobra.Command, args []string) {
		disableBlocker := false
		screenRecorder := false
		debug := false
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
		debug, _ = cmd.Root().Flags().GetBool("debug")

		if debug {
			slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})))
		}

		slog.Debug("starting blocker and resetting dns")
		blocker := blocker.NewBlocker(disableBlocker)
		if err = blocker.BlockAndReset(); err != nil {
			slog.Error(err.Error())
		}
		slog.Debug("successfully enabled blocker and reset dns")

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
		slog.Debug("Rendering progress bar")
		go interactive.RenderProgressBar(r)
		slog.Debug("Polling input")
		go interactive.PollInput(r)

		if screenRecorder {
			r.Wg.Add(1)
			slog.Debug("Starting screen recorder")
			go interactive.FfmpegCaptureScreen(r)
		}

		// wait for the goroutines to finish
		r.Wg.Wait()
		slog.Debug("Finished waiting on goroutines")

		// calculation
		finishedAt := time.Now()
		actualDuration := finishedAt.Sub(createdAt)

		// persist calculations
		if err = tasks.UpdateFinishTimeAndDuration(currentTask, finishedAt, actualDuration); err != nil {
			log.Fatal(err)
		}

		slog.Debug("stopping blocker and reset dns")
		if err = blocker.Unblock(); err != nil {
			log.Print(err)
		}
		slog.Debug("successfully stopped blocker and reset dns")

		fmt.Println("Goodbye.")
	},
}

func init() {
	rootCmd.AddCommand(
		historyCmd,
		deleteTaskCmd,
		resetDNSCmd,
		generateCmd,
		serveCmd,
	)

	rootCmd.PersistentFlags().BoolP("noBlock", "f", false, "Do not block hosts file.")
	rootCmd.PersistentFlags().BoolP("screenRecorder", "x", false, "Enable screen recorder.")
	rootCmd.PersistentFlags().Bool("debug", false, "Logs additional details.")

}

func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		return err
	}
	return nil
}
