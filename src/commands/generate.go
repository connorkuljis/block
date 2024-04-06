package commands

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/connorkuljis/block-cli/src/interactive"
	"github.com/connorkuljis/block-cli/src/tasks"
	"github.com/urfave/cli/v2"
)

var GenerateCmd = &cli.Command{
	Name:  "generate",
	Usage: "Concatenate capture recording files into a seperate file.",
	Action: func(ctx *cli.Context) error {
		if ctx.NArg() < 1 {
			log.Fatal("Invalid arguments, expected either 'today' or [timestamp] in yyyy-mm-dd")
		}

		arg1 := ctx.Args().First()
		var t time.Time
		if strings.ToLower(arg1) == "today" {
			t = time.Now()
		} else {
			var err error
			t, err = time.Parse("2006-01-02", arg1)
			if err != nil {
				return err
			}
		}

		tasks, err := tasks.GetCapturedTasksByDate(t)
		if err != nil {
			return err
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
			return err
		}

		fmt.Println("Generated concatenated recording: " + outfile)
		return nil
	},
}
