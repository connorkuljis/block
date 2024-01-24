package server

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"path/filepath"
	"text/template"
	"time"

	"github.com/connorkuljis/task-tracker-cli/tasks"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/sessions"
)

const SessionName = "session"

// Server encapsulates all dependencies for the web server.
// HTTP handlers access information via receiver types.
type Server struct {
	Port         string
	Router       *chi.Mux
	TemplatesDir string // location of html templates, makes template parsing less verbose.
	StaticDir    string // location of static assets
	FileSystem   fs.FS  // in-memory or disk
	Sessions     *sessions.CookieStore
}

//go:embed templates/* static/*
var inMemoryFS embed.FS

type HTMLFile string

const (
	RootHTML   HTMLFile = "root.html"
	HeadHTML   HTMLFile = "head.html"
	LayoutHTML HTMLFile = "layout.html"
	HeroHTML   HTMLFile = "components/hero.html"
	FooterHTML HTMLFile = "components/footer.html"
)

func Serve() {
	port := "8080"
	router := chi.NewMux()
	store := sessions.NewCookieStore([]byte("special_key"))
	templateDir := "templates"
	staticDir := "static"

	log.Println("[ 💿 Spinning up server on http://localhost:" + port + " ]")

	s := Server{
		Router:       router,
		Port:         port,
		TemplatesDir: templateDir,
		StaticDir:    staticDir,
		FileSystem:   inMemoryFS,
		Sessions:     store,
	}

	s.routes()

	err := http.ListenAndServe(":"+s.Port, s.Router)
	if err != nil {
		panic(err)
	}
}

func compileTemplates(templateName string, s *Server, files []HTMLFile, funcMap template.FuncMap) *template.Template {
	var filenames []string
	for i := range files {
		currentFilename := string(files[i])
		filenames = append(filenames, filepath.Join(s.TemplatesDir, currentFilename))
	}

	tmpl, err := template.New(templateName).Funcs(funcMap).ParseFS(s.FileSystem, filenames...)
	if err != nil {
		panic(err)
	}

	return tmpl
}

func (s *Server) routes() {
	s.Router.Handle("/static/*", http.FileServer(http.FS(s.FileSystem)))
	s.Router.HandleFunc("/", s.handleIndex())
}

func (s *Server) handleIndex() http.HandlerFunc {
	type PageData struct {
		Tasks [][]tasks.Task

		Date         time.Time
		TotalMinutes float64
		NumTasks     int
	}

	var indexHTML = []HTMLFile{
		RootHTML,
		HeadHTML,
		LayoutHTML,
	}

	var funcMap = template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
		"mul": func(a, b int) int {
			return a * b
		},
	}

	return func(w http.ResponseWriter, r *http.Request) {
		date := time.Now()

		t, err := tasks.GetAllCompletedTasks()
		if err != nil {
			log.Print(err)
		}

		grouped := groupByDate(t)

		total := 0.0
		for i := range t {
			total += t[i].ActualDuration.Float64
		}

		data := PageData{
			Date:         date,
			Tasks:        grouped,
			TotalMinutes: total,
		}

		tmpl := compileTemplates("index.html", s, indexHTML, funcMap)

		tmpl.ExecuteTemplate(w, "root", data)
	}
}

type Day struct {
	Tasks []tasks.Task

	DateStr      string
	TotalMinutes int
}

type Collection []Day

// input:    ["01-01", "01-02", "02-01", "02-02",]
// output:   [["01-01", "01-02"], ["02-01", "02-02"]]
func groupByDate(items []tasks.Task) [][]tasks.Task {
	var groupedData [][]tasks.Task
	var currArr []tasks.Task
	var prev tasks.Task
	for i, item := range items {
		if i == 0 {
			currArr = append(currArr, item)
			prev = item
			continue
		}

		if item.CreatedAt.Day() == prev.CreatedAt.Day() {
			currArr = append(currArr, item)
			continue
		}

		groupedData = append(groupedData, currArr)
		currArr = []tasks.Task{}
		currArr = append(currArr, item)
		prev = item
	}

	// add the last array when none remaining
	groupedData = append(groupedData, currArr)

	return groupedData
}
