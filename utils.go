package main

import (
	"fmt"
	"os"
	"os/exec"
	"text/tabwriter"
)

func SendNotification(message string) {
	cmd := exec.Command("terminal-notifier", "-title", "task-tracker-cli", "-sound", "default", "-message", message)
	err := cmd.Start()
	if err != nil {
		fmt.Println(err.Error())
	}
	cmd.Wait()
}

func DefaultTabWriter() *tabwriter.Writer {
	output := os.Stdout
	minWidth := 0
	tabWidth := 8
	padding := 4
	padChar := '\t'
	flags := 0

	return tabwriter.NewWriter(
		output,
		minWidth,
		tabWidth,
		padding,
		byte(padChar),
		uint(flags),
	)
}
