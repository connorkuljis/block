package cmd

import (
	"log"
	"strings"
	"time"

	"github.com/connorkuljis/task-tracker-cli/interactive"
	"github.com/connorkuljis/task-tracker-cli/tasks"
	"github.com/spf13/cobra"
)

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "Show task history.",
	Run: func(cmd *cobra.Command, args []string) {
		var all []tasks.Task

		if len(args) == 0 {
			var err error
			all, err = tasks.GetAllTasks()
			if err != nil {
				log.Fatal(err)
			}
			interactive.RenderTable(all)
			return
		}

		if len(args) == 1 {
			switch strings.ToLower(args[0]) {
			case "today":
				all, err := tasks.GetTasksByDate(time.Now())
				if err != nil {
					log.Fatal(err)
				}
				interactive.RenderTable(all)
				return
			default:
				inDate, err := time.Parse("2006-01-02", args[0])
				if err != nil {
					log.Fatal("Error parsing date: " + args[0])
				}

				all, err = tasks.GetTasksByDate(inDate)
				if err != nil {
					log.Fatal(err)
				}
				interactive.RenderTable(all)
				return
			}
		}
	},
}
