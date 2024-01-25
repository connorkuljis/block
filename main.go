package main

import (
	"log"

	"github.com/connorkuljis/block-cli/cmd"
	"github.com/connorkuljis/block-cli/config"
	"github.com/connorkuljis/block-cli/tasks"
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

	if err = cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
