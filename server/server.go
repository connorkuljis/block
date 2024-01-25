package server

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"path/filepath"
	"text/template"

	"github.com/connorkuljis/block-cli/tasks"
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
		Collection Collection
		Docket     Docket
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
		"seq": func(start, end int) []int {
			var result []int
			for i := start; i <= end; i++ {
				result = append(result, i)
			}
			return result
		},
	}

	return func(w http.ResponseWriter, r *http.Request) {
		t, err := tasks.GetAllCompletedTasks()
		if err != nil {
			log.Print(err)
		}

		grouped := groupByDate(t)

		totalMinutes := 0.0
		for _, day := range grouped {
			totalMinutes += day.TotalMinutes
		}

		docket := minuteToDocket(totalMinutes)

		data := PageData{
			Collection: grouped,
			Docket:     docket,
		}

		tmpl := compileTemplates("index.html", s, indexHTML, funcMap)

		tmpl.ExecuteTemplate(w, "root", data)
	}
}

type Collection []Day

type Day struct {
	Tasks []tasks.Task

	DateStr      string
	TotalMinutes float64
	Docket       Docket
}

type Docket struct {
	Hours   int
	Minutes int
}

// input:    ["01-01", "01-02", "02-01", "02-02",]
// output:   [["01-01", "01-02"], ["02-01", "02-02"]]
func groupByDate(items []tasks.Task) Collection {
	var collection Collection
	var currDay Day
	var prev tasks.Task
	for i, item := range items {
		if i == 0 {
			currDay.Tasks = append(currDay.Tasks, item)
			currDay.DateStr = item.CreatedAt.Format("Mon Jan 02 2006")
			currDay.TotalMinutes += item.ActualDuration.Float64
			prev = item
			continue
		}

		if item.CreatedAt.Day() == prev.CreatedAt.Day() {
			currDay.Tasks = append(currDay.Tasks, item)
			currDay.TotalMinutes += item.ActualDuration.Float64
			continue
		}

		currDay.Docket = minuteToDocket(currDay.TotalMinutes)
		collection = append(collection, currDay)

		currDay = Day{}
		currDay.Tasks = append(currDay.Tasks, item)
		currDay.DateStr = item.CreatedAt.Format("Mon Jan 02 2006")
		currDay.TotalMinutes += item.ActualDuration.Float64
		prev = item
	}

	// add the last array when none remaining
	currDay.Docket = minuteToDocket(currDay.TotalMinutes)
	collection = append(collection, currDay)

	return collection
}

func minuteToDocket(minutes float64) Docket {
	hours := int(minutes / 60)
	remainingMinutes := int(minutes) % 60

	return Docket{Hours: hours, Minutes: remainingMinutes}
}
