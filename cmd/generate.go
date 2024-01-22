package cmd

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/connorkuljis/task-tracker-cli/interactive"
	"github.com/connorkuljis/task-tracker-cli/tasks"
	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Concatenate capture recording files into a seperate file.",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			log.Fatal("Invalid arguments, expected either 'today' or [timestamp] in yyyy-mm-dd")
		}

		arg1 := args[0]
		var t time.Time

		if strings.ToLower(arg1) == "today" {
			t = time.Now()
		} else {
			var err error
			t, err = time.Parse("2006-01-02", arg1)
			if err != nil {
				log.Fatal("Error parsing date: " + arg1)
			}
		}

		tasks, err := tasks.GetCapturedTasksByDate(t)
		if err != nil {
			log.Fatal(err)
		}

		var screenCaptureFiles []string
		for _, task := range tasks {
			screenCaptureFile := task.ScreenURL.String
			if screenCaptureFile != "" {
				screenCaptureFiles = append(screenCaptureFiles, screenCaptureFile)
			}
		}

		outfile, err := interactive.FfmpegConcatenateScreenRecordings(t, screenCaptureFiles)
		if err != nil {
			fmt.Println("Unable to concatenate recordings")
			log.Fatal(err)
		}

		fmt.Println("Generated concatenated recording: " + outfile)
	},
}
