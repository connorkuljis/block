package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/spf13/cobra"
)

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
		fmt.Println("# Welcome to task-tracker-cli #")

		if disableBlocker {
			startTask(minutes)
		} else {
			blocker := NewBlocker()
			blocker.Start()
			startTask(minutes)
			blocker.Stop()
		}

		say("Goodbye")
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
	go say(fmt.Sprintf("Starting a timer for %.0f minutes.", minutes))

	startTime := time.Now()
	fmt.Println("Start: " + startTime.Format("3:04:05pm"))

	wg := sync.WaitGroup{}
	wg.Add(3)

	pauseCh := make(chan bool, 1)
	cancelCh := make(chan bool, 1)
	finishCh := make(chan bool, 1)

	go RenderProgressBar(minutes, pauseCh, cancelCh, finishCh, &wg)
	go PollInput(pauseCh, cancelCh, finishCh, &wg)
	go recordScreen(cancelCh, finishCh, &wg)

	wg.Wait()

	endTime := time.Now()
	duration := time.Now().Sub(startTime)

	fmt.Printf("End: " + endTime.Format("3:04:05pm"))
	fmt.Printf(", %dh, %dm, %ds\n", int(duration.Hours()), int(duration.Minutes())%60, int(duration.Seconds())%60)
}

func say(msg string) {
	cmd := exec.Command("say", "-v", "Bubbles", msg)
	err := cmd.Run()
	if err != nil {
		fmt.Println(err.Error())
	}
}

func recordScreen(cancelCh, finishCh chan bool, wg *sync.WaitGroup) {
	volume := "."

	currentTime := time.Now()

	timestamp := currentTime.Format("2006-01-02_15-04-05")

	file := filepath.Join(volume, timestamp) + ".mp4"

	cmd := exec.Command("screencapture", "-v", file)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Start()

	select {
	case <-cancelCh:
		wg.Done()
		return
	case <-finishCh:
		cmd.Wait()
		wg.Done()
		return
	}
}
