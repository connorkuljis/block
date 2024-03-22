package server

import (
	"bytes"
	"embed"
	"io/fs"
	"log"
	"net/http"
	"path/filepath"
	"text/template"
)

// Server encapsulates all dependencies for the web Server.
// HTTP handlers access information via receiver types.
type Server struct {
	FileSystem fs.FS // in-memory or disk
	Router     *http.ServeMux
	Templates  Templates
	AppData    AppData // global app data

	Port         string
	StaticDir    string // location of static assets
	TemplatesDir string // location of html templates, makes template parsing less verbose.
}

type Templates struct {
	BaseLayout BaseLayout
	Components Components
	Views      Views
}

type BaseLayout struct {
	Root   string
	Head   string
	Layout string
}

type Components struct {
	DevTool string
	Header  string
	Nav     string
	Footer  string
	Tasks   string
}

type Views struct {
	Index    string
	Projects string
}

type AppData struct {
	Title   string
	DevMode bool
}

const (
	StaticDirName    = "www/static"
	TemplatesDirName = "www/templates"
)

//go:embed www/templates www/static
var embedFS embed.FS

// NewServer returns a new pointer Server struct.
//
// Server encapsulates all dependencies for the web Server.
// HTTP handlers access information via receiver types.
func NewServer(port string) *Server {
	return &Server{
		FileSystem:   embedFS,
		Router:       http.NewServeMux(),
		Port:         port,
		TemplatesDir: TemplatesDirName,
		StaticDir:    StaticDirName,
		Templates:    NewTemplates(TemplatesDirName),
	}
}

// NewTemplates returns a new Templates struct
//
// Templates encapulates all definitions of html files to load from the Server.
func NewTemplates(dir string) Templates {
	return Templates{
		BaseLayout: BaseLayout{
			Root:   filepath.Join(dir, "root.html"),
			Head:   filepath.Join(dir, "head.html"),
			Layout: filepath.Join(dir, "layout.html"),
		},

		Components: Components{
			DevTool: filepath.Join(dir, "components", "dev-tool.html"),
			Header:  filepath.Join(dir, "components", "header.html"),
			Nav:     filepath.Join(dir, "components", "nav.html"),
			Footer:  filepath.Join(dir, "components", "footer.html"),
			Tasks:   filepath.Join(dir, "components", "tasks.html"),
		},
		Views: Views{
			Index: filepath.Join(dir, "views", "index.html"),
		},
	}
}

// getIndexTemplate parses joined base and index view templates.
func IndexTemplate(s *Server) *template.Template {
	view := []string{
		s.Templates.BaseLayout.Head,
		s.Templates.BaseLayout.Root,
		s.Templates.BaseLayout.Layout,
		s.Templates.Components.DevTool,
		s.Templates.Components.Header,
		s.Templates.Components.Footer,
		s.Templates.Components.Nav,
		s.Templates.Components.Tasks,
		s.Templates.Views.Index,
	}

	return BuildTemplates(s, "index.html", nil, view...)
}

// buildTemplates is a fast way to parse a collection of templates in the server filesystem.
func BuildTemplates(s *Server, name string, funcs template.FuncMap, templates ...string) *template.Template {
	// give the template a name
	tmpl := template.New(name)

	// custom template functions if exists
	if funcs != nil {
		tmpl.Funcs(funcs)
	}

	// generate a template from the files in the server fs (usually embedded)
	tmpl, err := tmpl.ParseFS(s.FileSystem, templates...)
	if err != nil {
		log.Fatal(err)
	}

	return tmpl
}

// safeTmplParse executes a given template to a bytes buffer. It returns the resulting buffer or nil, err if any error occurred.
//
// Templates are checked for missing keys to prevent partial data being written to the writer.
func SafeTmplExec(tmpl *template.Template, name string, data any) ([]byte, error) {
	var buf bytes.Buffer
	tmpl.Option("missingkey=error")
	err := tmpl.ExecuteTemplate(&buf, name, data)
	if err != nil {
		log.Print(err)
		return buf.Bytes(), err
	}
	return buf.Bytes(), nil
}

// sendHTML writes a buffer a response writer as html
func SendHTML(w http.ResponseWriter, buf []byte) {
	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	_, err := w.Write(buf)
	if err != nil {
		log.Println(err)
	}
}
