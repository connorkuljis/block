package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/eiannone/keyboard"
	"github.com/manifoldco/promptui"
	"github.com/schollz/progressbar/v3"
)

func greeting() {
	fmt.Println(`# Welcome to task-tracker-cli
To exit press ^C.`)
}

func promptTaskTime() int {

	validate := func(input string) error {
		_, err := strconv.Atoi(input)

		// Check for conversion error
		if err != nil {
			return errors.New("Invalid number")
		}
		return nil
	}

	prompt := promptui.Prompt{
		Label:    "Enter task time (milliseconds)",
		Validate: validate,
	}

	result, err := prompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("You choose %q\n", result)

	n, err := strconv.Atoi(result)
	if err != nil {
		log.Fatalln("Could not convert string to int : " + result)
	}

	return n
}

func main() {
	greeting()

	err := keyboard.Open()
	if err != nil {
		panic(err)
	}

	n := promptTaskTime()

	pauseCh := make(chan bool, 1)

	// synchronise progress bar and exit key
	var wg sync.WaitGroup
	wg.Add(1)

	go progressBarWorker(n, pauseCh, &wg)
	go checkKeys(pauseCh, &wg)

	wg.Wait()

	keyboard.Close()
}

func progressBarWorker(n int, pauseCh chan bool, wg *sync.WaitGroup) {
	bar := progressbar.Default(int64(n))

	for i := 0; i < n; i++ {
		select {
		case <-pauseCh:
			<-pauseCh
		default:
			bar.Add(1)
			time.Sleep(1 * time.Millisecond)
		}
	}
	wg.Done()
}

func checkKeys(pauseCh chan bool, wg *sync.WaitGroup) {
	keysEvents, err := keyboard.GetKeys(10)
	if err != nil {
		panic(err)
	}

	for {
		select {
		case event := <-keysEvents:
			if event.Err != nil {
				panic(event.Err)
			}

			if event.Key == keyboard.KeyCtrlC || event.Key == keyboard.KeyEsc || event.Rune == 'q' {
				fmt.Println("\nExiting!")
				wg.Done()
				return
			} else {
				pauseCh <- true
			}
		}
	}
}
