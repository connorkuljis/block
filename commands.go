package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
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
	Use:   "start",
	Short: "Expects a duration in minutes, followed by a task name. Eg: block start [duration] [task name]",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 2 {
			log.Fatal(fmt.Errorf("Invalid number of arguments, expected 2, recieved: %d", len(args)))
		}

		durationStr := args[0]
		name := args[1]

		duration, err := strconv.ParseFloat(durationStr, 64)
		if err != nil {
			log.Fatal(fmt.Errorf("Invalid argument, error converting %s to float. Please provide a valid float.", durationStr))
		}

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

		currentTask := InsertTask(NewTask(name, duration))

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
			RenderTable([]Task{currentTask})
		}

		fmt.Println("Goodbye.")
	},
}

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "Show task history.",
	Run: func(cmd *cobra.Command, args []string) {
		var tasks []Task

		if len(args) == 0 {
			var err error
			tasks, err = GetAllTasks()
			if err != nil {
				log.Fatal(err)
			}
		} else if len(args) == 1 {
			arg1 := args[0]
			if strings.ToLower(arg1) == "today" {
				var err error
				tasks, err = GetTasksByDate(time.Now())
				if err != nil {
					log.Fatal(err)
				}
			} else {
				inDate, err := time.Parse("2006-01-02", arg1)
				if err != nil {
					log.Fatal("Error parsing date: " + arg1)
				}
				tasks, err = GetTasksByDate(inDate)
				if err != nil {
					log.Fatal(err)
				}
			}
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
			return
		}
		fmt.Println("Deleted: ", id)
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

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Concatenate capture recording files into a seperate file.",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			log.Fatal("Invalid arguments, expected either 'today' or [timestamp] in yyyy-mm-dd")
		}

		arg1 := args[0]
		var t time.Time

		if strings.ToLower(arg1) == "today" {
			t = time.Now()
		} else {
			var err error
			t, err = time.Parse("2006-01-02", arg1)
			if err != nil {
				log.Fatal("Error parsing date: " + arg1)
			}
		}

		tasks, err := GetCapturedTasksByDate(t)
		if err != nil {
			log.Fatal(err)
		}

		var screenCaptureFiles []string
		for _, task := range tasks {
			screenCaptureFile := task.ScreenURL.String
			if screenCaptureFile != "" {
				screenCaptureFiles = append(screenCaptureFiles, screenCaptureFile)
			}
		}

		outfile, err := FfmpegConcatenateScreenRecordings(t, screenCaptureFiles)
		if err != nil {
			fmt.Println("Unable to concatenate recordings")
			log.Fatal(err)
		}

		fmt.Println("Generated concatenated recording: " + outfile)
	},
}
