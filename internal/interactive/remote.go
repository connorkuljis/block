package interactive

import (
	"fmt"
	"io"
	"log/slog"
	"sync"

	"github.com/connorkuljis/block-cli/internal/blocker"
	"github.com/connorkuljis/block-cli/internal/tasks"
	"github.com/jmoiron/sqlx"
)

type Remote struct {
	Task    *tasks.Task
	Blocker blocker.Blocker
	Db      *sqlx.DB

	W                 io.Writer
	Wg                *sync.WaitGroup
	Pause             chan bool
	Cancel            chan bool
	Finish            chan error
	CompletionPercent chan float64
	TotalTimeSeconds  chan int
}

func Run(w io.Writer, task *tasks.Task, blocker blocker.Blocker, db *sqlx.DB) (int, float64) {
	remote := &Remote{
		Task:              task,
		Blocker:           blocker,
		Db:                db,
		Wg:                &sync.WaitGroup{},
		W:                 w,
		Pause:             make(chan bool, 1),
		Cancel:            make(chan bool, 1),
		Finish:            make(chan error, 1),
		CompletionPercent: make(chan float64, 1),
		TotalTimeSeconds:  make(chan int, 1),
	}

	remote.Wg.Add(2)

	slog.Info("Rendering progress bar")
	go RenderProgressBar(remote)

	slog.Info("Polling input")
	go PollInput(remote)

	if task.ScreenEnabled == 1 {
		remote.Wg.Add(1)
		slog.Info("Capturing screen.")
		go FfmpegCaptureScreen(remote)
	}

	fmt.Println("---")
	fmt.Println("Press [q] or [esc] or [control-C] to quit.")
	fmt.Println("Press [space] key to pause (re-enables sites temporarily).")

	remote.Wg.Wait()

	percent := <-remote.CompletionPercent
	totalTimeSeconds := <-remote.TotalTimeSeconds

	return totalTimeSeconds, percent
}
