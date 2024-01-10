package main

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/spf13/cobra"
)

type Args struct {
	Name     string
	Duration float64
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
		var duration float64
		var name string

		myArgs, err := parseArgs(args)
		if err != nil {
			cmd.Usage()
			log.Fatal(err)
		}

		duration = myArgs.Duration
		name = myArgs.Name

		currentTask = InsertTask(NewTask(name, duration))
		startTime := time.Now()

		fmt.Printf("ESC or 'q' to exit. Press any key to pause.\n")

		if flags.DisableBlocker {
			start()
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

			start()

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

func parseArgs(args []string) (Args, error) {
	var myArgs = Args{}

	if len(args) != 2 {
		return myArgs, fmt.Errorf("Invalid number of arguments, expected 2, recieved: %d", len(args))
	}

	inDuration := args[0]
	inName := args[1]

	duration, err := strconv.ParseFloat(inDuration, 64)
	if err != nil {
		return myArgs, fmt.Errorf("Error converting %s to float. Please provide a valid float.", inDuration)
	}

	myArgs.Duration = duration
	myArgs.Name = inName

	return myArgs, nil
}
