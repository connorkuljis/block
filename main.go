package main

import (
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
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
		fmt.Println("# Welcome to task-tracker-cli")
		fmt.Println()

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
	go recordScreen(minutes, cancelCh, finishCh, &wg)

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

func recordScreen(minutes float64, cancelCh, finishCh chan bool, wg *sync.WaitGroup) {
	recordingsDir := "/Users/connor/Code/golang/task-tracker-cli/recordings" // TODO: source this from config file.
	filetype := ".mkv"

	timestamp := time.Now().Format("2006-01-02_15-04-05")

	filename := filepath.Join(recordingsDir, timestamp) + filetype

	cmd := exec.Command("ffmpeg",
		"-f", "avfoundation",
		"-i", "1:0",
		"-pix_fmt", "yuv420p",
		"-r", "25",
		filename,
	)

	fmt.Println("Starting screen capture")
	err := cmd.Start()
	if err != nil {
		panic(err)
	}

	select {
	case <-cancelCh:
		cmd.Process.Signal(syscall.SIGTERM)
	case <-finishCh:
		cmd.Process.Signal(syscall.SIGTERM)
	}

	fmt.Println("\nScreen capture saved to: " + filename)
	wg.Done()
	return
}
