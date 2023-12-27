package main

import (
	"fmt"
	"log"
	"os/exec"
	"sync"
	"time"

	"github.com/spf13/cobra"
)

const debug = false

var (
	minutes        float64
	disableBlocker bool
)

var rootCmd = &cobra.Command{
	Use:   "block",
	Short: "Block removes distractions when you work on tasks.",
	Long: `Block saves you time by blocking websites at IP level.
		Progress bar is displayed directly in the terminal.
		Automatically unblock sites when the task is complete.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("# Welcome to task-tracker-cli")

		if disableBlocker {
			startTask(minutes)
		} else {
			blocker := NewBlocker()
			blocker.Start()
			startTask(minutes)
			blocker.Stop()
		}
	},
}

func init() {
	rootCmd.PersistentFlags().Float64VarP(&minutes, "minutes", "m", 28.0, "Task duration in minutes.")
	rootCmd.PersistentFlags().BoolVarP(&disableBlocker, "disable-blocker", "d", false, "Disable blocker.")
}

func main() {
	err := rootCmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}

func startTask(minutes float64) {
	fmt.Printf("Starting a timer for %.1f minutes.\n", minutes)

	startTime := time.Now()
	fmt.Println("Start: " + startTime.Format("3:04:05pm"))

	pauseCh := make(chan bool, 1)
	cancelCh := make(chan bool, 1)
	finishCh := make(chan bool, 1)

	wg := sync.WaitGroup{}
	wg.Add(3)

	go RenderProgressBar(minutes, pauseCh, cancelCh, finishCh, &wg)
	go PollInput(pauseCh, cancelCh, finishCh, &wg)
	go FfmpegCaptureScreen(minutes, cancelCh, finishCh, &wg)

	wg.Wait()

	endTime := time.Now()
	duration := time.Now().Sub(startTime)

	fmt.Println("End: " + endTime.Format("3:04:05pm"))
	fmt.Printf("Task time: %d hours, %d minutes and %d seconds.\n", int(duration.Hours()), int(duration.Minutes())%60, int(duration.Seconds())%60)
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
