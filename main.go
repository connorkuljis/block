package main

import (
	"database/sql"
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
	taskName             string
	disableBocker        bool
	showHistory          bool
	enableScreenRecorder bool
	verbose              bool

	// globals
	db     *sqlx.DB
	config UserConfig
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&taskName, "task", "t", "", "Record optional task name.")
	rootCmd.PersistentFlags().BoolVar(&disableBocker, "no-block", false, "Disables internet blocker.")
	rootCmd.PersistentFlags().BoolVar(&showHistory, "history", false, "Display history table.")
	rootCmd.PersistentFlags().BoolVarP(&enableScreenRecorder, "screen-recorder", "x", false, "Enables screen recorder.")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Logs additional details.")
}

func main() {

	// global state
	initConfig()
	initDB()
	// end global state
	// displayTable()

	err := rootCmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
	db.Close()
}

var rootCmd = &cobra.Command{
	Use:   "block",
	Short: "Block removes distractions when you work on tasks.",
	Long: `
Block saves you time by blocking websites at IP level.
Progress bar is displayed directly in the terminal. 
Automatically unblock sites when the task is complete.`,
	Run: func(cmd *cobra.Command, args []string) {
		if showHistory {
			renderHistory()
			return
		}

		minutes := config.DefaultDuration

		if len(args) == 0 {
			fmt.Printf("No arguments provided. Using default value %.1f.\n", minutes)
		} else {
			s := args[0]
			f, err := strconv.ParseFloat(s, 64)
			if err != nil {
				fmt.Printf("Error parsing argument '%s' to float. Using default value %.1f.\n", s, minutes)
			} else {
				minutes = f
			}
		}

		currentTask := NewTask(taskName, minutes)
		id := InsertTask(currentTask)
		currentTask.ID = id

		fmt.Printf("Setting a timer for %.1f minutes.\n", minutes)

		w := DefaultTabWriter()
		b := NewBlocker()

		fmt.Printf("ESC or 'q' to exit. Press any key to pause.\n")

		if !disableBocker {
			StartBlockerWrapper(b, w)
		}

		startTime := time.Now()

		startInteractiveTimer(minutes, w)

		if !disableBocker {
			StopBlockerWrapper(b, w)
		}

		endTime := time.Now()
		duration := endTime.Sub(startTime)

		currentTask.FinishedAt = sql.NullTime{Time: endTime, Valid: true}
		currentTask.ActualDuration = sql.NullFloat64{Float64: duration.Minutes(), Valid: true}

		UpdateTask(currentTask)

		fmt.Fprintf(w, "Start time:\t%s\n", startTime.Format("3:04:05pm"))
		fmt.Fprintf(w, "End time:\t%s\n", endTime.Format("3:04:05pm"))
		fmt.Fprintf(w, "Duration:\t%d hours, %d minutes and %d seconds.\n", int(duration.Hours()), int(duration.Minutes())%60, int(duration.Seconds())%60)

		w.Flush()
	},
}

func startInteractiveTimer(minutes float64, w *tabwriter.Writer) {
	pauseCh := make(chan bool, 1)
	cancelCh := make(chan bool, 1)
	finishCh := make(chan bool, 1)

	wg := sync.WaitGroup{}

	if enableScreenRecorder {
		fmt.Fprintf(w, "Screen Recorder:\tstarted\n")
		wg.Add(1)
		go FfmpegCaptureScreen(minutes, w, cancelCh, finishCh, &wg)
	}

	// tabwriter needs all the content in the buffer to calculate padding, thus we flush
	// after the screen recorder check.
	w.Flush()
	// ensure the writer is flushed before starting other goroutines.

	wg.Add(2)
	go RenderProgressBar(minutes, pauseCh, cancelCh, finishCh, &wg)
	go PollInput(pauseCh, cancelCh, finishCh, &wg)

	wg.Wait()

	if enableScreenRecorder {
		fmt.Fprintf(w, "Screen Recorder:\tstopped\n")
	}
}
