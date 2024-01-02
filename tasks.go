package main

import (
	"database/sql"
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

// Task represents the structure of the database table
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

var schema = `
CREATE TABLE IF NOT EXISTS Tasks(
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
);
`

func InitDB() {
	db_file := filepath.Join(config.AppInfo.AppDir, "app_data.db")

	if flags.Verbose {
		log.Println("dbfile: " + db_file)
	}

	var err error
	db, err = sqlx.Connect("sqlite3", db_file)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(schema)
	if err != nil {
		log.Fatal(err)
	}
}

func NewTask(inName string, inDuration float64) *Task {
	return &Task{
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

func InsertTask(task *Task) *Task {
	insertQuery := `
		INSERT INTO Tasks (name, planned_duration_minutes, blocker_enabled, screen_enabled, screen_url, created_at, completed, completion_percent)
		VALUES (:name, :planned_duration_minutes, :blocker_enabled, :screen_enabled, :screen_url, :created_at, :completed, :completion_percent)
	`

	result, err := db.NamedExec(insertQuery, task)
	if err != nil {
		log.Fatal(err)
	}

	lastInsertID, err := result.LastInsertId()
	if err != nil {
		log.Fatal(err)
	}

	if flags.Verbose {
		fmt.Printf("Last Inserted ID: %d\n", lastInsertID)
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

func GetAllTasks() []Task {
	var tasks []Task
	err := db.Select(&tasks, "SELECT * FROM Tasks")
	if err != nil {
		log.Fatal(err)
	}
	return tasks
}

func UpdateTask(inTask *Task) {
	updateQuery := `UPDATE Tasks SET 
finished_at=:finished_at,
actual_duration_minutes=:actual_duration_minutes,
completed=:completed,
completion_percent:=completion_percent
WHERE id = :id`

	result, err := db.NamedExec(updateQuery, inTask)
	if err != nil {
		log.Fatal(err)
	}

	lastInsertID, err := result.LastInsertId()
	if err != nil {
		log.Fatal(err)
	}

	if flags.Verbose {
		fmt.Printf("Last Updated ID: %d\n", lastInsertID)
	}
}

func UpdateTaskVodByID(id int64, filename string) {
	updateQuery := `
	UPDATE Tasks SET 
	screen_url = $1 
	WHERE id = $2`

	result, err := db.Exec(updateQuery, filename, id)
	if err != nil {
		log.Fatal(err)
	}

	lastInsertID, err := result.LastInsertId()
	if err != nil {
		log.Fatal(err)
	}

	if flags.Verbose {
		fmt.Printf("Updated vod: %d\n", lastInsertID)
	}
}

func DeleteTaskByID(id string) {
	q := `DELETE FROM Tasks WHERE id = $1`

	result, err := db.Exec(q, id)
	if err != nil {
		log.Fatal(err)
	}

	r, err := result.RowsAffected()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Deleted task %s, (%d rows affected.)\n", id, r)
}

func boolToInt(cond bool) int {
	var v int
	if cond {
		v = 1
	}
	return v
}
