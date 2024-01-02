package main

import (
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

			if event.Key == keyboard.KeyCtrlC || event.Key == keyboard.KeyEsc || event.Rune == 'q' {
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
				} else {
					spinner.Start()
				}
				paused = !paused
				r.Pause <- true
			}
		}
	}
}
