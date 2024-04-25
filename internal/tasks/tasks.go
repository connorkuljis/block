package tasks

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/connorkuljis/block-cli/internal/config"
	"github.com/connorkuljis/block-cli/internal/utils"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

type Task struct {
	TaskId                   int64           `db:"task_id"`
	TaskName                 string          `db:"task_name"`
	EstimatedDurationSeconds int64           `db:"estimated_duration_seconds"`
	ActualDurationSeconds    sql.NullInt64   `db:"actual_duration_seconds"`
	BlockerEnabled           int             `db:"blocker_enabled"`
	ScreenEnabled            int             `db:"screen_enabled"`
	ScreenURL                sql.NullString  `db:"screen_url"`
	CreatedAt                time.Time       `db:"created_at"`
	FinishedAt               sql.NullTime    `db:"finished_at"`
	Completed                int             `db:"completed"`
	CompletionPercent        sql.NullFloat64 `db:"completion_percent"`
}

var db *sqlx.DB

var Schema = `CREATE TABLE IF NOT EXISTS Tasks(
    task_id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_name TEXT NOT NULL,
    estimated_duration_seconds INTEGER NOT NULL,
    actual_duration_seconds INTEGER,
    blocker_enabled INTEGER DEFAULT 0,
    screen_enabled INTEGER DEFAULT 0,
    screen_url TEXT,
    created_at TIMESTAMP NOT NULL,
    finished_at TIMESTAMP,
    completed INTEGER,
    completion_percent REAL
)
`

func InitDB() error {
	var err error

	db, err = sqlx.Connect("sqlite", config.GetDBPath())
	if err != nil {
		return err
	}

	_, err = db.Exec(Schema)
	if err != nil {
		return fmt.Errorf("Error initalising db schema: %w", err)
	}

	return nil
}

func NewTask(taskName string, durationSeconds int64, blockerEnabled bool, screenEnabled bool, createdAt time.Time) *Task {
	return &Task{
		TaskName:                 taskName,
		EstimatedDurationSeconds: durationSeconds,
		ActualDurationSeconds:    sql.NullInt64{Valid: false},
		BlockerEnabled:           utils.BoolToInt(blockerEnabled),
		ScreenEnabled:            utils.BoolToInt(screenEnabled),
		ScreenURL:                sql.NullString{Valid: false},
		CreatedAt:                createdAt,
		FinishedAt:               sql.NullTime{Valid: false},
		Completed:                0,
		CompletionPercent:        sql.NullFloat64{Valid: false},
	}
}

func (task *Task) SetCompletionPercent(completionPercent float64) {
	if completionPercent == 100.0 {
		task.Completed = 1
	}

	task.CompletionPercent = sql.NullFloat64{
		Valid:   true,
		Float64: completionPercent,
	}
}

func (task *Task) UpdateFinishTime(finishedAt time.Time) {
	task.FinishedAt = sql.NullTime{Time: finishedAt, Valid: true}
}

func (task *Task) UpdateActualDuration(actualDurationSeconds int) {
	task.ActualDurationSeconds = sql.NullInt64{Int64: int64(actualDurationSeconds), Valid: true}
}

func InsertTask(task *Task) {
	insertQuery := `INSERT INTO Tasks (
	task_name, 
	estimated_duration_seconds, 
	blocker_enabled, 
	screen_enabled, 
	screen_url, 
	created_at, 
	completed, 
	completion_percent) 
	VALUES (
	:task_name, 
	:estimated_duration_seconds, 
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

	task.TaskId = lastInsertID
}

func GetTaskByID(id int64) Task {
	var task Task
	err := db.Get(&task, "SELECT * FROM Tasks WHERE task_id = ?", id)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Retrieved Record: %+v\n", task)
	return task
}

func GetAllTasks() ([]Task, error) {
	var tasks []Task

	rows, err := db.Queryx("SELECT * FROM Tasks ORDER BY created_at DESC")
	if err != nil {
		log.Fatal(err)
	}

	for rows.Next() {
		var t Task
		err = rows.StructScan(&t)
		if err != nil {
			log.Fatal(err)
		}
		tasks = append(tasks, t)
	}

	// err := db.Select(&tasks, "SELECT * FROM Tasks")
	// if err != nil {
	// 	return tasks, nil
	// }

	return tasks, nil
}

func GetAllCompletedTasks() ([]Task, error) {
	var tasks []Task

	err := db.Select(&tasks, "SELECT * FROM Tasks WHERE completed = 1 AND actual_duration_seconds > 5 ORDER BY created_at ASC")
	if err != nil {
		return tasks, nil
	}

	return tasks, nil
}

func UpdateTaskAsFinished(task Task) error {
	query := "UPDATE Tasks SET finished_at = ?, actual_duration_seconds = ?, completion_percent = ?, completed = ? WHERE task_id = ?"

	result, err := db.Exec(query, task.FinishedAt, task.ActualDurationSeconds, task.CompletionPercent, task.Completed, task.TaskId)
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
	query := "UPDATE Tasks SET screen_url = ? WHERE task_id = ?"

	result, err := db.Exec(query, target, task.TaskId)
	if err != nil {
		return err
	}

	_, err = result.LastInsertId()
	if err != nil {
		return err
	}

	return nil
}

func UpdateCompletionPercent(inTask Task, inCompletionPercent float64) error {
	query := "UPDATE Tasks SET completion_percent = ?, completed = ? WHERE task_id = ?"

	completionPercent := sql.NullFloat64{
		Float64: inCompletionPercent,
		Valid:   true,
	}

	// allow nullable for timer/stopwatch
	if inCompletionPercent < 0 {
		completionPercent.Valid = false
	}

	completed := 0
	if inCompletionPercent == 100.0 {
		completed = 1
	}

	result, err := db.Exec(query, completionPercent, completed, inTask.TaskId)
	if err != nil {
		return err
	}

	_, err = result.LastInsertId()
	if err != nil {
		return err
	}

	return nil
}

func DeleteTaskByID(id string) (int64, error) {
	query := "DELETE FROM Tasks WHERE task_id = ?"
	var rowsAffected int64

	result, err := db.Exec(query, id)
	if err != nil {
		return rowsAffected, err
	}

	rowsAffected, err = result.RowsAffected()
	if err != nil {
		return rowsAffected, err
	}

	return rowsAffected, nil
}

func GetTasksByDate(inDate time.Time) ([]Task, error) {
	query := `SELECT * FROM Tasks WHERE DATE(created_at) = DATE(?)`

	var tasks []Task

	err := db.Select(&tasks, query, inDate)
	if err != nil {
		return tasks, err
	}

	return tasks, nil
}

func GetCapturedTasksByDate(inDate time.Time) ([]Task, error) {
	query := `SELECT * FROM Tasks 
	WHERE DATE(created_at) = DATE(?)
	AND screen_enabled = 1
	AND completed = 1`

	var tasks []Task

	err := db.Select(&tasks, query, inDate)
	if err != nil {
		return tasks, nil
	}

	return tasks, nil
}
