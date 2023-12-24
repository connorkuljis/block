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

	pauseCh := make(chan bool)
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
	n := 100
	bar := progressbar.Default(int64(n))

	for i := 0; i < n; i++ {
		select {
		case <-pauseCh:
			<-pauseCh // sleeps until signal
		case <-stopCh:
			wg.Done()
			return
		default:
			bar.Add(1)
			time.Sleep(40 * time.Millisecond)
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
			fmt.Println("**Exiting keys thread.")
			return
		case event := <-keysEvents:
			if event.Err != nil {
				panic(event.Err)
			}
			fmt.Printf("You pressed: rune %q, key %X\r\n", event.Rune, event.Key)
			if event.Key == keyboard.KeyEsc {
				stopCh <- true
			} else {
				pauseCh <- true
			}
		}
	}
}
