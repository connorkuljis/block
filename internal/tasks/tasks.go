package tasks

import (
	"database/sql"
	"fmt"
	"log"
	"time"

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
	Status                   sql.NullString  `db:"status"`
	BucketId                 sql.NullInt64   `db:"bucket_id"`
}

var TasksSchema = `
	CREATE TABLE IF NOT EXISTS Tasks(
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
    completion_percent REAL,
    status TEXT,
    bucket_id INTEGER,
    FOREIGN KEY (bucket_id) REFERENCES Buckets(bucket_id)
	);
`

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
		BucketId:                 sql.NullInt64{Valid: false},
	}
}

func (task *Task) AddBucketTag(bucketId int64) {
	task.BucketId = sql.NullInt64{Int64: bucketId, Valid: true}
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

func InsertTask(db *sqlx.DB, task *Task) error {
	insertQuery := `INSERT INTO Tasks (
	task_name, 
	estimated_duration_seconds, 
	blocker_enabled, 
	screen_enabled, 
	screen_url, 
	created_at, 
	completed, 
	completion_percent,
	bucket_id) 
	VALUES (
	:task_name, 
	:estimated_duration_seconds, 
	:blocker_enabled, 
	:screen_enabled, 
	:screen_url, 
	:created_at, 
	:completed, 
	:completion_percent,
	:bucket_id)
	`

	result, err := db.NamedExec(insertQuery, task)
	if err != nil {
		return err
	}

	lastInsertID, err := result.LastInsertId()
	if err != nil {
		return err
	}

	task.TaskId = lastInsertID

	return nil
}

func GetTaskByID(db *sqlx.DB, id int64) Task {
	var task Task
	err := db.Get(&task, "SELECT * FROM Tasks WHERE task_id = ?", id)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Retrieved Record: %+v\n", task)
	return task
}

func GetAllTasks(db *sqlx.DB) ([]Task, error) {
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

	return tasks, nil
}

func GetTasksByBucketId(db *sqlx.DB, bucketId int64) ([]Task, error) {
	var tasks []Task

	err := db.Select(&tasks, "SELECT * FROM Tasks WHERE bucket_id = ?", bucketId)
	if err != nil {
		log.Fatal(err)
	}

	return tasks, nil
}

func GetAllCompletedTasks(db *sqlx.DB) ([]Task, error) {
	var tasks []Task

	err := db.Select(&tasks, "SELECT * FROM Tasks WHERE completed = 1 AND actual_duration_seconds > 5 ORDER BY created_at ASC")
	if err != nil {
		return tasks, nil
	}

	return tasks, nil
}

func UpdateTaskAsFinished(db *sqlx.DB, task Task) error {
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

func UpdateScreenURL(db *sqlx.DB, task Task, target string) error {
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

func UpdateCompletionPercent(db *sqlx.DB, inTask Task, inCompletionPercent float64) error {
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

func DeleteTaskByID(db *sqlx.DB, id string) (int64, error) {
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

func GetTasksByDate(db *sqlx.DB, inDate time.Time) ([]Task, error) {
	query := `SELECT * FROM Tasks WHERE DATE(created_at) = DATE(?)`

	var tasks []Task

	err := db.Select(&tasks, query, inDate)
	if err != nil {
		return tasks, err
	}

	return tasks, nil
}

func GetTasksByDateRange(db *sqlx.DB, startDate, endDate time.Time) ([]Task, error) {
	query := `SELECT * FROM Tasks WHERE DATE(created_at) BETWEEN ? AND ?`

	var tasks []Task

	err := db.Select(&tasks, query, startDate, endDate)
	if err != nil {
		return tasks, err
	}

	return tasks, nil
}

func GetRecentTasks(db *sqlx.DB, startDate time.Time, daysBack int) ([]Task, error) {
	var tasks []Task

	query := `SELECT * FROM Tasks WHERE created_at >= ?`

	prevDate := startDate.AddDate(0, 0, -daysBack)

	err := db.Select(&tasks, query, prevDate)
	if err != nil {
		return tasks, err
	}

	return tasks, nil
}

func GetCapturedTasksByDate(db *sqlx.DB, inDate time.Time) ([]Task, error) {
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
