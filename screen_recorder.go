package main

import (
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

func FfmpegCaptureScreen(r Remote) {
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	taskName := currentTask.Name
	filetype := ".mkv"
	filename := ""

	if taskName == "" {
		filename = fmt.Sprintf("%s%s", timestamp, filetype)
	} else {
		taskName = strings.ReplaceAll(taskName, " ", "_")
		filename = fmt.Sprintf("%s-%s%s", timestamp, taskName, filetype)
	}

	filepath := filepath.Join(cfg.FfmpegRecordingsPath, filename)

	fmt.Println("Saving recording to: " + filepath)

	inputs := cfg.AvfoundationDevice

	cmd := exec.Command("ffmpeg",
		"-f", "avfoundation",
		"-i", inputs,
		"-pix_fmt", "yuv420p",
		"-r", "25",
		filepath,
	)

	err := cmd.Start()
	if err != nil {
		log.Println("Error executing ffmpeg: " + err.Error())
		r.wg.Done()
		return
	}

	select {
	case <-r.Cancel:
		cmd.Process.Signal(syscall.SIGTERM)
	case <-r.Finish:
		cmd.Process.Signal(syscall.SIGTERM)
	}

	UpdateTaskVodByID(currentTask.ID, filepath)

	r.wg.Done()
	return
}
