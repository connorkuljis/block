package cmd

import (
	"log"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

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
