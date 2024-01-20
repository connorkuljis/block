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
		progressbar.OptionSetPredictTime(true),
		progressbar.OptionShowCount(),
		progressbar.OptionShowElapsedTimeOnFinish(),
		progressbar.OptionFullWidth(),
		progressbar.OptionShowIts(),
	)
}

func RenderProgressBar(r Remote) {
	length := int(r.Task.PlannedDuration * 60)
	bar := progressBar(length)
	ticker := time.NewTicker(time.Second * 1)

	i := 0
	for {
		select {
		case <-r.Cancel:
			saveBarState(r.Task, bar)
			r.wg.Done()
			return
		case <-r.Pause:
			<-r.Pause
		case <-ticker.C:
			if i == length {
				saveBarState(r.Task, bar)
				SendNotification()
				close(r.Finish)
				r.wg.Done()
				return
			}

			bar.Add(1)
			i++
		}
	}
}

func saveBarState(task Task, bar *progressbar.ProgressBar) {
	fmt.Println()
	percent := bar.State().CurrentPercent * 100
	if err := UpdateCompletionPercent(task, percent); err != nil {
		log.Print(err)
	}
}
