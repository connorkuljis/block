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

	"github.com/connorkuljis/block-cli/config"
	"github.com/connorkuljis/block-cli/tasks"
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
	InputFile   string
	OutputFile  string
	FrameRate   string

	Resolution string // linux only
}

func FfmpegCaptureScreen(remote *Remote) {
	var cmd *exec.Cmd
	var cmdArgs []string

	recording := filepath.Join(config.GetFfmpegRecordingPath(), conventionalFilename(
		time.Now().Format(TimeFormat),
		remote.Task.Name,
		".mkv",
	))

	opts := FfmpegCommandOpts{
		FrameRate:  "25",
		OutputFile: recording,
	}

	switch runtime.GOOS {
	case "darwin":
		opts.InputFormat = "avfoundation"
		opts.InputFile = config.GetAvfoundationDevice()

		cmdArgs = append(cmdArgs, "-f", opts.InputFormat)
		cmdArgs = append(cmdArgs, "-i", opts.InputFile)
		cmdArgs = append(cmdArgs, "-r", opts.FrameRate)
		cmdArgs = append(cmdArgs, opts.OutputFile)

	case "linux":
		log.Println("Warning. Screen capture is experiemental on linux")

		opts.InputFormat = "x11grab"
		opts.InputFile = ":0,0"
		opts.Resolution = "1920x1080"

		cmdArgs = append(cmdArgs, "-f", opts.InputFormat)
		cmdArgs = append(cmdArgs, "-i", opts.InputFile)
		cmdArgs = append(cmdArgs, "-framerate", opts.FrameRate)
		cmdArgs = append(cmdArgs, "-video_size", opts.Resolution)
		cmdArgs = append(cmdArgs, opts.OutputFile)

	// case "windows":
	// 	log.Println("Warning. Screen capture is experiemental on windows")

	// 	opts.InputFormat = "dshow"
	// 	opts.InputFile = "video=screen-capture-recorder"

	// 	cmdArgs = append(cmdArgs, "-f", opts.InputFormat)
	// 	cmdArgs = append(cmdArgs, "-i", opts.InputFile)
	// 	cmdArgs = append(cmdArgs, opts.OutputFile)

	default:
		log.Println("Screen capture is not supported on this platform. Continuing...")
		remote.Wg.Done()
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
		remote.Wg.Done()
		return
	}

	select {
	case <-remote.Cancel:
		terminate(cmd)
	case <-remote.Finish:
		terminate(cmd)
	}

	if err = cmd.Wait(); err != nil {
		// 255 exit code is expected to be logged on success.
		log.Print(err)
	}

	_, err = os.Stat(recording)
	if err != nil {
		log.Print(stdout.String())
		log.Print(stderr.String())
		log.Print(err)
		remote.Wg.Done()
		return
	}

	log.Print("Successfully captured screen recording at: " + recording)

	if err = tasks.UpdateScreenURL(*remote.Task, recording); err != nil {
		log.Print(err)
	}

	remote.Wg.Done()
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
