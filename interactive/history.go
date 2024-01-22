package interactive

import (
	"fmt"
	"os"

	"github.com/connorkuljis/task-tracker-cli/tasks"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
)

func RenderTable(tasks []tasks.Task) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Date", "Name", "Planned (min)", "Actual (min)", "Completion Percent", "Completed"})
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

	totalMinutes := 0.0
	for _, task := range tasks {
		id := fmt.Sprint(task.ID)
		name := task.Name
		planned := fmt.Sprintf("%.0f", task.PlannedDuration)
		actual := fmt.Sprintf("%.0f", task.ActualDuration.Float64)
		date := task.CreatedAt.Format("Mon Jan 02 15:04:05")

		completionPercent := fmt.Sprintf("%.2f%%", task.CompletionPercent.Float64)

		var completed string
		if task.Completed == 1 {
			completed = "✅"
		}

		row := []string{id, date, name, planned, actual, completionPercent, completed}

		if task.ActualDuration.Valid {
			totalMinutes += task.ActualDuration.Float64
		}

		table.Append(row)
	}
	table.Render()

	fmt.Println()
	color.Cyan(fmt.Sprintf("Total: %.0f minutes", totalMinutes))
}
