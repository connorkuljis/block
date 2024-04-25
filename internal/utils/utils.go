package utils

import (
	"log"

	"github.com/gen2brain/beeep"
)

func SendNotification() {
	var icon = ""

	err := beeep.Beep(beeep.DefaultFreq, beeep.DefaultDuration)
	if err != nil {
		log.Printf("Error, could not send notification beep: %v", err)
	}

	err = beeep.Notify("block-cli", "Your session has finished!", icon)
	if err != nil {
		log.Printf("Error, could not send notification alert: %v", err)
	}
}

func BoolToInt(cond bool) int {
	var v int
	if cond {
		v = 1
	}
	return v
}
