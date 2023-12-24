package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/eiannone/keyboard"
	"github.com/schollz/progressbar/v3"
)

func worker(pauseCh, stopCh chan bool, wg *sync.WaitGroup) {
	n := 100

	bar := progressbar.Default(int64(n))

	for i := 0; i < n; i++ {
		select {
		case <-pauseCh:
			// The goroutine will be "suspended" until a resume signal is received
			fmt.Println("Paused. Waiting to be resumed.")
			<-pauseCh

		case <-stopCh:
			fmt.Println("Exiting.")
			wg.Done()
			return

		default:
			bar.Add(1)
			time.Sleep(40 * time.Millisecond)
		}

	}
	fmt.Println("worker done.")
	wg.Done()
}

func printKeyPressed(pauseCh, stopCh chan bool) {
	for {
		select {

		case <-stopCh:
			fmt.Printf("stopped looking for key.")
			return

		default:
			char, key, err := keyboard.GetKey()
			if err != nil {
				fmt.Println("error: " + err.Error())
				panic(err)
			}

			if key == keyboard.KeyCtrlC {
				fmt.Println("Ctrl+C pressed. Exiting.")
				return
			}

			fmt.Printf("Key pressed: %c\n", char)

			pauseCh <- true
		}
	}
}

func main() {
	err := keyboard.Open()
	if err != nil {
		panic(err)
	}

	pauseCh := make(chan bool)
	stopCh := make(chan bool)

	var wg sync.WaitGroup
	wg.Add(1)

	go worker(pauseCh, stopCh, &wg)
	go printKeyPressed(pauseCh, stopCh)

	wg.Wait()

	// keyboard.Close()
	fmt.Println("Goodbye.")
}

func progressBarDemo() {
	bar := progressbar.Default(100)
	for i := 0; i < 100; i++ {
		bar.Add(1)
		time.Sleep(40 * time.Millisecond)
	}
}

func spinnerDemo() {
}

func loop() {
	max, iter := 255, 2
	for i := 0; i < iter; i++ {
		iterateNums(max)
	}
}

func iterateNums(n int) {
	for i := 0; i < n; i++ {
		fmt.Printf("%d\t", i)
	}
}
