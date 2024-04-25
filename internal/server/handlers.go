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
	s.MuxRouter.HandleFunc("/", s.HandleIndex())
	s.MuxRouter.HandleFunc("/api/tasks", s.HandleTasks())

	return nil
}

func (s *Server) HandleTasks() http.HandlerFunc {
	funcMap := template.FuncMap{
		"secsToMinSec": func(secs int64) string {
			minutes := secs / 60
			seconds := secs % 60

			minutesStr := strconv.Itoa(int(minutes))
			if minutes < 10 {
				minutesStr = "0" + minutesStr
			}
			secondsStr := strconv.Itoa(int(seconds))
			if seconds < 10 {
				secondsStr = "0" + secondsStr
			}

			formattedString := fmt.Sprintf("%s:%s", minutesStr, secondsStr)
			return formattedString
		},
	}

	tasksTemplatePartial := s.BuildTemplates("tasks-partial", funcMap, s.TemplateFragments.Components["tasks-table.html"])

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			r.ParseForm()
			strStartDate := r.FormValue("start_date")
			strEndDate := r.FormValue("end_date")
			if strStartDate == "" {
				http.Error(w, "error: start date must not be empty", http.StatusInternalServerError)
				return
			}

			layout := "2006-01-02"

			// Parse the string into a time.Time object
			startTime, err := time.Parse(layout, strStartDate)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			endTime, err := time.Parse(layout, strEndDate)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			tasks, err := tasks.GetTasksByDateRange(s.Db, startTime, endTime)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			data := map[string]any{
				"Tasks": tasks,
			}

			htmlBytes, err := SafeTmplExec(tasksTemplatePartial, "tasks-table", data)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			SendHTML(w, htmlBytes)
			return
		}
	}
}

func (s *Server) HandleIndex() http.HandlerFunc {
	indexTemplateFragments := []string{
		s.TemplateFragments.Base["root.html"],
		s.TemplateFragments.Base["layout.html"],
		s.TemplateFragments.Base["head.html"],
		s.TemplateFragments.Components["header.html"],
		s.TemplateFragments.Components["footer.html"],
		s.TemplateFragments.Components["nav.html"],
		s.TemplateFragments.Components["tasks-table.html"],
		s.TemplateFragments.Views["index.html"],
	}

	funcMap := template.FuncMap{
		"secsToMinSec": func(secs int64) string {
			minutes := secs / 60
			seconds := secs % 60

			minutesStr := strconv.Itoa(int(minutes))
			if minutes < 10 {
				minutesStr = "0" + minutesStr
			}
			secondsStr := strconv.Itoa(int(seconds))
			if seconds < 10 {
				secondsStr = "0" + secondsStr
			}

			formattedString := fmt.Sprintf("%s:%s", minutesStr, secondsStr)
			return formattedString
		},
	}

	indexTemplate := s.BuildTemplates("index", funcMap, indexTemplateFragments...)

	return func(w http.ResponseWriter, r *http.Request) {
		tasks, err := tasks.GetAllTasks(s.Db)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		htmlBytes, err := SafeTmplExec(indexTemplate, "root", map[string]interface{}{
			"Tasks": tasks,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		SendHTML(w, htmlBytes)
	}
}
