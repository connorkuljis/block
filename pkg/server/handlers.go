package server

import (
	"io/fs"
	"log"
	"net/http"

	"github.com/connorkuljis/block-cli/internal/app"
	"github.com/connorkuljis/block-cli/internal/tasks"
)

// Routes instatiates http Handlers and associated patterns on the server.
func (s *Server) Routes() error {
	scfs, err := fs.Sub(s.FileSystem, s.StaticDir) // static content sub fs from the server's embedded fs
	if err != nil {
		return err
	}

	s.Router.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(scfs))))
	s.Router.HandleFunc("/", s.HandleIndex())
	s.Router.HandleFunc("/start", s.HandleStart())
	s.Router.HandleFunc("/greeting", s.HandleGreeting())

	return nil
}

func (s *Server) HandleIndex() http.HandlerFunc {
	tmpl := IndexTemplate(s)

	data := map[string]interface{}{
		"AppData": s.AppData,
		"Tasks":   nil,
	}

	return func(w http.ResponseWriter, r *http.Request) {
		tasks, err := tasks.GetAllTasks()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		data["Tasks"] = tasks

		buf, err := SafeTmplExec(tmpl, "root", data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		SendHTML(w, buf)
	}
}

func (s *Server) HandleStart() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// r.ParseForm()

		// name := r.Form.Get("taskname")
		// duration := r.Form.Get("duration")

		// durationFloat, err := strconv.ParseFloat(duration, 64)
		// if err != nil {
		// 	http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		// }

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		durationFloat := 0.1
		name := "test"
		err := app.Start(w, flusher, durationFloat, name, true, false, false)
		if err != nil {
			log.Print(err)
		}
	}
}

func (s *Server) HandleGreeting() http.HandlerFunc {
	tmpl := BuildTemplates(s, "greeting", nil, s.Templates.Components.Morning)

	return func(w http.ResponseWriter, r *http.Request) {
		b, err := SafeTmplExec(tmpl, "greeting", nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		SendHTML(w, b)
	}
}
