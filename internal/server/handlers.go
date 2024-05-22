package server

import (
	"fmt"
	"net/http"
	"strconv"
	"text/template"
	"time"

	"github.com/connorkuljis/block-cli/internal/buckets"
	"github.com/connorkuljis/block-cli/internal/tasks"
)

var funcMap = template.FuncMap{
	"PrintTimeHHMMSS": func(secs int64) string {
		hours := secs / 3600
		minutes := (secs % 3600) / 60
		seconds := secs % 60
		return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
	},
	"ParseTimeHHMMSS": func(secs int64) map[string]int64 {
		hours := secs / 3600
		minutes := (secs % 3600) / 60
		seconds := secs % 60
		return map[string]int64{
			"Hours":   hours,
			"Minutes": minutes,
			"Seconds": seconds,
		}
	},
}

// Routes instatiates http Handlers and associated patterns on the server.
func (s *Server) Routes() {
	s.MuxRouter.Handle("/static/", http.StripPrefix("/static/", s.StaticContentHandler))
	s.MuxRouter.HandleFunc("/", s.HandleHome())
	s.MuxRouter.HandleFunc("/tasks", s.HandleTasks())
	s.MuxRouter.HandleFunc("/tasks/show/{taskId}", s.HandleShowTasks())
	s.MuxRouter.HandleFunc("/tasks/edit/{taskId}", s.HandleEditTasks())
	s.MuxRouter.HandleFunc("/daily/", s.HandleDaily())
	s.MuxRouter.HandleFunc("/buckets", s.HandleBuckets())
}

func (s *Server) HandleHome() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/tasks", http.StatusSeeOther)
		return
	}
}

func (s *Server) HandleEditTasks() http.HandlerFunc {
	editPage := []string{"root.html", "head.html", "layout.html", "header.html", "nav.html", "footer.html", "edit_tasks.html"}

	editPageTemplate := s.ParseTemplates("edit-tasks", funcMap, editPage...)

	return func(w http.ResponseWriter, r *http.Request) {
		strTaskId := r.PathValue("taskId")
		taskId, err := strconv.ParseInt(strTaskId, 10, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		switch r.Method {
		case "GET":
			task, err := tasks.GetTaskByID(s.Db, int64(taskId))
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			parcel := map[string]interface{}{"Task": task}
			htmlBytes, err := SafeTmplExec(editPageTemplate, "root", parcel)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			SendHTML(w, htmlBytes)
		case "POST":
			r.ParseForm()
			taskName := r.FormValue("taskname")
			hours, err := strconv.Atoi(r.FormValue("hours"))
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			minutes, err := strconv.Atoi(r.FormValue("minutes"))
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			seconds, err := strconv.Atoi(r.FormValue("seconds"))
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if minutes > 59 || seconds > 59 {
				http.Error(w, "Error, minutes or seconds value must not exceed 59", http.StatusBadRequest)
			}
			totalSeconds := int64(hours*3600 + minutes*60 + seconds)
			err = tasks.UpdateTaskFinishById(s.Db, int64(taskId), taskName, totalSeconds)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, fmt.Sprintf("/tasks/show/%d", taskId), http.StatusSeeOther)
			return
		default:
			fmt.Fprintln(w, "Unsupported request type")
			return
		}
	}
}

func (s *Server) HandleShowTasks() http.HandlerFunc {
	page := []string{
		"root.html",
		"head.html",
		"layout.html",
		"header.html",
		"nav.html",
		"footer.html",
		"show_tasks.html",
	}

	t := s.ParseTemplates("show-tasks", funcMap, page...)

	return func(w http.ResponseWriter, r *http.Request) {
		strTaskId := r.PathValue("taskId")
		taskId, err := strconv.Atoi(strTaskId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		task, err := tasks.GetTaskByID(s.Db, int64(taskId))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		parcel := map[string]interface{}{"Task": task}

		htmlBytes, err := SafeTmplExec(t, "root", parcel)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		SendHTML(w, htmlBytes)
	}
}
func (s *Server) HandleDaily() http.HandlerFunc {
	templates := []string{
		"root.html",
		"layout.html",
		"head.html",
		"header.html",
		"footer.html",
		"nav.html",
		"tasks-table.html",
		"daily.html",
	}

	t := s.ParseTemplates("daily", funcMap, templates...)

	return func(w http.ResponseWriter, r *http.Request) {
		timestamp := r.URL.Query().Get("created_at")
		format := "2006-01-02"

		var dateCurrent time.Time
		if timestamp != "" {
			var err error
			dateCurrent, err = time.Parse(format, timestamp)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		} else {
			dateCurrent = time.Now().Truncate(24 * time.Hour)
		}

		datePrev := dateCurrent.Add(-24 * time.Hour)

		// TODO: validate if overflows current date. if so, don't display the control in the html
		dateNext := dateCurrent.Add(24 * time.Hour)

		tasks, err := tasks.GetTasksByDate(s.Db, dateCurrent)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		taskSummary := summariseTasks(tasks)

		parcel := map[string]any{
			"Tasks":       tasks,
			"DateCurrent": dateCurrent.Format(format),
			"DatePrev":    datePrev.Format(format),
			"DateNext":    dateNext.Format(format),
			"TaskSummary": taskSummary,
		}

		htmlBytes, err := SafeTmplExec(t, "root", parcel)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		SendHTML(w, htmlBytes)
	}
}

func (s *Server) HandleBuckets() http.HandlerFunc {
	bucketsTemplateFragments := []string{
		"root.html",
		"layout.html",
		"head.html",
		"header.html",
		"footer.html",
		"nav.html",
		"buckets.html",
	}

	bucketsTemplate := s.ParseTemplates("index", nil, bucketsTemplateFragments...)

	return func(w http.ResponseWriter, r *http.Request) {
		buckets, err := buckets.GetAllBuckets(s.Db)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		parcel := map[string]any{"Buckets": buckets}

		htmlBytes, err := SafeTmplExec(bucketsTemplate, "root", parcel)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		SendHTML(w, htmlBytes)
	}
}

func (s *Server) HandleTasks() http.HandlerFunc {
	tasksPageTemplateFragments := []string{
		"root.html",
		"layout.html",
		"head.html",
		"header.html",
		"footer.html",
		"nav.html",
		"form-get-tasks.html",
		"tasks-table.html",
		"index.html",
	}

	taskPartialTemplateFragment := "tasks-table.html"

	tasksPage := s.ParseTemplates("tasks-page", funcMap, tasksPageTemplateFragments...)

	tasksPartial := s.ParseTemplates("tasks-partial", funcMap, taskPartialTemplateFragment)

	return func(w http.ResponseWriter, r *http.Request) {
		var daysBack = 7
		if strPastDays := r.URL.Query().Get("past"); strPastDays != "" {
			if parsedDays, err := strconv.Atoi(strPastDays); err == nil {
				daysBack = parsedDays
			} else {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		}

		tasks, err := tasks.GetRecentTasks(s.Db, time.Now().Truncate(24*time.Hour), daysBack)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		taskSummary := summariseTasks(tasks)

		parcel := map[string]any{
			"Tasks":       tasks,
			"TaskSummary": taskSummary,
		}

		var htmlBytes []byte
		switch r.Header.Get("HX-Target") {
		case "tasks_body":
			htmlBytes, err = SafeTmplExec(tasksPartial, "tasks-table", parcel)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		default:
			htmlBytes, err = SafeTmplExec(tasksPage, "root", parcel)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		SendHTML(w, htmlBytes)
	}
}

type TasksSummary struct {
	TaskCount                    int64
	TaskTotalSeconds             int64
	TaskAverageSeconds           int64
	TaskAverageCompletionPercent float64
}

func summariseTasks(tasks []tasks.Task) TasksSummary {
	var taskCount int64
	var taskTotalSeconds int64
	var taskAverageSeconds int64
	var taskTotalCompletionPercent float64
	var taskAverageCompletionPercent float64

	taskCount = int64(len(tasks))

	if taskCount > 0 {
		for i := range tasks {
			taskTotalSeconds += tasks[i].ActualDurationSeconds.Int64
			taskTotalCompletionPercent += tasks[i].CompletionPercent.Float64
		}
		taskAverageSeconds = taskTotalSeconds / taskCount
		taskAverageCompletionPercent = float64(taskTotalCompletionPercent) / float64(taskCount)
	}

	return TasksSummary{
		TaskCount:                    taskCount,
		TaskTotalSeconds:             taskTotalSeconds,
		TaskAverageSeconds:           taskAverageSeconds,
		TaskAverageCompletionPercent: taskAverageCompletionPercent,
	}
}
