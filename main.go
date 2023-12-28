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

// const verbose = false // debug prints additional details

var rootCmd = &cobra.Command{
	Use:   "block",
	Short: "Block removes distractions when you work on tasks.",
	Long: `Block saves you time by blocking websites at IP level.
Progress bar is displayed directly in the terminal. 
Automatically unblock sites when the task is complete.`,
	Run: func(cmd *cobra.Command, args []string) {
		minWidth := 0
		tabWidth := 8
		padding := 2
		padChar := '\t'
		flags := 0
		w := tabwriter.NewWriter(os.Stdout, minWidth, tabWidth, padding, byte(padChar), uint(flags))

		fmt.Fprintf(w, "Starting a timer for %.1f minutes.\n", minutes)

		startTime := time.Now()
		fmt.Fprintf(w, "Start time:\t%s\n", startTime.Format("3:04:05pm"))

		blocker := NewBlocker()

		if !disableBocker {
			blocker.Start()
			fmt.Fprintf(w, "Blocker:\tenabled\n")
		}

		w.Flush()

		startInteractiveTimer(minutes)

		if !disableBocker {
			blocker.Stop()
			fmt.Fprintf(w, "Blocker:\tdisabled\n")
		}

		endTime := time.Now()
		duration := time.Now().Sub(startTime)

		fmt.Fprintf(w, "End time:\t%s\n", endTime.Format("3:04:05pm"))
		fmt.Fprintf(w, "Task time:\t%d hours, %d minutes and %d seconds.\n", int(duration.Hours()), int(duration.Minutes())%60, int(duration.Seconds())%60)

		w.Flush()
	},
}

var (
	taskName              string
	minutes               float64
	disableBocker         bool
	disableScreenRecorder bool
	verbose               bool
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&taskName, "task", "t", "", "Record optional task name.")
	rootCmd.PersistentFlags().Float64VarP(&minutes, "minutes", "m", 5.0, "Duration of program in minutes")
	rootCmd.PersistentFlags().BoolVarP(&disableBocker, "no-blocker", "b", false, "Disables internet blocker.")
	rootCmd.PersistentFlags().BoolVarP(&disableScreenRecorder, "no-screen-recorder", "x", false, "Disables screen recorder.")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Logs additional details.")
}

func main() {
	err := rootCmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}

func startInteractiveTimer(minutes float64) {
	pauseCh := make(chan bool, 1)
	cancelCh := make(chan bool, 1)
	finishCh := make(chan bool, 1)

	wg := sync.WaitGroup{}
	wg.Add(2)

	go RenderProgressBar(minutes, pauseCh, cancelCh, finishCh, &wg)
	go PollInput(pauseCh, cancelCh, finishCh, &wg)

	if !disableScreenRecorder {
		wg.Add(1)
		go FfmpegCaptureScreen(minutes, cancelCh, finishCh, &wg)
	}

	wg.Wait()
}

func SendNotification(message string) {
	fmt.Println("Sending notification.")

	cmd := exec.Command("terminal-notifier", "-title", "task-tracker-cli", "-sound", "default", "-message", message)
	err := cmd.Start()
	if err != nil {
		fmt.Println(err.Error())
	}

	cmd.Wait()
}
