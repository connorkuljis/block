package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/briandowns/spinner"
	"github.com/eiannone/keyboard"
)

func PollInput(pauseCh, cancelCh, finishCh chan bool, wg *sync.WaitGroup) {
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
		case <-finishCh:
			wg.Done()
			return
		case event := <-keysEvents:
			if event.Err != nil {
				panic(event.Err)
			}

			if event.Key == keyboard.KeyCtrlC || event.Key == keyboard.KeyEsc || event.Rune == 'q' {
				fmt.Println("\nExiting!")
				if paused {
					spinner.Disable()
					close(pauseCh)
					close(cancelCh)
				} else {
					close(cancelCh)
				}
				wg.Done()
				return
			} else {
				paused = !paused
				if paused {
					spinner.Start()
				} else {
					spinner.Stop()
				}
				pauseCh <- true
			}
		}
	}
}
