package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
	"unicode"
)

// generates an path combining the current timestamp, taskname and filetype
func generateOutFilename(formatStr string, name string, filetype string) string {
	const separator = "_"

	timestamp := time.Now().Format(formatStr)

	if name == "" {
		return timestamp + filetype
	}

	// uppercase first letter in each word of task name joined by empty space
	// eg: "software demo" -> 'SoftwareDemo'
	label := ""
	parts := strings.Split(name, " ")
	for i := range parts {
		runes := []rune(parts[i])
		first := runes[0]
		first = unicode.ToUpper(first)
		label += string(runes)
	}

	return timestamp + separator + label + filetype
}

func FfmpegCaptureScreen(r Remote) {
	var cmd *exec.Cmd

	filename := generateOutFilename("2006-01-02_15-04", r.Task.Name, ".mkv")
	filename = filepath.Join(cfg.FfmpegRecordingsPath, filename)

	switch runtime.GOOS {
	case "darwin":
		inputs := cfg.AvfoundationDevice
		cmd = exec.Command(
			"ffmpeg",
			"-f", "avfoundation",
			"-i", inputs,
			"-pix_fmt", "yuv420p",
			"-r", "25",
			filename,
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
			filename,
		)
	case "windows":
		log.Println("Warning. Screen capture is experiemental on windows")
		cmd = exec.Command(
			"ffmpeg",
			"-f", "dshow",
			"-i", "video=screen-capture-recorder",
			filename,
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

	_, err = os.Stat(filename)
	if err != nil {
		log.Print(err)
		log.Print(stdout.String())
		log.Print(stderr.String())
		r.wg.Done()
		return
	}

	log.Print("Successfully captured screen recording at: " + filename)

	if err = UpdateScreenURL(r.Task, filename); err != nil {
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

func FfmpegGenerateTimelapse(files []string) (string, error) {
	var args []string
	filename := ""

	if len(files) == 0 {
		return filename, errors.New("Need at least one file to generate timelapse.")
	}

	filename = generateOutFilename("2006-01-02", "", ".mkv")
	filename = filepath.Join(cfg.FfmpegRecordingsPath, filename)

	listFilename := generateOutFilename("2006-01-02", "list", ".txt")
	listFilename = filepath.Join(cfg.FfmpegRecordingsPath, listFilename)

	listFile, err := os.Create(listFilename)
	if err != nil {
		return filename, err
	}
	defer listFile.Close()

	for _, file := range files {
		input := filepath.Join(cfg.FfmpegRecordingsPath, file)
		listFile.WriteString(fmt.Sprintf("file '%s'\n", input))
	}

	args = append(args, "-f", "concat", "-safe", "0", "-i", listFilename, "-c", "copy", "-y", filename)

	cmd := exec.Command("ffmpeg", args...)
	log.Println("Generating timelapse: " + cmd.String())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return "", err
	}

	err = os.Remove(listFilename)
	if err != nil {
		return "", err
	}

	return filename, nil
}
