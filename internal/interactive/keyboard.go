package interactive

import (
	"fmt"
	"log"
	"time"

	"github.com/briandowns/spinner"
	"github.com/eiannone/keyboard"
)

func PollInput(remote *Remote) {
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
		case <-remote.Finish:
			remote.Wg.Done()
			return
		case event := <-keysEvents:
			if event.Err != nil {
				panic(event.Err)
			}

			if event.Key == keyboard.KeyCtrlC || event.Key == keyboard.KeyEsc {
				if paused {
					spinner.Stop()
					close(remote.Pause)
				}
				close(remote.Cancel)
				remote.Wg.Done()
				return
			} else { // any other key press

				paused = !paused

				if paused {
					err := remote.Blocker.Stop()
					if err != nil {
						log.Print(err)
					}
					fmt.Println("paused, unblocking sites")
					remote.Pause <- true
					spinner.Start()
				} else {
					spinner.Stop()
					err := remote.Blocker.Start()
					if err != nil {
						log.Print(err)
					}

					remote.Pause <- true
				}
			}
		}
	}
}
