package interactive

import (
	"io"
	"log"
	"sync"

	"github.com/connorkuljis/block-cli/internal/tasks"
	"github.com/connorkuljis/block-cli/pkg/blocker"
)

type Remote struct {
	Task    *tasks.Task
	Blocker blocker.Blocker
	W       io.Writer

	Wg                *sync.WaitGroup
	Pause             chan bool
	Cancel            chan bool
	Finish            chan bool
	CompletionPercent chan float64
	TotalTimeSeconds  chan int
}

func RunTasks(w io.Writer, task *tasks.Task, blocker blocker.Blocker) (int, float64) {
	remote := &Remote{
		Task:              task,
		Blocker:           blocker,
		Wg:                &sync.WaitGroup{},
		W:                 w,
		Pause:             make(chan bool, 1),
		Cancel:            make(chan bool, 1),
		Finish:            make(chan bool, 1),
		CompletionPercent: make(chan float64, 1),
		TotalTimeSeconds:  make(chan int, 1),
	}

	// run the configured goroutines
	remote.Wg.Add(2)

	log.Println("Rendering progress bar")
	go RenderProgressBar(remote)

	log.Println("Polling input")
	go PollInput(remote)

	if task.ScreenEnabled == 1 {
		remote.Wg.Add(1)
		log.Println("Starting screen recorder")
		go FfmpegCaptureScreen(remote)
	}

	// wait for the goroutines to finish
	remote.Wg.Wait()

	percent := <-remote.CompletionPercent
	totalTimeSeconds := <-remote.TotalTimeSeconds

	return totalTimeSeconds, percent

}
