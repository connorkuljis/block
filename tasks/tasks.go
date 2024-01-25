package tasks

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/connorkuljis/block-cli/config"
	"github.com/connorkuljis/block-cli/utils"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
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

var Schema = `CREATE TABLE IF NOT EXISTS Tasks(
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

var db *sqlx.DB

func InitDB() error {
	var err error

	db, err = sqlx.Connect("sqlite", config.GetDBPath())
	if err != nil {
		return err
	}

	_, err = db.Exec(Schema)
	if err != nil {
		return err
	}
	return nil
}

func NewTask(inName string, inDuration float64, blockerEnabled bool, screenEnabled bool) Task {
	return Task{
		Name:              inName,
		PlannedDuration:   inDuration,
		ActualDuration:    sql.NullFloat64{Valid: false},
		BlockerEnabled:    utils.BoolToInt(blockerEnabled),
		ScreenEnabled:     utils.BoolToInt(screenEnabled),
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

	err := db.Select(&tasks, "SELECT * FROM Tasks ORDER BY created_at ASC")
	if err != nil {
		return tasks, nil
	}

	return tasks, nil
}

func GetAllCompletedTasks() ([]Task, error) {
	var tasks []Task

	err := db.Select(&tasks, "SELECT * FROM Tasks WHERE completed = 1 AND actual_duration_minutes > 5 ORDER BY created_at ASC")
	if err != nil {
		return tasks, nil
	}

	return tasks, nil
}

func UpdateFinishTimeAndDuration(task Task, inFinishedAt time.Time, inActuralDuration time.Duration) error {
	query := "UPDATE Tasks SET finished_at = ?, actual_duration_minutes = ? WHERE id = ?"

	finishedAt := sql.NullTime{
		Time:  inFinishedAt,
		Valid: true,
	}

	actualDuration := sql.NullFloat64{
		Float64: inActuralDuration.Minutes(),
		Valid:   true,
	}

	result, err := db.Exec(query, finishedAt, actualDuration, task.ID)
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

func UpdateCompletionPercent(inTask Task, inCompletionPercent float64) error {
	query := "UPDATE Tasks SET completion_percent = ?, completed = ? WHERE id = ?"

	completionPercent := sql.NullFloat64{
		Float64: inCompletionPercent,
		Valid:   true,
	}

	completed := 0
	if inCompletionPercent == 100.0 {
		completed = 1
	}

	result, err := db.Exec(query, completionPercent, completed, inTask.ID)
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
