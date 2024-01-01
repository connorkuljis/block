package main

import (
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
)

func RenderHistory() {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Date", "Name", "Duration", "VOD"})
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetTablePadding("\t") // pad with tabs
	table.SetNoWhiteSpace(true)

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
