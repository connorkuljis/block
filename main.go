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

	err = tasks.InitDB()
	if err != nil {
		log.Fatal(err)
	}

	// if err = cmd.Execute(); err != nil {
	// 	log.Fatal(err)
	// }

	app := &cli.App{
		Commands: []*cli.Command{cmd.RootCmd},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
