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
func (s *Server) Routes() error {
	s.MuxRouter.Handle("/static/", http.StripPrefix("/static/", s.StaticContentHandler))
	s.MuxRouter.HandleFunc("/", s.HandleHome())
	s.MuxRouter.HandleFunc("/tasks", s.HandleTasks())
	s.MuxRouter.HandleFunc("/tasks/{taskId}/show", s.HandleShowTasks())
	s.MuxRouter.HandleFunc("/tasks/{taskId}/edit", s.HandleEditTasks())
	s.MuxRouter.HandleFunc("/buckets", s.HandleBuckets())
	return nil
}

func (s *Server) HandleHome() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/tasks", http.StatusSeeOther)
		return
	}
}

type TaskEditForm struct {
	taskName string
	hours    string
	minutes  string
	seconds  string
}

type SanitisedTaskEditForm struct {
	hours   int
	minutes int
	seconds int
}

func validateForm(form TaskEditForm) (SanitisedTaskEditForm, error) {
	var sanitisedForm SanitisedTaskEditForm
	type FormField struct {
		String string
		Result int
		Max    int
	}

	fields := []FormField{
		{String: form.hours, Max: 99},
		{String: form.minutes, Max: 59},
		{String: form.seconds, Max: 59},
	}

	for i := range fields {
		res, err := strconv.Atoi(fields[i].String)
		if err != nil {
			return sanitisedForm, err
		}
		if res > fields[i].Max {
			return sanitisedForm, fmt.Errorf("Error, input exeeds maximum allowed value")
		}
		fields[i].Result = res
	}

	sanitisedForm.hours = fields[0].Result
	sanitisedForm.minutes = fields[1].Result
	sanitisedForm.seconds = fields[2].Result

	return sanitisedForm, nil
}

func (s *Server) HandleEditTasks() http.HandlerFunc {
	editPage := []string{
		"root.html",
		"head.html",
		"layout.html",
		"header.html",
		"nav.html",
		"footer.html",
		"edit_tasks.html",
	}

	editPageTemplate := s.ParseTemplates("edit-tasks", funcMap, editPage...)

	return func(w http.ResponseWriter, r *http.Request) {
		strTaskId := r.PathValue("taskId")
		taskId, err := strconv.Atoi(strTaskId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if r.Method == "POST" {
			r.ParseForm()
			var form TaskEditForm
			form.taskName = r.FormValue("taskname")
			form.hours = r.FormValue("hours")
			form.minutes = r.FormValue("minutes")
			form.seconds = r.FormValue("seconds")

			sanitisedForm, err := validateForm(form)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			totalSeconds := sanitisedForm.hours*3600 + sanitisedForm.minutes*60 + sanitisedForm.seconds

			err = tasks.UpdateTaskFinishById(s.Db, int64(taskId), form.taskName, int64(totalSeconds))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			http.Redirect(w, r, fmt.Sprintf("/tasks/%d/show", taskId), http.StatusSeeOther)
			return
		}

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

		parcel := summariseTasks(tasks)

		var htmlBytes []byte
		switch r.Header.Get("HX-Request") {
		case "true":
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

func summariseTasks(tasks []tasks.Task) map[string]any {
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

	return map[string]any{
		"Tasks":                        tasks,
		"TaskCount":                    taskCount,
		"TaskTotalSeconds":             taskTotalSeconds,
		"TaskAverageSeconds":           taskAverageSeconds,
		"TaskAverageCompletionPercent": taskAverageCompletionPercent,
	}
}
