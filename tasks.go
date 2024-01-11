package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Task struct {
	ID                int64           `db:"id"`
	Name              string          `db:"name"`
	PlannedDuration   float64         `db:"planned_duration_minutes"`
	ActualDuration    sql.NullFloat64 `db:"actual_duration_minutes"`
	BlockerEnabled    int             `db:"blocker_enabled"`
	ScreenEnabled     int             `db:"screen_enabled"`
	ScreenURL         sql.NullString  `db:"screen_url"`
	CreatedAt         time.Time       `db:"created_at"`
	FinishedAt        sql.NullTime    `db:"finished_at"`
	Completed         int             `db:"completed"`
	CompletionPercent sql.NullFloat64 `db:"completion_percent"`
}

var schema = `CREATE TABLE IF NOT EXISTS Tasks(
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    planned_duration_minutes REAL NOT NULL,
    actual_duration_minutes REAL,
    blocker_enabled INTEGER NOT NULL,
    screen_enabled INTEGER NOT NULL,
    screen_url TEXT,
    created_at TIMESTAMP NOT NULL,
    finished_at TIMESTAMP,
    completed INTEGER,
    completion_percent REAL
)`

func NewTask(inName string, inDuration float64) Task {
	return Task{
		Name:              inName,
		PlannedDuration:   inDuration,
		ActualDuration:    sql.NullFloat64{Valid: false},
		BlockerEnabled:    boolToInt(!flags.DisableBlocker),
		ScreenEnabled:     boolToInt(flags.ScreenRecorder),
		ScreenURL:         sql.NullString{Valid: false},
		CreatedAt:         time.Now(),
		FinishedAt:        sql.NullTime{Valid: false},
		Completed:         0,
		CompletionPercent: sql.NullFloat64{Valid: false},
	}
}

func InsertTask(task Task) Task {
	insertQuery := `INSERT INTO Tasks (
	name, 
	planned_duration_minutes, 
	blocker_enabled, 
	screen_enabled, 
	screen_url, 
	created_at, 
	completed, 
	completion_percent) 
	VALUES (
	:name, 
	:planned_duration_minutes, 
	:blocker_enabled, 
	:screen_enabled, 
	:screen_url, 
	:created_at, 
	:completed, 
	:completion_percent)`

	result, err := db.NamedExec(insertQuery, task)
	if err != nil {
		log.Fatal(err)
	}

	lastInsertID, err := result.LastInsertId()
	if err != nil {
		log.Fatal(err)
	}

	task.ID = lastInsertID

	return task
}

func GetTaskByID(id int64) Task {
	var task Task
	err := db.Get(&task, "SELECT * FROM Tasks WHERE id = ?", id)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Retrieved Record: %+v\n", task)
	return task
}

func GetAllTasks() ([]Task, error) {
	var tasks []Task

	err := db.Select(&tasks, "SELECT * FROM Tasks")
	if err != nil {
		return tasks, nil
	}

	return tasks, nil
}

func UpdateFinishTimeAndDuration(task Task, finishedAt time.Time, acturalDuration time.Duration) error {
	query := "UPDATE Tasks SET finished_at = ?, actual_duration_minutes = ? WHERE id = ?"

	parsedFinishedAt := sql.NullTime{
		Time:  finishedAt,
		Valid: true,
	}

	parsedActualDuration := sql.NullFloat64{
		Float64: acturalDuration.Minutes() * 100,
		Valid:   true,
	}

	result, err := db.Exec(query, parsedFinishedAt, parsedActualDuration, task.ID)
	if err != nil {
		return err
	}

	_, err = result.LastInsertId()
	if err != nil {
		return err
	}

	return nil
}

func UpdateScreenURL(task Task, target string) error {
	query := "UPDATE Tasks SET screen_url = ? WHERE id = ?"

	result, err := db.Exec(query, target, task.ID)
	if err != nil {
		return err
	}

	_, err = result.LastInsertId()
	if err != nil {
		return err
	}

	return nil
}

func UpdateCompletionPercent(task Task, completionPercent float64) error {
	query := "UPDATE Tasks SET completion_percent = ?, completed = ? WHERE id = ?"

	parsedCompletionPercent := sql.NullFloat64{
		Float64: completionPercent,
		Valid:   true,
	}

	completed := 0
	if completionPercent == 100.0 {
		completed = 1
	}

	result, err := db.Exec(query, parsedCompletionPercent, completed, task.ID)
	if err != nil {
		return err
	}

	_, err = result.LastInsertId()
	if err != nil {
		return err
	}

	return nil
}

func DeleteTaskByID(id string) error {
	query := "DELETE FROM Tasks WHERE id = ?"

	result, err := db.Exec(query, id)
	if err != nil {
		return err
	}

	_, err = result.RowsAffected()
	if err != nil {
		return err
	}

	return nil
}

func boolToInt(cond bool) int {
	var v int
	if cond {
		v = 1
	}
	return v
}
