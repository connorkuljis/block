package app

import (
	"fmt"
	"io"
	"log"
	"time"

	"github.com/connorkuljis/block-cli/internal/blocker"
	"github.com/connorkuljis/block-cli/internal/interactive"
	"github.com/connorkuljis/block-cli/internal/tasks"
	"github.com/jmoiron/sqlx"
)

func Start(w io.Writer, db *sqlx.DB, durationSeconds int64, taskname string, block bool, capture bool, debug bool) error {
	blocker := blocker.NewBlocker()
	if block {
		fmt.Println("enabling blocker")
		err := blocker.Start()
		if err != nil {
			return err
		}
		log.Println("Blocker started.")
	}

	startTime := time.Now()

	currentTask := tasks.NewTask(taskname, durationSeconds, block, capture, startTime)

	tasks.InsertTask(db, currentTask)

	totalTimeSeconds, percent := interactive.RunTasks(w, currentTask, blocker, db)

	finishTime := time.Now()

	currentTask.SetCompletionPercent(percent)
	currentTask.UpdateFinishTime(finishTime)
	currentTask.UpdateActualDuration(totalTimeSeconds)

	log.Println("Finish Time (time):", currentTask.FinishedAt.Time)
	log.Println("Completion Percent (%):", currentTask.CompletionPercent.Float64)
	log.Println("Total Time (seconds):", totalTimeSeconds)
	log.Println("Total Time (minutes):", currentTask.ActualDurationSeconds.Int64)

	err := tasks.UpdateTaskAsFinished(db, *currentTask)
	if err != nil {
		return err
	}

	if block {
		err = blocker.Stop()
		if err != nil {
			return err
		}
		log.Println("Blocker stopped.")
	}

	return nil
}
