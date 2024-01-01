package main

import (
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
)

func renderHistory() {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Date", "Name", "Duration", "VOD"})

	tasks := GetAllTasks()

	for _, task := range tasks {
		id := fmt.Sprint(task.ID)
		name := task.Name
		duration := fmt.Sprintf("%.2f", task.ActualDuration.Float64)
		date := task.CreatedAt.Format("Mon Jan 02 15:04:05")
		vod := task.ScreenURL.String

		row := []string{id, date, name, duration, vod}

		table.Append(row)
	}

	table.Render()
}
