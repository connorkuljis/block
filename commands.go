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
	Task    Task
	Blocker Blocker

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
		duration := 0.0
		name := ""

		// check args
		if len(args) < 1 || len(args) > 2 {
			log.Fatal("Invalid number of arguments. Usage: block start <float> [name]")
		}

		duration, err := strconv.ParseFloat(args[0], 64)
		if err != nil {
			log.Fatal(fmt.Errorf("Invalid argument, error converting %s to float. Please provide a valid float.", args[0]))
		}

		if len(args) == 2 {
			name = args[1]
		}

		// get current time
		createdAt := time.Now()

		// start blocker
		blocker := NewBlocker(flags.DisableBlocker)
		if err = blocker.BlockAndReset(); err != nil {
			log.Print(err)
		}

		// print to standard out
		color.Red("To exit press Control-C. Any key to pause.")
		currentTask := InsertTask(NewTask(name, duration))
		fmt.Printf("Starting a timer for %0.1f minutes\n", currentTask.PlannedDuration)

		// initialise remote
		r := Remote{
			Task:    currentTask,
			Blocker: blocker,
			wg:      &sync.WaitGroup{},
			Pause:   make(chan bool, 1),
			Cancel:  make(chan bool, 1),
			Finish:  make(chan bool, 1),
		}

		// run the configured goroutines
		r.wg.Add(2)
		go RenderProgressBar(r)
		go PollInput(r)

		if flags.ScreenRecorder {
			r.wg.Add(1)
			go FfmpegCaptureScreen(r)
		}

		// wait for the goroutines to finish
		r.wg.Wait()

		// stop blocker
		if err = blocker.UnblockAndReset(); err != nil {
			log.Print(err)
		}

		// calculation
		finishedAt := time.Now()
		actualDuration := finishedAt.Sub(createdAt)

		// persist calculations
		if err = UpdateFinishTimeAndDuration(currentTask, finishedAt, actualDuration); err != nil {
			log.Fatal(err)
		}

		// print to standard out
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
			RenderTable(tasks)
			return
		}

		if len(args) == 1 {
			switch strings.ToLower(args[0]) {
			case "today":
				tasks, err := GetTasksByDate(time.Now())
				if err != nil {
					log.Fatal(err)
				}
				RenderTable(tasks)
				return
			default:
				inDate, err := time.Parse("2006-01-02", args[0])
				if err != nil {
					log.Fatal("Error parsing date: " + args[0])
				}

				tasks, err = GetTasksByDate(inDate)
				if err != nil {
					log.Fatal(err)
				}
				RenderTable(tasks)
				return
			}
		}
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
