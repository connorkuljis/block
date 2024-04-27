package server

import (
	"fmt"
	"io/fs"
	"net/http"
	"strconv"
	"text/template"
	"time"

	"github.com/connorkuljis/block-cli/internal/tasks"
)

// Routes instatiates http Handlers and associated patterns on the server.
func (s *Server) Routes() error {
	scfs, err := fs.Sub(s.FileSystem, StaticDirStr) // static content sub fs from the server's embedded fs
	if err != nil {
		return err
	}

	s.MuxRouter.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(scfs))))
	s.MuxRouter.HandleFunc("/tasks", s.HandleTasks())

	return nil
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

	tasksTemplateFragments := []string{
		s.TemplateFragments.Base["root.html"],
		s.TemplateFragments.Base["layout.html"],
		s.TemplateFragments.Base["head.html"],
		s.TemplateFragments.Components["header.html"],
		s.TemplateFragments.Components["footer.html"],
		s.TemplateFragments.Components["nav.html"],
		s.TemplateFragments.Components["tasks-table.html"],
		s.TemplateFragments.Components["tasks-date-filter.html"],
		s.TemplateFragments.Views["index.html"],
	}

	tasksTemplate := s.BuildTemplates("index", funcMap, tasksTemplateFragments...)
	tasksTemplatePartial := s.BuildTemplates("tasks-partial", funcMap, s.TemplateFragments.Components["tasks-table.html"])

	return func(w http.ResponseWriter, r *http.Request) {
		var daysBack = 30
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

		data := SummariseTasks(tasks)

		var htmlBytes []byte
		switch r.Header.Get("HX-Request") {
		case "true":
			htmlBytes, err = SafeTmplExec(tasksTemplatePartial, "tasks-table", data)
		default:
			htmlBytes, err = SafeTmplExec(tasksTemplate, "root", data)
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		SendHTML(w, htmlBytes)
		return
	}
}

func SummariseTasks(tasks []tasks.Task) map[string]any {
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

	data := map[string]interface{}{"Tasks": tasks, "TaskCount": taskCount, "TaskTotalSeconds": taskTotalSeconds, "TaskAverageSeconds": taskAverageSeconds, "TaskAverageCompletionPercent": taskAverageCompletionPercent}
	return data
}
