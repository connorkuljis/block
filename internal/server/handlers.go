package server

import (
	"fmt"
	"io/fs"
	"net/http"
	"strconv"
	"text/template"
	"time"

	"github.com/connorkuljis/block-cli/internal/buckets"
	"github.com/connorkuljis/block-cli/internal/tasks"
)

// Routes instatiates http Handlers and associated patterns on the server.
func (s *Server) Routes() error {
	scfs, err := fs.Sub(s.FileSystem, StaticDirStr) // static content sub fs from the server's embedded fs
	if err != nil {
		return err
	}

	s.MuxRouter.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(scfs))))
	s.MuxRouter.HandleFunc("/", s.HandleHome())
	s.MuxRouter.HandleFunc("/tasks", s.HandleTasks())
	s.MuxRouter.HandleFunc("/buckets", s.HandleBuckets())

	return nil
}

func (s *Server) HandleHome() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/tasks", http.StatusSeeOther)
		return
	}
}

func (s *Server) HandleBuckets() http.HandlerFunc {
	bucketsTemplateFragments := []string{
		s.TemplateFragments.Base["root.html"],
		s.TemplateFragments.Base["layout.html"],
		s.TemplateFragments.Base["head.html"],
		s.TemplateFragments.Components["header.html"],
		s.TemplateFragments.Components["footer.html"],
		s.TemplateFragments.Components["nav.html"],
		s.TemplateFragments.Views["buckets.html"],
	}

	bucketsTemplate := s.BuildTemplates("index", nil, bucketsTemplateFragments...)

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
		return
	}
}

func (s *Server) HandleTasks() http.HandlerFunc {
	funcMap := template.FuncMap{
		"secsToHHMMSS": func(secs int64) string {
			hours := secs / 3600
			minutes := (secs % 3600) / 60
			seconds := secs % 60

			var formattedString string
			if hours > 0 {
				formattedString = fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
			} else {
				formattedString = fmt.Sprintf("%02d:%02d", minutes, seconds)
			}

			return formattedString
		},
	}

	tasksPageTemplateFragments := []string{
		s.TemplateFragments.Base["root.html"],
		s.TemplateFragments.Base["layout.html"],
		s.TemplateFragments.Base["head.html"],
		s.TemplateFragments.Components["header.html"],
		s.TemplateFragments.Components["footer.html"],
		s.TemplateFragments.Components["nav.html"],
		s.TemplateFragments.Components["form-get-tasks.html"],
		s.TemplateFragments.Components["tasks-table.html"],
		s.TemplateFragments.Views["index.html"],
	}

	taskPartialTemplateFragment := s.TemplateFragments.Components["tasks-table.html"]

	tasksPage := s.BuildTemplates("tasks-page", funcMap, tasksPageTemplateFragments...)

	tasksPartial := s.BuildTemplates("tasks-partial", funcMap, taskPartialTemplateFragment)

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

		tasks, err := tasks.GetRecentTasks(s.Db, time.Now(), daysBack)
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
		return
	}
}

func summariseTasks(tasks []tasks.Task) map[string]any {
	var taskCount int64
	var taskTotalSeconds int64
	var taskAverageSeconds int64
	var taskTotalCompletionPercent float64
	var taskAverageCompletionPercent float64

	taskCount = int64(len(tasks))

	for i := range tasks {
		taskTotalSeconds += tasks[i].ActualDurationSeconds.Int64
		taskTotalCompletionPercent += tasks[i].CompletionPercent.Float64
	}
	taskAverageSeconds = taskTotalSeconds / taskCount
	taskAverageCompletionPercent = float64(taskTotalCompletionPercent) / float64(taskCount)

	return map[string]any{
		"Tasks":                        tasks,
		"TaskCount":                    taskCount,
		"TaskTotalSeconds":             taskTotalSeconds,
		"TaskAverageSeconds":           taskAverageSeconds,
		"TaskAverageCompletionPercent": taskAverageCompletionPercent,
	}
}
