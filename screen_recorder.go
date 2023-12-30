package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"text/tabwriter"
	"time"
)

func FfmpegCaptureScreen(minutes float64, w *tabwriter.Writer, cancelCh, finishCh chan bool, wg *sync.WaitGroup) {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		log.Println("Error executing ffmpeg: " + err.Error())
		wg.Done()
		return
	}

	downloadsDir := "/Downloads"

	userHomeDir = filepath.Join(userHomeDir, downloadsDir)

	filename := ""
	filetype := ".mkv"
	timestamp := time.Now().Format("2006-01-02_15-04-05")

	if taskName == "" {
		filename = timestamp + filetype
	} else {
		filename = timestamp + "_" + taskName + filetype
	}

	filename = strings.ReplaceAll(filename, " ", "-")

	filepath := filepath.Join(userHomeDir, filename)

	cmd := exec.Command("ffmpeg",
		"-f", "avfoundation",
		"-i", "1:0",
		"-pix_fmt", "yuv420p",
		"-r", "25",
		filepath,
	)

	err = cmd.Start()
	if err != nil {
		log.Println("Error executing ffmpeg: " + err.Error())
		wg.Done()
		return
	}

	select {
	case <-cancelCh:
		cmd.Process.Signal(syscall.SIGTERM)
	case <-finishCh:
		cmd.Process.Signal(syscall.SIGTERM)
	}

	fmt.Fprintln(w, "Saved recording to: "+filepath)

	wg.Done()
	return
}
