package utils

import (
	"fmt"
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

func SecsToHHMMSS(secs int64) string {
	hours := secs / 3600
	minutes := (secs % 3600) / 60
	seconds := secs % 60

	var formattedString string
	if hours > 0 {
		formattedString = fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
	} else {
		formattedString = fmt.Sprintf("%02d:%02d", minutes, seconds)
	}

	return formattedString
}
