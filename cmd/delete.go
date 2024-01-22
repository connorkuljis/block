package cmd

import (
	"fmt"
	"log"

	"github.com/connorkuljis/task-tracker-cli/tasks"
	"github.com/spf13/cobra"
)

var deleteTaskCmd = &cobra.Command{
	Use:   "delete",
	Short: "Deletes a task by given ID.",
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		err := tasks.DeleteTaskByID(id)
		if err != nil {
			log.Fatal(err)
			return
		}
		fmt.Println("Deleted: ", id)
	},
}
