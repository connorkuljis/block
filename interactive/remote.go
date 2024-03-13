package interactive

import (
	"log"
	"sync"

	"github.com/connorkuljis/block-cli/blocker"
	"github.com/connorkuljis/block-cli/tasks"
)

type Remote struct {
	Task    tasks.Task
	Blocker blocker.Blocker

	Wg     *sync.WaitGroup
	Pause  chan bool
	Cancel chan bool
	Finish chan bool
}

func NewRemote(task tasks.Task, blocker blocker.Blocker) Remote {
	return Remote{
		Task:    task,
		Blocker: blocker,
		Wg:      &sync.WaitGroup{},
		Pause:   make(chan bool, 1),
		Cancel:  make(chan bool, 1),
		Finish:  make(chan bool, 1),
	}
}

func (remote *Remote) RunTasks(withScreenRecorder bool) {
	// run the configured goroutines
	remote.Wg.Add(2)

	log.Println("Rendering progress bar")
	go RenderProgressBar(remote)

	log.Println("Polling input")
	go PollInput(remote)

	if withScreenRecorder {
		remote.Wg.Add(1)
		log.Println("Starting screen recorder")
		go FfmpegCaptureScreen(remote)
	}

	// wait for the goroutines to finish
	remote.Wg.Wait()
}
