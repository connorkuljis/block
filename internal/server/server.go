package server

import (
	"bytes"
	"errors"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"path/filepath"

	"github.com/jmoiron/sqlx"
)

// Server encapsulates all dependencies for the web Server.
// HTTP handlers access information via receiver types.
type Server struct {
	FileSystem        fs.FS
	MuxRouter         *http.ServeMux
	Db                *sqlx.DB
	TemplateFragments TemplateFragments

	Port string
}

type TemplateFragments struct {
	Base       map[string]string
	Components map[string]string
	Views      map[string]string
}

const (
	TemplatesDirStr  = "www/templates"
	StaticDirStr     = "www/static"
	ComponentsDirStr = "components"
	ViewsDirStr      = "views"
)

// NewServer returns a new pointer Server struct.
//
// Server encapsulates all dependencies for the web Server.
// HTTP handlers access information via receiver types.
func NewServer(fileSystem fs.FS, db *sqlx.DB, port string) (*Server, error) {
	templateFragments, err := ExtractTemplateFragmentsFromFilesystem(fileSystem)
	if err != nil {
		return nil, err
	}

	s := &Server{
		FileSystem:        fileSystem,
		MuxRouter:         http.NewServeMux(),
		Db:                db,
		Port:              port,
		TemplateFragments: templateFragments,
	}

	return s, nil
}

func (s *Server) ListenAndServe() error {
	log.Println("[ 💿 Spinning up server on http://localhost:" + s.Port + " ]")
	if err := http.ListenAndServe(":"+s.Port, s.MuxRouter); err != nil {
		log.Println("Error starting server.")
		return err
	}

	return nil
}

// ExtractTemplateFragmentsFromFilesystem traverses the base, components and views directory in the given filesystem and returns a Fragments structure, or an error if an error occurs.
func ExtractTemplateFragmentsFromFilesystem(filesystem fs.FS) (TemplateFragments, error) {
	var templateFragments TemplateFragments
	var err error

	// load root templates
	templatesPath := TemplatesDirStr
	templateFragments.Base, err = buildFilePathMap(filesystem, templatesPath)
	if err != nil {
		return templateFragments, err
	}

	// load components templates
	componentsPath := filepath.Join(TemplatesDirStr, ComponentsDirStr)
	templateFragments.Components, err = buildFilePathMap(filesystem, componentsPath)
	if err != nil {
		return templateFragments, err
	}

	// load views templates
	viewsPath := filepath.Join(TemplatesDirStr, ViewsDirStr)
	templateFragments.Views, err = buildFilePathMap(filesystem, viewsPath)
	if err != nil {
		return templateFragments, err
	}

	return templateFragments, nil
}

// buildFilePathMap reads the filepath of all regular files into a map, keyed by the filename
func buildFilePathMap(filesystem fs.FS, path string) (map[string]string, error) {
	filePathMap := make(map[string]string)

	files, err := fs.ReadDir(filesystem, path)
	if err != nil {
		return filePathMap, err
	}

	for _, file := range files {
		if !file.IsDir() {
			name := file.Name()               // "index.html"
			path := filepath.Join(path, name) // "www/static/templates/views/index/html"
			filePathMap[name] = path          // "index.html" => "www/static/templates/views/index/html"
		}
	}

	return filePathMap, nil
}

// buildTemplates is a fast way to parse a collection of templates in the server filesystem.
//
// template files are provided as strings to be parsed from the filesystem
func (s *Server) BuildTemplates(name string, funcs template.FuncMap, templates ...string) *template.Template {
	for _, template := range templates {
		if template == "" {
			log.Fatal(errors.New("Error building template for (" + name + "): an empty template was detected..."))
		}
	}
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
	var bufBytes bytes.Buffer
	tmpl.Option("missingkey=error")
	err := tmpl.ExecuteTemplate(&bufBytes, name, data)
	if err != nil {
		log.Print(err)
		return bufBytes.Bytes(), err
	}
	return bufBytes.Bytes(), nil
}

// sendHTML writes a buffer a response writer as html
func SendHTML(w http.ResponseWriter, bufBytes []byte) {
	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	_, err := w.Write(bufBytes)
	if err != nil {
		log.Println(err)
	}
}
