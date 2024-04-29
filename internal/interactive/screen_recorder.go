package interactive

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

	"github.com/connorkuljis/block-cli/internal/config"
	"github.com/connorkuljis/block-cli/internal/tasks"

	"github.com/fatih/color"
)

const TimeFormat = "2006-01-02_15-04"

func conventionalFilename(timestamp, name, filetype string) string {
	seperator := "_"
	concatenator := "-"
	return timestamp + seperator + strings.ReplaceAll(name, " ", concatenator) + filetype
}

type FfmpegCommandOpts struct {
	InputFormat string

	InputDevice string

	FrameRate   string
	PixelFormat string
	Demuxer     string
	VideoCodec  string

	OutputFile string

	Resolution string // linux only
}

func FfmpegCaptureScreen(remote *Remote) {
	var filename string
	var cmd *exec.Cmd
	var cmdArgs []string

	timestamp := remote.Task.CreatedAt.Format(TimeFormat)
	name := remote.Task.TaskName
	if name == "" {
		filename = fmt.Sprintf("%s.mkv", timestamp)
	} else {
		name = strings.ReplaceAll(name, " ", "-")
		filename = fmt.Sprintf("%s-%s.mkv", timestamp, name)
	}

	recordingPath := config.GetFfmpegRecordingPath()
	outputFile := filepath.Join(recordingPath, filename)

	switch runtime.GOOS {
	case "darwin":
		opts := FfmpegCommandOpts{
			InputFormat: "avfoundation",
			InputDevice: config.GetAvfoundationDevice(),
			FrameRate:   "30",
			// Demuxer:     "avfoundation",
			PixelFormat: "yuv420p",
			VideoCodec:  "libx264",
			OutputFile:  outputFile,
		}
		cmdArgs = append(cmdArgs, "-f", opts.InputFormat)
		cmdArgs = append(cmdArgs, "-i", opts.InputDevice)
		cmdArgs = append(cmdArgs, "-r", opts.FrameRate)
		cmdArgs = append(cmdArgs, "-pix_fmt", opts.PixelFormat)
		cmdArgs = append(cmdArgs, "-c:v", opts.VideoCodec)
		cmdArgs = append(cmdArgs, opts.OutputFile)
	case "linux":
		log.Println("Warning. Screen capture is experiemental on linux")
		opts := FfmpegCommandOpts{
			InputFormat: "x11grab",
			InputDevice: ":0,0",
			Resolution:  "1920x1080",
		}
		cmdArgs = append(cmdArgs, "-f", opts.InputFormat)
		cmdArgs = append(cmdArgs, "-i", opts.InputDevice)
		cmdArgs = append(cmdArgs, "-framerate", opts.FrameRate)
		cmdArgs = append(cmdArgs, "-video_size", opts.Resolution)
		cmdArgs = append(cmdArgs, opts.OutputFile)
	default:
		log.Println("Screen capture is not supported on this platform. Continuing...")
		remote.Wg.Done()
		return
	}

	cmd = exec.Command("ffmpeg", cmdArgs...)
	log.Println(cmd.String())

	// TODO: check input device is valid before forking child process.
	if err := cmd.Start(); err != nil {
		log.Print(err)
		close(remote.Cancel)
		remote.Wg.Done()
		return
	}

	select {
	case <-remote.Cancel:
		terminate(cmd)
	case <-remote.Finish:
		terminate(cmd)
	}

	if err := cmd.Wait(); err != nil {
		log.Print(err)
	} else {
		log.Print("Successfully captured screen recording at: " + outputFile)
		if err = tasks.UpdateScreenURL(remote.Db, *remote.Task, outputFile); err != nil {
			log.Print(err)
		}
	}

	remote.Wg.Done()
	return
}

func terminate(cmd *exec.Cmd) {
	err := cmd.Process.Signal(syscall.SIGTERM)
	if err != nil {
		log.Print(err)
	}
}

func FfmpegConcatenateScreenRecordings(inTime time.Time, files []string) (string, error) {
	var args []string

	if len(files) == 0 {
		return "", errors.New("Need at least one file to generate timelapse.")
	}

	filename := filepath.Join(config.GetFfmpegRecordingPath(), conventionalFilename(
		inTime.Format(TimeFormat),
		"concatenated",
		".mkv",
	))

	// $ cat mylist.txt <-- we use a temporary file
	// file '/path/to/file1'
	// file '/path/to/file2'
	// file '/path/to/file3'

	// $ ffmpeg -f concat -safe 0 -i mylist.txt -c copy output.mp4
	temp, err := os.CreateTemp("", "listfile")
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
