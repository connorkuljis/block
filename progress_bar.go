package main

import (
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
			cancel(bar)
			r.wg.Done()
			return
		case <-r.Pause:
			<-r.Pause
		default:
			if i == max {
				finish(bar)
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

func cancel(bar *progressbar.ProgressBar) {
	fmt.Println()
	currentTask.CompletionPercent.Float64 = bar.State().CurrentPercent
	currentTask.CompletionPercent.Valid = true
}

func finish(bar *progressbar.ProgressBar) {
	fmt.Println()
	currentTask.CompletionPercent.Float64 = bar.State().CurrentPercent
	currentTask.CompletionPercent.Valid = true
	SendNotification()
}
