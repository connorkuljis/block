package app

import (
	"io"
	"log"
	"net/http"
	"time"

	"github.com/connorkuljis/block-cli/internal/interactive"
	"github.com/connorkuljis/block-cli/internal/tasks"
	"github.com/connorkuljis/block-cli/pkg/blocker"
)

func Start(w io.Writer, flusher http.Flusher, duration float64, taskname string, block bool, capture bool, debug bool) error {
	// TODO: check duration for errors

	b := blocker.NewBlocker()
	if block {
		err := b.Start()
		if err != nil {
			return err
		}
		log.Println("Blocker started.")

	}

	startTime := time.Now()

	currentTask := tasks.NewTask(taskname, duration, block, capture, startTime)
	tasks.InsertTask(currentTask)

	percent := interactive.RunTasks(w, flusher, currentTask, b)
	log.Println("Percent: ", percent)

	currentTask.SetCompletionPercent(percent)

	finishTime := time.Now()
	currentTask.SetFinishTime(finishTime)

	err := tasks.UpdateTaskAsFinished(*currentTask)
	if err != nil {
		return err
	}

	if block {
		err = b.Stop()
		if err != nil {
			return err
		}
		log.Println("Blocker stopped.")
	}

	return nil
}
