package main

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/schollz/progressbar/v3"
)

func progressBar(max int) *progressbar.ProgressBar {
	return progressbar.NewOptions(max,
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionSetPredictTime(false),
		progressbar.OptionShowCount(),
		progressbar.OptionShowElapsedTimeOnFinish(),
		progressbar.OptionFullWidth(),
		progressbar.OptionShowIts(),
	)
}

func completeProgressBar() {
	fmt.Println()
}

func RenderProgressBar(r Remote) {
	calculateTotalTicks := func(minutes float64, tickIntervalMs int) int {
		return int((minutes * 60 * 1000) / float64(tickIntervalMs))
	}

	ticksPerSeconds := 15
	interval := 1000 / ticksPerSeconds
	max := calculateTotalTicks(currentTask.PlannedDuration, interval)

	bar := progressBar(max)

	i := 0
	for {
		select {
		case <-r.Cancel:
			completeProgressBar()
			currentTask.CompletionPercent = sql.NullFloat64{Float64: bar.State().CurrentPercent * 100, Valid: true}
			r.wg.Done()
			return
		case <-r.Pause:
			<-r.Pause
		default:
			if i == max {
				// TODO: fix notifications
				SendNotification("Complete")
				completeProgressBar()
				currentTask.Completed = 1
				currentTask.CompletionPercent = sql.NullFloat64{Float64: bar.State().CurrentPercent * 100, Valid: true}
				close(r.Finish)
				r.wg.Done()
				return
			}

			bar.Add(1)
			i++
			time.Sleep(time.Millisecond * time.Duration(interval))
		}
	}
}
