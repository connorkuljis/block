package main

import (
	"database/sql"
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
	cfg         Config
	flags       Flags
)

type Flags struct {
	DisableBlocker bool
	ScreenRecorder bool
	Verbose        bool
}

func main() {
	var err error
	cfg, err = InitConfig()
	if err != nil {
		log.Fatal(err)
	}

	InitDB()

	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(historyCmd)
	rootCmd.AddCommand(deleteTaskCmd)
	rootCmd.AddCommand(resetDNSCommand)

	err = rootCmd.Execute()
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
		// fmt.Printf("Setting a timer for %.1f minutes.\n", duration)

		fmt.Printf("ESC or 'q' to exit. Press any key to pause.\n")

		startTime := time.Now()

		if flags.DisableBlocker {
			startInteractiveTimer()
		} else {
			b := NewBlocker()

			err := b.Block()
			if err != nil {
				log.Fatal(err)
			}

			err = ResetDNS()
			if err != nil {
				log.Fatal(err)
			}

			startInteractiveTimer()

			err = b.Unblock()
			if err != nil {
				log.Fatal(err)
			}
		}

		endTime := time.Now()
		totalTime := endTime.Sub(startTime)

		currentTask.FinishedAt = sql.NullTime{Time: endTime, Valid: true}
		currentTask.ActualDuration = sql.NullFloat64{Float64: totalTime.Minutes(), Valid: true}

		UpdateTask(currentTask)

		if flags.Verbose {
			fmt.Printf("Start time:\t%s\n", startTime.Format("3:04:05pm"))
			fmt.Printf("End time:\t%s\n", endTime.Format("3:04:05pm"))
			fmt.Printf("Duration:\t%d hours, %d minutes and %d seconds.\n", int(totalTime.Hours()), int(totalTime.Minutes())%60, int(totalTime.Seconds())%60)
		}
	},
}

type Args struct {
	Name     string
	Duration float64
}

func stringToFloat(argStr string) interface{} {
	floatval, err := strconv.ParseFloat(argStr, 64)
	if err != nil {
		return err
	}
	return floatval
}

// first arg is either float or string
func parseArgs(args []string) (Args, error) {
	myArgs := Args{}

	if len(args) != 2 {
		return myArgs, fmt.Errorf("Invalid number of arguments, expected 2, recieved: %d", len(args))
	}

	inDuration := args[0]
	inName := args[1]

	duration, err := strconv.ParseFloat(inDuration, 64)
	if err != nil {
		return myArgs, fmt.Errorf("Error converting %s to float. Please provide a valid float.", inDuration)
	}

	myArgs = Args{
		Duration: duration,
		Name:     inName,
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

var resetDNSCommand = &cobra.Command{
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
