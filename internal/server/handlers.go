package server

import (
	"fmt"
	"io/fs"
	"net/http"
	"strconv"
	"text/template"

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

	return nil
}

func (s *Server) HandleIndex() http.HandlerFunc {
	indexTemplateFragments := []string{
		s.TemplateFragments.Base["root.html"],
		s.TemplateFragments.Base["layout.html"],
		s.TemplateFragments.Base["head.html"],
		s.TemplateFragments.Components["header.html"],
		s.TemplateFragments.Components["footer.html"],
		s.TemplateFragments.Components["nav.html"],
		s.TemplateFragments.Components["tasks.html"],
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
		tasks, err := tasks.GetAllTasks()
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
