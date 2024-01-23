package main

import (
	"log"

	"github.com/connorkuljis/task-tracker-cli/cmd"
	"github.com/connorkuljis/task-tracker-cli/config"
	"github.com/connorkuljis/task-tracker-cli/tasks"
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
