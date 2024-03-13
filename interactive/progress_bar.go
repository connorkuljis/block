package interactive

import (
	"fmt"
	"log"
	"time"

	"github.com/connorkuljis/block-cli/tasks"
	"github.com/connorkuljis/block-cli/utils"
	"github.com/schollz/progressbar/v3"
)

func progressBar(max int) *progressbar.ProgressBar {
	return progressbar.NewOptions(max,
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionSetPredictTime(true),
		progressbar.OptionShowCount(),
		progressbar.OptionShowElapsedTimeOnFinish(),
		progressbar.OptionFullWidth(),
		progressbar.OptionShowIts(),
	)
}

func RenderProgressBar(remote *Remote) {
	length := int(remote.Task.PlannedDuration * 60) // convert minutes to seconds.
	bar := progressBar(length)
	ticker := time.NewTicker(time.Second * 1)

	i := 0
	paused := false
	for {
		select {
		case <-remote.Cancel:
			saveBarState(remote.Task, bar)
			remote.Wg.Done()
			return
		case <-remote.Pause:
			paused = !paused
		case <-ticker.C:
			if i == length {
				saveBarState(remote.Task, bar)
				utils.SendNotification()
				close(remote.Finish)
				remote.Wg.Done()
				return
			}

			if !paused {
				bar.Add(1)
				i++
			}
		}
	}
}

func saveBarState(task tasks.Task, bar *progressbar.ProgressBar) {
	fmt.Println()
	percent := bar.State().CurrentPercent * 100
	if err := tasks.UpdateCompletionPercent(task, percent); err != nil {
		log.Print(err)
	}
}
