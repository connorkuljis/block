package main

import (
	"log"
	"os"

	"github.com/connorkuljis/block-cli/cmd"
	"github.com/connorkuljis/block-cli/config"
	"github.com/connorkuljis/block-cli/tasks"
	"github.com/urfave/cli/v2"
)

func main() {
	err := config.InitConfig()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Loaded config.")

	err = tasks.InitDB()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Loaded db.")

	app := &cli.App{
		Name:  "block",
		Usage: "block-cli blocks distractions from the command line. track tasks and capture your screen.",
		Commands: []*cli.Command{
			cmd.RootCmd,
			cmd.HistoryCmd,
			cmd.DeleteTaskCmd,
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
