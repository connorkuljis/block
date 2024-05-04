package app

import (
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/connorkuljis/block-cli/internal/blocker"
	"github.com/connorkuljis/block-cli/internal/interactive"
	"github.com/connorkuljis/block-cli/internal/tasks"
	"github.com/jmoiron/sqlx"
)

func Start(w io.Writer, db *sqlx.DB, currentTask tasks.Task) error {
	blocker := blocker.NewBlocker()
	if currentTask.BlockerEnabled == 1 {
		n, err := blocker.Start()
		if err != nil {
			return err
		}
		slog.Info(fmt.Sprintf("Blocker started (%d bytes written).", n))
	}

	err := tasks.InsertTask(db, &currentTask)
	if err != nil {
		return err
	}

	totalTimeSeconds, percent := interactive.Run(w, &currentTask, blocker, db)
	finishTime := time.Now()

	currentTask.SetActualDuration(totalTimeSeconds)
	currentTask.SetCompletionPercent(percent)
	currentTask.SetFinishTime(finishTime)

	err = tasks.UpdateTaskAsFinished(db, currentTask)
	if err != nil {
		return err
	}

	if currentTask.BlockerEnabled == 1 {
		n, err := blocker.Stop()
		if err != nil {
			return err
		}
		slog.Info(fmt.Sprintf("Blocker stopped (%d bytes written).", n))
	}

	return nil
}
