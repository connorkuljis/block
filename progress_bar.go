package main

import (
	"fmt"
	"log"
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
			recordBarState(bar)
			r.wg.Done()
			return
		case <-r.Pause:
			<-r.Pause
		default:
			if i == max {
				recordBarState(bar)
				SendNotification()
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

func recordBarState(bar *progressbar.ProgressBar) {
	fmt.Println()

	task := currentTask
	percent := bar.State().CurrentPercent * 100

	if err := UpdateCompletionPercent(task, percent); err != nil {
		log.Print(err)
	}
}
