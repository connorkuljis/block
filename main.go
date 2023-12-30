package main

import (
	"fmt"
	"log"
	"strconv"
	"sync"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
)

const defaultDurationInMinutes = 5.0

var rootCmd = &cobra.Command{
	Use:   "block",
	Short: "Block removes distractions when you work on tasks.",
	Long: `Block saves you time by blocking websites at IP level.
Progress bar is displayed directly in the terminal. 
Automatically unblock sites when the task is complete.`,

	Run: func(cmd *cobra.Command, args []string) {
		var minutes = defaultDurationInMinutes // TODO: source from config file.

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

		fmt.Printf("Setting a timer for %.1f minutes.\n", minutes)

		w := DefaultTabWriter()
		b := NewBlocker()

		fmt.Printf("ESC or 'q' to exit. Press any key to pause.\n")

		if enableBocker {
			StartBlockerWrapper(b, w)
		}

		startTime := time.Now()

		startInteractiveTimer(minutes, w)

		if enableBocker {
			StopBlockerWrapper(b, w)
		}

		endTime := time.Now()
		duration := endTime.Sub(startTime)

		fmt.Fprintf(w, "Start time:\t%s\n", startTime.Format("3:04:05pm"))
		fmt.Fprintf(w, "End time:\t%s\n", endTime.Format("3:04:05pm"))
		fmt.Fprintf(w, "Duration:\t%d hours, %d minutes and %d seconds.\n", int(duration.Hours()), int(duration.Minutes())%60, int(duration.Seconds())%60)

		w.Flush()
	},
}

// flags
var (
	taskName             string
	enableBocker         bool
	enableScreenRecorder bool
	verbose              bool
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&taskName, "task", "t", "", "Record optional task name.")
	rootCmd.PersistentFlags().BoolVarP(&enableBocker, "blocker", "b", false, "Enables internet blocker.")
	rootCmd.PersistentFlags().BoolVarP(&enableScreenRecorder, "screen-recorder", "x", false, "Enables screen recorder.")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Logs additional details.")
}

func main() {
	err := rootCmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
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

	w.Flush()

	wg.Add(2)
	go RenderProgressBar(minutes, pauseCh, cancelCh, finishCh, &wg)
	go PollInput(pauseCh, cancelCh, finishCh, &wg)

	wg.Wait()

	if enableScreenRecorder {
		fmt.Fprintf(w, "Screen Recorder:\tstopped\n")
	}
}
