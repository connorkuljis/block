package db

import (
	"fmt"

	"github.com/connorkuljis/block-cli/internal/buckets"
	"github.com/connorkuljis/block-cli/internal/config"
	"github.com/connorkuljis/block-cli/internal/tasks"
	"github.com/jmoiron/sqlx"
)

func InitDB() (*sqlx.DB, error) {
	db, err := sqlx.Connect("sqlite", config.GetDBPath())
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(buckets.BucketsSchema)
	if err != nil {
		return nil, fmt.Errorf("Error initalising db schema: %w", err)
	}

	_, err = db.Exec(tasks.TasksSchema)
	if err != nil {
		return nil, fmt.Errorf("Error initalising db schema: %w", err)
	}

	return db, nil
}
