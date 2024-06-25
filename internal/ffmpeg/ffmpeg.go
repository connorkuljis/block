package ffmpeg

import (
	"bufio"
	"log"
	"os"
	"os/exec"
	"sync"
)

func RecordScreen(inputDevice string, outputPath string, cancel chan bool, finish chan error, wg *sync.WaitGroup) error {
	inputFormat := "avfoundation" // input format.
	frameRate := "25"             // frame rate.
	codec := "libx264"            // codec.
	rescale := "scale=-1:1080"    // keep scale to 1080p.
	overwrite := "-y"             // allows overwriting existing file.

	// not sure if pixel format is needed.
	// pixelFormat := "yuv420p"

	cmd := exec.Command("ffmpeg",
		"-f", inputFormat,
		"-r", frameRate, // !must be before input device.
		"-i", inputDevice,
		"-c:v", codec,
		"-vf", rescale,
		overwrite,
		outputPath,
	)

	log.Println("Starting ffmpeg:", cmd.String())

	// Create a pipe to capture stderr.
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return err
	}
	go func() {
		finish <- cmd.Wait()
	}()

	// Start a goroutine to read and log stderr
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			log.Println("FFmpeg:", scanner.Text())
		}
	}()

	// Wait for either the stop signal or the process to finish
	select {
	case <-cancel:
		log.Println("Received stop signal, terminating FFmpeg")
		if err := cmd.Process.Signal(os.Interrupt); err != nil {
			log.Println("Failed to send interrupt signal:", err)
			cmd.Process.Kill()
		}
		wg.Done()
		return <-finish
	case err := <-finish:
		wg.Done()
		return err
	}
}
