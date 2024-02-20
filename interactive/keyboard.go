package interactive

import (
	"fmt"
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
			r.Wg.Done()
			return
		case event := <-keysEvents:
			if event.Err != nil {
				panic(event.Err)
			}

			if event.Key == keyboard.KeyCtrlC || event.Key == keyboard.KeyEsc {
				if paused {
					spinner.Stop()
					close(r.Pause)
				}
				close(r.Cancel)
				r.Wg.Done()
				return
			} else { // any other key press

				paused = !paused

				if paused {
					err := r.Blocker.Start()
					if err != nil {
						log.Print(err)
					}
					fmt.Println("paused, unblocking sites")
					r.Pause <- true
					spinner.Start()
				} else {
					spinner.Stop()
					err := r.Blocker.Stop()
					if err != nil {
						log.Print(err)
					}

					r.Pause <- true
				}
			}
		}
	}
}
