package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/eiannone/keyboard"
	"github.com/schollz/progressbar/v3"
)

func main() {
	err := keyboard.Open()
	if err != nil {
		panic(err)
	}

	fmt.Println("Press any key to pause/resume the progress bar...")

	pauseCh := make(chan bool, 1)
	stopCh := make(chan bool)

	// synchronise progress bar
	var wg sync.WaitGroup
	wg.Add(1)

	go progressBarWorker(pauseCh, stopCh, &wg)
	go checkKeys(pauseCh, stopCh)

	// wait for progress bar to complete
	wg.Wait()

	stopCh <- true // signal keychecker to exit
	keyboard.Close()
}

func progressBarWorker(pauseCh, stopCh chan bool, wg *sync.WaitGroup) {
	n := 5000
	bar := progressbar.Default(int64(n))

	for i := 0; i < n; i++ {
		select {
		case <-pauseCh:
			<-pauseCh
		case <-stopCh:
			wg.Done()
			return
		default:
			bar.Add(1)
			time.Sleep(1 * time.Millisecond)
		}
	}
	wg.Done()
}

func checkKeys(pauseCh, stopCh chan bool) {
	keysEvents, err := keyboard.GetKeys(10)
	if err != nil {
		panic(err)
	}

	for {
		select {
		case <-stopCh:
			return
		case event := <-keysEvents:
			if event.Err != nil {
				panic(event.Err)
			}
			fmt.Printf("You pressed: rune %q, key %X\r\n", event.Rune, event.Key)
			if event.Key == keyboard.KeyEsc {
				fmt.Println("Exiting loop!")
				stopCh <- true
			} else {
				pauseCh <- true
			}
		}
	}
}
