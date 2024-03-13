package commands

import (
	"errors"
	"fmt"

	"github.com/connorkuljis/block-cli/internal/tasks"
	"github.com/urfave/cli/v2"
)

var DeleteTaskCmd = &cli.Command{
	Name:  "delete",
	Usage: "Deletes a task by given ID.",
	Action: func(ctx *cli.Context) error {
		if ctx.NArg() < 1 {
			return errors.New("Empty arguments")
		}

		id := ctx.Args().Get(0)
		rowsAffected, err := tasks.DeleteTaskByID(id)
		if err != nil {
			return err
		}

		if rowsAffected > 0 {
			fmt.Printf("Successfully deleted element by id: %s, (%d rows affected).\n", id, rowsAffected)
		} else {
			return errors.New("Unable to delete element with id: " + id)
		}

		return nil
	},
}
