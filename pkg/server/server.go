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
	FileSystem embed.FS
	Router     *http.ServeMux
	Templates  Templates
	AppData    AppData // global app data

	Port         string
	StaticDir    string // location of static assets
	TemplatesDir string // location of html templates, makes template parsing less verbose.
}

type Templates struct {
	Components Components
	Views      Views

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
	Morning string
}

type Views struct {
	Index string
}

type AppData struct {
	Title   string
	DevMode bool
}

const (
	StaticDirName    = "www/static"
	TemplatesDirName = "www/templates"
)

//go:embed www/templates/* www/static/*
var embedFS embed.FS

// NewServer returns a new pointer Server struct.
//
// Server encapsulates all dependencies for the web Server.
// HTTP handlers access information via receiver types.
func NewServer(port string) *Server {
	s := &Server{
		FileSystem:   embedFS,
		Router:       http.NewServeMux(),
		Port:         port,
		TemplatesDir: TemplatesDirName,
		StaticDir:    StaticDirName,
	}

	s.Templates = LoadTemplatesFrom(s.FileSystem)

	return s
}

// LoadTemplatesFrom returns a new Templates struct
//
// Templates encapulates all definitions of html files to load from the Server.
func LoadTemplatesFrom(filesys fs.FS) Templates {
	return Templates{
		Head:   MustNewTemplateFile(filesys, "head.html"),
		Layout: MustNewTemplateFile(filesys, "layout.html"),
		Root:   MustNewTemplateFile(filesys, "root.html"),
		Components: Components{
			DevTool: MustNewComponentFile(filesys, "dev-tool.html"),
			Header:  MustNewComponentFile(filesys, "header.html"),
			Nav:     MustNewComponentFile(filesys, "nav.html"),
			Footer:  MustNewComponentFile(filesys, "footer.html"),
			Tasks:   MustNewComponentFile(filesys, "tasks.html"),
			Morning: MustNewComponentFile(filesys, "morning.html"),
		},
		Views: Views{
			Index: MustNewViewFile(filesys, "index.html"),
		},
	}
}

func CheckFileExists(filesystem fs.FS, target string) error {
	_, err := filesystem.Open(target)
	if err != nil {
		return err
	}
	return nil
}

func MustNewTemplateFile(filesys fs.FS, filename string) string {
	layoutFile := filepath.Join(TemplatesDirName, filename)
	err := CheckFileExists(filesys, layoutFile)
	if err != nil {
		log.Fatal("New Layout File error: ", err)
	}

	return layoutFile
}

func MustNewComponentFile(filesys fs.FS, filename string) string {
	layoutFile := filepath.Join(TemplatesDirName, "components", filename)
	err := CheckFileExists(filesys, layoutFile)
	if err != nil {
		log.Fatal("New Component File error: ", err)
	}

	return layoutFile
}

func MustNewViewFile(filesys fs.FS, filename string) string {
	layoutFile := filepath.Join(TemplatesDirName, "views", filename)
	err := CheckFileExists(filesys, layoutFile)
	if err != nil {
		log.Fatal("New View File error: ", err)
	}

	return layoutFile
}

// getIndexTemplate parses joined base and index view templates.
func IndexTemplate(s *Server) *template.Template {
	return BuildTemplates(s, "index.html", nil,
		s.Templates.Head,
		s.Templates.Root,
		s.Templates.Layout,
		s.Templates.Components.DevTool,
		s.Templates.Components.Header,
		s.Templates.Components.Footer,
		s.Templates.Components.Nav,
		s.Templates.Components.Tasks,
		s.Templates.Components.Morning,
		s.Templates.Views.Index,
	)
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
		log.Fatalf("Error building template: %s %s", name, err)
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
