package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strconv"
	"sync"
	"text/tabwriter"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/spf13/cobra"
)

var (
	// flags
	disableBocker        bool
	deleteTaskID         int64
	enableScreenRecorder bool
	verbose              bool

	// globals
	db         *sqlx.DB
	config     UserConfig
	globalArgs Args
)

type Args struct {
	Duration      float64
	TaskName      string
	CurrentTaskID int64
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&disableBocker, "no-block", "d", false, "Do not block hosts file.")
	rootCmd.PersistentFlags().BoolVarP(&enableScreenRecorder, "screen-recorder", "x", false, "Enable screen recorder.")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Logs additional details.")
}

func main() {
	// global state
	InitConfig()
	InitDB()

	rootCmd.AddCommand(startCmd, historyCmd, deleteTaskCmd)

	err := rootCmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
	db.Close()
}

func stringToFloatOrString(arg string) interface{} {
	f, err := strconv.ParseFloat(arg, 64)
	if err != nil {
		return arg
	}
	return f
}

func setGlobalArgs(args []string) error {
	switch len(args) {
	case 0:
		d := config.DefaultDuration
		globalArgs.Duration = d
		fmt.Printf("No arguments provided. Using default value %.1f.\n", d)
	case 1:
		switch val := stringToFloatOrString(args[0]).(type) {
		case float64:
			globalArgs.Duration = val
		case string:
			globalArgs.TaskName = val
		}
	case 2:
		switch val := stringToFloatOrString(args[0]).(type) {
		case float64:
			globalArgs.Duration = val
			globalArgs.TaskName = args[1]
		default:
			return errors.New("Error, given 2 args, must be float followed by string.")
		}
	default:
		return errors.New("Error, given 2 args, must be float followed by string.")
	}

	return nil
}

var rootCmd = &cobra.Command{
	Use:   "block",
	Short: "Block removes distractions when you work on tasks.",
	Long: `
Block saves you time by blocking websites at IP level.
Progress bar is displayed directly in the terminal. 
Automatically unblock sites when the task is complete.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Usage()
	},
}

var startCmd = &cobra.Command{
	Use: "start",
	Run: func(cmd *cobra.Command, args []string) {
		err := setGlobalArgs(args)
		if err != nil {
			log.Fatal(err)
		}

		currentTask := NewTask(globalArgs.TaskName, globalArgs.Duration)
		id := InsertTask(currentTask)
		globalArgs.CurrentTaskID = id

		fmt.Printf("Setting a timer for %.1f minutes.\n", globalArgs.Duration)

		w := DefaultTabWriter()
		b := NewBlocker()

		fmt.Printf("ESC or 'q' to exit. Press any key to pause.\n")

		if !disableBocker {
			StartBlockerWrapper(b, w)
		}

		startTime := time.Now()

		startInteractiveTimer(w)

		if !disableBocker {
			StopBlockerWrapper(b, w)
		}

		endTime := time.Now()
		totalTime := endTime.Sub(startTime)

		currentTask.FinishedAt = sql.NullTime{Time: endTime, Valid: true}
		currentTask.ActualDuration = sql.NullFloat64{Float64: totalTime.Minutes(), Valid: true}

		UpdateTask(currentTask)

		fmt.Fprintf(w, "Start time:\t%s\n", startTime.Format("3:04:05pm"))
		fmt.Fprintf(w, "End time:\t%s\n", endTime.Format("3:04:05pm"))
		fmt.Fprintf(w, "Duration:\t%d hours, %d minutes and %d seconds.\n", int(totalTime.Hours()), int(totalTime.Minutes())%60, int(totalTime.Seconds())%60)

		w.Flush()
	},
}

func startInteractiveTimer(w *tabwriter.Writer) {
	pauseCh := make(chan bool, 1)
	cancelCh := make(chan bool, 1)
	finishCh := make(chan bool, 1)

	wg := sync.WaitGroup{}

	if enableScreenRecorder {
		fmt.Fprintf(w, "Screen Recorder:\tstarted\n")
		wg.Add(1)
		go FfmpegCaptureScreen(w, cancelCh, finishCh, &wg)
	}

	// tabwriter needs all the content in the buffer to calculate padding, thus we flush
	// after the screen recorder check.
	w.Flush()
	// ensure the writer is flushed before starting other goroutines.

	wg.Add(2)
	go RenderProgressBar(pauseCh, cancelCh, finishCh, &wg)
	go PollInput(pauseCh, cancelCh, finishCh, &wg)

	wg.Wait()

	if enableScreenRecorder {
		fmt.Fprintf(w, "Screen Recorder:\tstopped\n")
	}
}

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "Show task history.",
	Run: func(cmd *cobra.Command, args []string) {
		RenderHistory()
	},
}

var deleteTaskCmd = &cobra.Command{
	Use:   "delete",
	Short: "Deletes a task by given ID.",
	Run: func(cmd *cobra.Command, args []string) {
		DeleteTaskByID(args[0])
	},
}
