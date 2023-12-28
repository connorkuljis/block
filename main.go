package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
)

func defaultTabWriter() *tabwriter.Writer {
	output := os.Stdout
	minWidth := 0
	tabWidth := 8
	padding := 4
	padChar := '\t'
	flags := 0

	return tabwriter.NewWriter(
		output,
		minWidth,
		tabWidth,
		padding,
		byte(padChar),
		uint(flags),
	)
}

func startBlockerWrapper(blocker Blocker, w *tabwriter.Writer) {
	blocker.Start()
	fmt.Fprintf(w, "Blocker:\tstarted\n")
}

func stopBlockerWrapper(blocker Blocker, w *tabwriter.Writer) {
	blocker.Stop()
	fmt.Fprintf(w, "Blocker:\tstopped\n")
}

var rootCmd = &cobra.Command{
	Use: "block",

	Short: "Block removes distractions when you work on tasks.",

	Long: `Block saves you time by blocking websites at IP level.
Progress bar is displayed directly in the terminal. 
Automatically unblock sites when the task is complete.`,

	Run: func(cmd *cobra.Command, args []string) {
		w := defaultTabWriter()
		blocker := NewBlocker()

		fmt.Printf("ESC or 'q' to exit. Press any key to pause timer. Starting a timer for %.1f minutes.\n", minutes)

		if enableBocker {
			startBlockerWrapper(blocker, w)
		}

		startTime := time.Now()

		startInteractiveTimer(minutes, w)

		if enableBocker {
			stopBlockerWrapper(blocker, w)
		}

		endTime := time.Now()
		duration := time.Now().Sub(startTime)

		fmt.Fprintf(w, "Start time:\t%s\n", startTime.Format("3:04:05pm"))
		fmt.Fprintf(w, "End time:\t%s\n", endTime.Format("3:04:05pm"))
		fmt.Fprintf(w, "Task time:\t%d hours, %d minutes and %d seconds.\n", int(duration.Hours()), int(duration.Minutes())%60, int(duration.Seconds())%60)

		w.Flush()
	},
}

// flags
var (
	taskName             string
	minutes              float64
	enableBocker         bool
	enableScreenRecorder bool
	verbose              bool
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&taskName, "task", "t", "", "Record optional task name.")
	rootCmd.PersistentFlags().Float64VarP(&minutes, "minutes", "m", 5.0, "Duration of program in minutes")
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

func SendNotification(message string) {
	cmd := exec.Command("terminal-notifier", "-title", "task-tracker-cli", "-sound", "default", "-message", message)
	err := cmd.Start()
	if err != nil {
		fmt.Println(err.Error())
	}
	cmd.Wait()
}
