package main

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

func FfmpegCaptureScreen(minutes float64, cancelCh, finishCh chan bool, wg *sync.WaitGroup) {
	fmt.Println("Screen recording started")

	recordingsDir := "/Users/connor/Code/golang/task-tracker-cli/recordings" // TODO: source this from config file.

	filename := ""
	filetype := ".mkv"
	timestamp := time.Now().Format("2006-01-02_15-04-05")

	// prepend task name to file
	if taskName == "" {
		filename = timestamp + filetype
	} else {
		filename = timestamp + "_" + taskName + filetype
	}

	filepath := filepath.Join(recordingsDir, filename)

	cmd := exec.Command("ffmpeg",
		"-f", "avfoundation",
		"-i", "1:0",
		"-pix_fmt", "yuv420p",
		"-r", "25",
		filepath,
	)

	err := cmd.Start()
	if err != nil {
		panic(err)
	}

	select {
	case <-cancelCh:
		cmd.Process.Signal(syscall.SIGTERM)
	case <-finishCh:
		cmd.Process.Signal(syscall.SIGTERM)
	}

	fmt.Println("Screen recording stopped. Saved to " + filepath)
	wg.Done()
	return
}
