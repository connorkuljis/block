package main

import (
	"log"
	"time"

	"github.com/briandowns/spinner"
	"github.com/eiannone/keyboard"
)

func PollInput(r Remote) {
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
		case <-r.Finish:
			r.wg.Done()
			return
		case event := <-keysEvents:
			if event.Err != nil {
				panic(event.Err)
			}

			if event.Key == keyboard.KeyCtrlC {
				if paused {
					spinner.Stop()
					close(r.Pause)
				}
				close(r.Cancel)
				r.wg.Done()
				return
			} else {
				if paused {
					spinner.Stop()
					err := r.Blocker.BlockAndReset()
					if err != nil {
						log.Print(err)
					}
				} else {
					spinner.Start()
					err := r.Blocker.UnblockAndReset()
					if err != nil {
						log.Print(err)
					}
				}
				paused = !paused
				r.Pause <- true
			}
		}
	}
}
