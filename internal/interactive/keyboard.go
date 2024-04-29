package interactive

import (
	"log"
	"log/slog"
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
	spinner := spinner.New(spinner.CharSets[40], 100*time.Millisecond)
	spinner.Prefix = "Press any key to resume:"
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
				slog.Info("Cancelling.")
				close(remote.Cancel)
				remote.Wg.Done()
				return
			} else if event.Key == keyboard.KeySpace {
				paused = !paused
				if paused {
					unpause(remote, spinner)
				} else {
					pause(remote, spinner)
				}
			}
		}
	}
}

func unpause(remote *Remote, spinner *spinner.Spinner) {
	_, err := remote.Blocker.Stop()
	if err != nil {
		log.Print(err)
	}
	remote.Pause <- true
	spinner.Start()
}

func pause(remote *Remote, spinner *spinner.Spinner) {
	spinner.Stop()
	_, err := remote.Blocker.Start()
	if err != nil {
		log.Print(err)
	}
	remote.Pause <- true
}
