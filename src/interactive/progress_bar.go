package interactive

import (
	"io"
	"time"

	"github.com/connorkuljis/block-cli/src/utils"
	"github.com/schollz/progressbar/v3"
)

func initProgressBar(max int, w io.Writer) *progressbar.ProgressBar {
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
	durationSeconds := int(remote.Task.PlannedDuration * 60) // convert minutes to seconds.

	pbar := initProgressBar(durationSeconds, remote.W)

	ticker := time.NewTicker(time.Second * 1)

	i := 0
	paused := false
	for {
		select {
		case <-remote.Cancel:
			remote.TotalTimeSeconds <- i
			remote.CompletionPercent <- pbar.State().CurrentPercent * 100
			remote.Wg.Done()
			return
		case <-remote.Pause:
			paused = !paused
		case <-ticker.C:
			if i == durationSeconds {
				remote.TotalTimeSeconds <- i
				remote.CompletionPercent <- pbar.State().CurrentPercent * 100
				utils.SendNotification()
				close(remote.Finish)
				remote.Wg.Done()
				return
			}

			if !paused {
				pbar.Add(1)
				i++
			}
		}
	}
}
