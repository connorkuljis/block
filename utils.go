package main

import (
	"log"

	"github.com/gen2brain/beeep"
)

func SendNotification() {
	if flags.Verbose {
		log.Println("Sending notification...")
	}

	var icon = ""

	err := beeep.Beep(beeep.DefaultFreq, beeep.DefaultDuration)
	if err != nil {
		log.Printf("Error, could not send notification beep: %v", err)
	}

	err = beeep.Notify("block-cli", "Your session has finished!", icon)
	if err != nil {
		log.Printf("Error, could not send notification alert: %v", err)
	}

	if flags.Verbose {
		log.Println("Notification sent!")
	}
}

func boolToInt(cond bool) int {
	var v int
	if cond {
		v = 1
	}
	return v
}
