package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
)

func targetFilename() string {
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	taskName := currentTask.Name
	filetype := ".mkv"
	filename := ""

	taskName = strings.ReplaceAll(taskName, " ", "_")
	filename = fmt.Sprintf("%s-%s%s", timestamp, taskName, filetype)

	return filepath.Join(cfg.FfmpegRecordingsPath, filename)
}

func FfmpegCaptureScreen(r Remote) {

	target := targetFilename()

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
			target,
		)
	case "linux":
		log.Println("Warning. Screen capture is experiemental on linux")
		res := "1920x1080"
		cmd = exec.Command(
			"ffmpeg",
			"-video_size", res,
			"-framerate", "25",
			"-f", "x11grab",
			"-i", ":0,0",
			target,
		)
	case "windows":
		log.Println("Warning. Screen capture is experiemental on windows")
		cmd = exec.Command(
			"ffmpeg",
			"-f", "dshow",
			"-i", "video=screen-capture-recorder",
			target,
		)
	default:
		log.Println("Screen capture is not supported on this platform. Continuing...")
		r.wg.Done()
		return
	}

	log.Println("Starting screen recorder.")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Start()
	if err != nil {
		log.Print(err)
		r.wg.Done()
		return
	}

	select {
	case <-r.Cancel:
		terminate(cmd)
	case <-r.Finish:
		terminate(cmd)
	}

	err = cmd.Wait()
	if err != nil {
		log.Print(err) // 255 exit code is expected.
	}

	_, err = os.Stat(target)
	if err != nil {
		log.Print(err)
		log.Print(stdout.String())
		log.Print(stderr.String())
	} else {
		log.Print("Successfully captured screen recording at:")
		log.Print(target)
		UpdateTaskVodByID(currentTask.ID, target)
	}

	r.wg.Done()
	return
}

func terminate(cmd *exec.Cmd) {
	err := cmd.Process.Signal(syscall.SIGTERM)
	if err != nil {
		log.Print(err)
	}
}
