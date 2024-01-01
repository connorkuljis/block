package main

import (
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"text/tabwriter"
	"time"
)

func FfmpegCaptureScreen(minutes float64, w *tabwriter.Writer, cancelCh, finishCh chan bool, wg *sync.WaitGroup) {
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filetype := ".mkv"

	filename := ""
	if taskName == "" {
		filename = fmt.Sprintf("%s%s", timestamp, filetype)
	} else {
		filename = fmt.Sprintf("%s-%s%s", timestamp, taskName, filetype)
	}

	filepath := filepath.Join(config.FfmpegRecordingsPath, filename)

	cmd := exec.Command("ffmpeg",
		"-f", "avfoundation",
		"-i", "1:0",
		"-pix_fmt", "yuv420p",
		"-r", "25",
		filepath,
	)

	err := cmd.Start()
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
