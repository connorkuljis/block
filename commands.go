package main

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
	Task Task

	wg     *sync.WaitGroup
	Pause  chan bool
	Cancel chan bool
	Finish chan bool
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
		if len(args) != 2 {
			log.Fatal(fmt.Errorf("Invalid number of arguments, expected 2, recieved: %d", len(args)))
		}

		durationStr := args[0]
		name := args[1]

		duration, err := strconv.ParseFloat(durationStr, 64)
		if err != nil {
			log.Fatal(fmt.Errorf("Error converting %s to float. Please provide a valid float.", durationStr))
		}

		currentTask := InsertTask(NewTask(name, duration))
		createdAt := time.Now()

		var b Blocker
		useBlocker := !flags.DisableBlocker
		if useBlocker {
			b = NewBlocker()
			if err := b.Block(); err != nil {
				log.Fatal(err)
			}
			if err = ResetDNS(); err != nil {
				log.Fatal(err)
			}
		}

		color.Red("ESC or 'q' to exit. Press any key to pause.")

		r := Remote{
			Task:   currentTask,
			wg:     &sync.WaitGroup{},
			Pause:  make(chan bool, 1),
			Cancel: make(chan bool, 1),
			Finish: make(chan bool, 1),
		}

		r.wg.Add(2)
		go RenderProgressBar(r)
		go PollInput(r)

		if flags.ScreenRecorder {
			r.wg.Add(1)
			go FfmpegCaptureScreen(r)
		}

		r.wg.Wait()

		if useBlocker {
			if err = b.Unblock(); err != nil {
				log.Fatal(err)
			}
		}

		finishedAt := time.Now()
		actualDuration := finishedAt.Sub(createdAt)

		if err = UpdateFinishTimeAndDuration(currentTask, finishedAt, actualDuration); err != nil {
			log.Fatal(err)
		}

		if flags.Verbose {
			fmt.Printf("Start time:\t%s\n", createdAt.Format("3:04:05pm"))
			fmt.Printf("End time:\t%s\n", finishedAt.Format("3:04:05pm"))
			fmt.Printf("Duration:\t%d hours, %d minutes and %d seconds.\n", int(actualDuration.Hours()), int(actualDuration.Minutes())%60, int(actualDuration.Seconds())%60)
		}

		fmt.Println("Goodbye.")
	},
}

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "Show task history.",
	Run: func(cmd *cobra.Command, args []string) {
		tasks, err := GetAllTasks()
		if err != nil {
			log.Fatal(err)
		}
		RenderTable(tasks)
	},
}

var deleteTaskCmd = &cobra.Command{
	Use:   "delete",
	Short: "Deletes a task by given ID.",
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		err := DeleteTaskByID(id)
		if err != nil {
			log.Fatal(err)
		}
	},
}

var resetDNSCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset DNS cache.",
	Run: func(cmd *cobra.Command, args []string) {
		flags.Verbose = true
		err := ResetDNS()
		if err != nil {
			log.Println(err)
		}
	},
}

var todayCmd = &cobra.Command{
	Use: "today",
	Run: func(cmd *cobra.Command, args []string) {
		tasks, err := GetTodaysCompletedCapturedTasks()
		if err != nil {
			log.Fatal(err)
		}

		RenderTable(tasks)
	},
}

var timelapseCmd = &cobra.Command{
	Use: "timelapse",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Generating timelapse.")
		tasks, err := GetTodaysCompletedCapturedTasks()
		if err != nil {
			log.Fatal(err)
		}

		if len(tasks) == 0 {
			fmt.Println("Exiting, found no tasks.")
			return
		}

		log.Println("Todays completed captured tasks:")

		var files []string
		for _, task := range tasks {
			files = append(files, task.ScreenURL.String)
		}

		log.Println(files)

		outfile, err := FfmpegGenerateTimelapse(files)
		if err != nil {
			fmt.Println("Unable to generate timelapse")
			log.Fatal(err)
		}

		fmt.Println("Generated timelapse: " + outfile)
	},
}
