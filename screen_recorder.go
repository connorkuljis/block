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

	"github.com/fatih/color"
)

const TimeFormat = "2006-01-02_15-04"

func conventionalFilename(timestamp, name, filetype string) string {
	separator := "_"
	concatenator := "-"
	return timestamp + separator + strings.ReplaceAll(name, " ", concatenator) + filetype
}

type FfmpegCommandOpts struct {
	Format     string
	Input      string
	FrameRate  string
	Resolution string
}

func FfmpegCaptureScreen(r Remote) {
	var cmd *exec.Cmd
	var cmdArgs []string
	opts := FfmpegCommandOpts{
		FrameRate: "25",
	}

	filename := filepath.Join(cfg.FfmpegRecordingsPath, conventionalFilename(
		time.Now().Format(TimeFormat),
		r.Task.Name,
		".mkv",
	))

	switch runtime.GOOS {
	case "darwin":
		opts.Format = "avfoundation"
		opts.Input = cfg.AvfoundationDevice

		cmdArgs = append(cmdArgs, "-f", opts.Format)
		cmdArgs = append(cmdArgs, "-i", opts.Input)
		cmdArgs = append(cmdArgs, "-r", opts.FrameRate)

	case "linux":
		log.Println("Warning. Screen capture is experiemental on linux")

		opts.Format = "x11grab"
		opts.Input = ":0,0"
		opts.Resolution = "1920x1080"

		cmdArgs = append(cmdArgs, "-f", opts.Format)
		cmdArgs = append(cmdArgs, "-i", opts.Input)
		cmdArgs = append(cmdArgs, "-framerate", opts.FrameRate)
		cmdArgs = append(cmdArgs, "-video_size", opts.Resolution)

	case "windows":
		log.Println("Warning. Screen capture is experiemental on windows")

		opts.Format = "dshow"
		opts.Input = "video=screen-capture-recorder"

		cmdArgs = append(cmdArgs, "-f", opts.Format)
		cmdArgs = append(cmdArgs, "-i", opts.Input)

	default:
		log.Println("Screen capture is not supported on this platform. Continuing...")
		r.wg.Done()
		return
	}

	log.Println("Starting screen recorder.")

	cmd = exec.Command("ffmpeg", cmdArgs...)
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
		// 255 exit code is expected to be logged on success.
		log.Print(err)
	}

	_, err = os.Stat(filename)
	if err != nil {
		log.Print(stdout.String())
		log.Print(stderr.String())
		log.Print(err)
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

func FfmpegConcatenateScreenRecordings(inTime time.Time, files []string) (string, error) {
	var args []string
	filename := ""

	if len(files) == 0 {
		return filename, errors.New("Need at least one file to generate timelapse.")
	}

	timestamp := inTime.Format(TimeFormat)
	filename = conventionalFilename(timestamp, "concatenated", ".mkv")
	filename = filepath.Join(cfg.FfmpegRecordingsPath, filename)

	temp, err := os.CreateTemp("", timestamp+"test")
	if err != nil {
		log.Println(err)
		return filename, err
	}
	defer os.Remove(temp.Name())

	for _, file := range files {
		temp.WriteString(fmt.Sprintf("file '%s'\n", file))
	}

	args = append(args,
		"-f", "concat",
		"-safe", "0",
		"-i", temp.Name(),
		"-c", "copy",
		"-y",
		filename,
	)

	cmd := exec.Command("ffmpeg", args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	color.Green("Generating timelapse: " + cmd.String())
	err = cmd.Run()
	if err != nil {
		log.Println(stderr.String())
		log.Println(stdout.String())
		return filename, err
	}

	return filename, nil
}
