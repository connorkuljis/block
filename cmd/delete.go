package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

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
