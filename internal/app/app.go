package app

import (
	"io"
	"log"
	"time"

	"github.com/connorkuljis/block-cli/internal/interactive"
	"github.com/connorkuljis/block-cli/internal/tasks"
	"github.com/connorkuljis/block-cli/pkg/blocker"
)

func Start(w io.Writer, duration float64, taskname string, block bool, capture bool, debug bool) error {
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

	totalTimeSeconds, percent := interactive.RunTasks(w, currentTask, b)

	finishTime := time.Now()

	currentTask.SetCompletionPercent(percent)
	currentTask.UpdateFinishTime(finishTime)
	currentTask.UpdateActualDuration(totalTimeSeconds)

	log.Println("Finish Time (time):", currentTask.FinishedAt.Time)
	log.Println("Completion Percent (%):", currentTask.CompletionPercent.Float64)
	log.Println("Total Time (seconds):", totalTimeSeconds)
	log.Println("Total Time (minutes):", currentTask.ActualDuration.Float64)

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
