package main

import (
	"log"
	"os"

	"github.com/connorkuljis/block-cli/cmd/block/commands"
	"github.com/connorkuljis/block-cli/internal/config"
	"github.com/connorkuljis/block-cli/internal/tasks"
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
			commands.RootCmd,
			commands.HistoryCmd,
			commands.DeleteTaskCmd,
			commands.ServeCmd,
			commands.GenerateCmd,
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
