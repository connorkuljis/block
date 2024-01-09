package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/spf13/cobra"
)

var (
	db          *sqlx.DB
	currentTask *Task
	config      Config
	flags       Flags
)

type Flags struct {
	DisableBlocker bool
	ScreenRecorder bool
	Verbose        bool
}

type Args struct {
	Name     string
	Duration float64
}

func main() {
	InitConfig()
	InitDB()

	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(historyCmd)
	rootCmd.AddCommand(deleteTaskCmd)

	err := rootCmd.Execute()
	if err != nil {
		log.Fatal(err)
	}

	db.Close()
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&flags.DisableBlocker, "no-block", "d", false, "Do not block hosts file.")
	rootCmd.PersistentFlags().BoolVarP(&flags.ScreenRecorder, "screen-recorder", "x", false, "Enable screen recorder.")
	rootCmd.PersistentFlags().BoolVarP(&flags.Verbose, "verbose", "v", false, "Logs additional details.")
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
		myArgs, err := parseArgs(args)
		if err != nil {
			cmd.Usage()
			log.Fatal(err)
		}

		name := myArgs.Name
		duration := myArgs.Duration

		currentTask = InsertTask(NewTask(name, duration))
		fmt.Printf("Setting a timer for %.1f minutes.\n", duration)

		b := NewBlocker()

		fmt.Printf("ESC or 'q' to exit. Press any key to pause.\n")

		startTime := time.Now()

		if flags.DisableBlocker {
			startInteractiveTimer()
		} else {
			b.Start()
			startInteractiveTimer()
			b.Stop()
		}

		endTime := time.Now()
		totalTime := endTime.Sub(startTime)

		currentTask.FinishedAt = sql.NullTime{Time: endTime, Valid: true}
		currentTask.ActualDuration = sql.NullFloat64{Float64: totalTime.Minutes(), Valid: true}

		UpdateTask(currentTask)

		fmt.Printf("Start time:\t%s\n", startTime.Format("3:04:05pm"))
		fmt.Printf("End time:\t%s\n", endTime.Format("3:04:05pm"))
		fmt.Printf("Duration:\t%d hours, %d minutes and %d seconds.\n", int(totalTime.Hours()), int(totalTime.Minutes())%60, int(totalTime.Seconds())%60)
	},
}

// first arg is either float or string
func parseArgs(args []string) (Args, error) {
	var myArgs Args

	stringToFloatOrString := func(arg string) interface{} {
		f, err := strconv.ParseFloat(arg, 64)
		if err != nil {
			return arg
		}
		return f
	}

	if len(args) == 0 {
		return myArgs, errors.New("Error, at least one argument is required.")
	}

	firstArg := args[0]
	switch v := stringToFloatOrString(firstArg).(type) {
	case float64:
		myArgs.Duration = v
	case string:
		myArgs.Name = v
	}

	if len(args) > 1 {
		nextArg := args[1]
		myArgs.Name = nextArg
	}

	if myArgs.Duration == 0.0 {
		fmt.Println("No duration provided. Using default value")
		myArgs.Duration = config.DefaultDuration
	}

	return myArgs, nil
}

type Remote struct {
	Pause  chan bool
	Cancel chan bool
	Finish chan bool
	wg     *sync.WaitGroup
}

func startInteractiveTimer() {
	r := Remote{
		Pause:  make(chan bool, 1),
		Cancel: make(chan bool, 1),
		Finish: make(chan bool, 1),
		wg:     &sync.WaitGroup{},
	}

	if flags.ScreenRecorder {
		r.wg.Add(3)

		fmt.Printf("Screen Recorder:\tstarted\n")

		go FfmpegCaptureScreen(r)
		go RenderProgressBar(r)
		go PollInput(r)

		r.wg.Wait()

		fmt.Printf("Screen Recorder:\tstopped\n")
	} else {
		r.wg.Add(2)

		go RenderProgressBar(r)
		go PollInput(r)

		r.wg.Wait()
	}
}

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "Show task history.",
	Run: func(cmd *cobra.Command, args []string) {
		RenderHistory()
	},
}

var deleteTaskCmd = &cobra.Command{
	Use:   "delete",
	Short: "Deletes a task by given ID.",
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		DeleteTaskByID(id)
	},
}
