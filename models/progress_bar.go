package models

import (
	"fmt"
	"sync"
	"time"

	"github.com/briandowns/spinner"
	"github.com/eiannone/keyboard"
	"github.com/schollz/progressbar/v3"
)

func calculateTotalTicks(minutes float64, tickIntervalMs int) int {
	return int((minutes * 60 * 1000) / float64(tickIntervalMs))
}

func progressBar(max int) *progressbar.ProgressBar {
	return progressbar.NewOptions(max,
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionSetPredictTime(false),
		progressbar.OptionShowCount(),
		progressbar.OptionShowElapsedTimeOnFinish(),
		progressbar.OptionFullWidth(),
	)
}

func ProgressBarWorker(minutes float64, pauseCh, exitCh chan bool, wg *sync.WaitGroup) {
	ticksPerSeconds := 5
	interval := 1000 / ticksPerSeconds
	max := calculateTotalTicks(minutes, interval)

	bar := progressBar(max)

	i := 0
	for {
		select {
		case <-pauseCh:
			<-pauseCh
		case <-exitCh:
			wg.Done()
		default:
			if i == max {
				wg.Done() // send signal to terminate routine
			}

			bar.Add(1)
			i++
			time.Sleep(time.Millisecond * time.Duration(interval))
		}
	}
}

func CheckKeysWorker(pauseCh, exitCh chan bool, wg *sync.WaitGroup) {
	err := keyboard.Open()
	if err != nil {
		panic(err)
	}

	defer keyboard.Close()

	keysEvents, err := keyboard.GetKeys(10)
	if err != nil {
		panic(err)
	}

	paused := false
	spinner := spinner.New(spinner.CharSets[39], 100*time.Millisecond)
	spinner.Suffix = " Press any key to resume."
	for {
		select {
		case <-exitCh:
			wg.Done()
		case event := <-keysEvents:
			if event.Err != nil {
				panic(event.Err)
			}

			if event.Key == keyboard.KeyCtrlC || event.Key == keyboard.KeyEsc || event.Rune == 'q' {
				fmt.Println("\nExiting!")
				wg.Done()
				return
			} else {
				if !paused {
					spinner.Start()
				} else {
					spinner.Stop()
				}
				paused = !paused
				pauseCh <- true
			}
		}
	}
}
