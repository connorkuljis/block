package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"time"

	"github.com/connorkuljis/task-tracker-cli/models"
	"github.com/manifoldco/promptui"
)

type Options struct {
	minutes float64
}

func main() {
	greeting()

	options := parseFlags()

	minutes := options.minutes
	if options.minutes == 0 {
		minutes = askMinutes()
	}

	startTimer(minutes)

}

func parseFlags() Options {
	options := Options{}

	var minutes float64

	flag.Float64Var(&minutes, "minutes", 0, "Number of minutes")

	flag.Parse()

	options.minutes = float64(minutes)

	return options
}

func greeting() {
	fmt.Println("# Welcome to task-tracker-cli #")
}

func askMinutes() float64 {
	validate := func(input string) error {
		_, err := strconv.ParseFloat(input, 64)

		// Check for conversion error
		if err != nil {
			return errors.New("Invalid number")
		}
		return nil
	}

	prompt := promptui.Prompt{
		Label:    "Enter task time (minutes)",
		Validate: validate,
	}

	result, err := prompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("You choose %q\n", result)

	n, err := strconv.ParseFloat(result, 32)
	if err != nil {
		log.Fatalln("Could not convert string to float: " + result)
	}

	return n
}

func startTimer(minutes float64) {
	go say(fmt.Sprintf("Starting a timer for %.0f minutes.", minutes))

	blocker := models.NewBlocker()

	n, err := blocker.Block()
	if err != nil {
		log.Println(err)
	}

	fmt.Printf(">> Distractions blocked. (%d bytes updated)\n", n)

	startTime := time.Now()
	fmt.Println("Start: " + startTime.Format("3:04:05pm"))

	var wg sync.WaitGroup // synchronise progress bar and key checker
	wg.Add(1)             // if any routine is done, both are done

	pauseCh := make(chan bool, 1)
	exitCh := make(chan bool, 1)

	go models.ProgressBarWorker(minutes, pauseCh, exitCh, &wg)
	go models.CheckKeysWorker(pauseCh, exitCh, &wg)

	wg.Wait()

	endTime := time.Now()
	duration := time.Now().Sub(startTime)

	fmt.Printf("End: " + endTime.Format("3:04:05pm"))
	fmt.Printf(" [duration: %d hours, %d minutes, %d seconds.]\n", int(duration.Hours()), int(duration.Minutes())%60, int(duration.Seconds())%60)

	n, err = blocker.Unblock()
	if err != nil {
		log.Println(err)
	}

	fmt.Printf(">> Blocker disabled. (%d bytes updated)\n", n)

	say("Goodbye")
}

func say(msg string) {
	cmd := exec.Command("say", "-v", "Bubbles", msg)
	err := cmd.Run()
	if err != nil {
		fmt.Println(err.Error())
	}
}
