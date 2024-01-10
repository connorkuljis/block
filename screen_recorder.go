package main

import (
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
)

func FfmpegCaptureScreen(r Remote) {
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	taskName := currentTask.Name
	filetype := ".mkv"
	filename := ""

	taskName = strings.ReplaceAll(taskName, " ", "_")
	filename = fmt.Sprintf("%s-%s%s", timestamp, taskName, filetype)

	filepath := filepath.Join(cfg.FfmpegRecordingsPath, filename)

	fmt.Println("Saving recording to: " + filepath)

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		inputs := cfg.AvfoundationDevice
		cmd = exec.Command(
			"ffmpeg",
			"-f", "avfoundation",
			"-i", inputs,
			"-pix_fmt", "yuv420p",
			"-r", "25",
			filepath,
		)
	case "linux":
		res := "1920x1080"
		cmd = exec.Command(
			"ffmpeg",
			"-video_size", res,
			"-framerate", "25",
			"-f", "x11grab",
			"-i", ":0,0",
			filepath,
		)
	case "windows":
		cmd = exec.Command(
			"ffmpeg",
			"-f", "dshow",
			"-i", "video=screen-capture-recorder",
			filepath,
		)
	default:
		log.Println("Screen capture is not supported on this platform. Continuing...")
		r.wg.Done()
		return
	}

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
