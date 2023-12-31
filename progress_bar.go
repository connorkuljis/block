package main

import (
	"fmt"
	"sync"
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

func RenderProgressBar(minutes float64, pauseCh, cancelCh, finishCh chan bool, wg *sync.WaitGroup) {
	calculateTotalTicks := func(minutes float64, tickIntervalMs int) int {
		return int((minutes * 60 * 1000) / float64(tickIntervalMs))
	}

	ticksPerSeconds := 15
	interval := 1000 / ticksPerSeconds
	max := calculateTotalTicks(minutes, interval)

	bar := progressBar(max)

	i := 0
	for {
		select {
		case <-cancelCh:
			completeProgressBar()
			wg.Done()
			return
		case <-pauseCh:
			<-pauseCh
		default:
			if i == max {
				// TODO: fix notifications
				// SendNotification("Complete")
				completeProgressBar()
				close(finishCh)
				wg.Done()
				return
			}

			bar.Add(1)
			i++
			time.Sleep(time.Millisecond * time.Duration(interval))
		}
	}
}
