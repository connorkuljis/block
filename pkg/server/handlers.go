package server

import (
	"net/http"

	"github.com/connorkuljis/block-cli/internal/app"
	"github.com/connorkuljis/block-cli/internal/tasks"
)

// Routes instatiates http Handlers and associated patterns on the server.
func (s *Server) Routes() {
	s.Router.Handle("/static/", http.FileServer(http.FS(s.FileSystem)))
	s.Router.HandleFunc("/", s.HandleIndex())
	s.Router.HandleFunc("/start", s.HandleStart())
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
		w.Write([]byte("OK"))

		app := app.App{}
		app.InitServer("0.1", "test", true, false)
		app.Start()
		app.SaveAndExit()
	}
}
