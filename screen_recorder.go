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
	filetype := ".mkv"

	timestamp := time.Now().Format("2006-01-02_15-04-05")

	filename := filepath.Join(recordingsDir, timestamp) + filetype

	cmd := exec.Command("ffmpeg",
		"-f", "avfoundation",
		"-i", "1:0",
		"-pix_fmt", "yuv420p",
		"-r", "25",
		filename,
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

	fmt.Println("Screen recording stopped. Saved to " + filename)
	wg.Done()
	return
}
