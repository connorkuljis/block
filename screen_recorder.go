package main

import (
	"bytes"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
)

// generates an path combining the current timestamp, taskname and filetype
func generateScreenRecordingName() string {
	var parts []string
	parts = append(parts, time.Now().Format("2006-01-02_15-04-05"))
	parts = append(parts, strings.ReplaceAll(currentTask.Name, " ", "_"))
	parts = append(parts, ".mkv")

	filename := strings.Join(parts, "-")

	return filename
}

func FfmpegCaptureScreen(r Remote) {
	filename := generateScreenRecordingName()
	target := filepath.Join(cfg.FfmpegRecordingsPath, filename)

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

	if err = cmd.Wait(); err != nil {
		log.Print(err) // 255 exit code is expected to be logged on success.
	}

	_, err = os.Stat(target)
	if err != nil {
		log.Print(err)
		log.Print(stdout.String())
		log.Print(stderr.String())
		r.wg.Done()
		return
	}

	log.Print("Successfully captured screen recording at:")
	log.Print(target)

	if err = UpdateScreenURL(currentTask, filename); err != nil {
		log.Print(err)
	}

	r.wg.Done()
}

func terminate(cmd *exec.Cmd) {
	err := cmd.Process.Signal(syscall.SIGTERM)
	if err != nil {
		log.Print(err)
	}
}
