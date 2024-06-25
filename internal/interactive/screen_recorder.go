package interactive

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/connorkuljis/block-cli/internal/config"
	"github.com/connorkuljis/block-cli/internal/ffmpeg"

	"github.com/fatih/color"
)

const TimeFormat = "2006-01-02_15-04"

func conventionalFilename(timestamp, name, filetype string) string {
	seperator := "_"
	concatenator := "-"
	return timestamp + seperator + strings.ReplaceAll(name, " ", concatenator) + filetype
}

func FfmpegCaptureScreen(remote *Remote) {
	var filename string

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

	ffmpeg.RecordScreen(config.GetAvfoundationDevice(), outputFile, remote.Cancel, remote.Finish, remote.Wg)

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
