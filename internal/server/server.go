package server

import (
	"bytes"
	"fmt"
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
	FileSystem           fs.FS
	StaticContentHandler http.Handler
	MuxRouter            *http.ServeMux
	TemplateMap          map[string]string
	Db                   *sqlx.DB

	Port string
}

// NewServer returns a new pointer Server struct.
//
// Server encapsulates all dependencies for the web Server.
// HTTP handlers access information via receiver types.
func NewServer(fileSystem fs.FS, db *sqlx.DB, port, templatesPath, staticPath string) (*Server, error) {
	templateMap, err := BuildTemplateMap(fileSystem, templatesPath)
	if err != nil {
		return nil, err
	}
	scfs, err := fs.Sub(fileSystem, staticPath)
	if err != nil {
		return nil, err
	}
	s := &Server{
		FileSystem:           fileSystem,
		MuxRouter:            http.NewServeMux(),
		Port:                 port,
		StaticContentHandler: http.FileServer(http.FS(scfs)),
		TemplateMap:          templateMap,
		Db:                   db,
	}
	return s, nil
}

func BuildTemplateMap(filesystem fs.FS, templatesPath string) (map[string]string, error) {
	templates := make(map[string]string)

	err := fs.WalkDir(filesystem, templatesPath, func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() && filepath.Ext(d.Name()) == ".html" {
			templates[d.Name()] = path
		}
		return nil
	})
	if err != nil {
		return templates, err
	}

	return templates, nil
}

// buildTemplates is a fast way to parse a collection of templates in the server filesystem.
//
// template files are provided as strings to be parsed from the filesystem
func (s *Server) ParseTemplates(name string, funcs template.FuncMap, templateKeys ...string) *template.Template {
	tmpl := template.New(name)
	if funcs != nil {
		tmpl.Funcs(funcs)
	}

	var templatePaths []string
	for _, key := range templateKeys {
		path, ok := s.TemplateMap[key]
		if !ok {
			log.Fatalf("Error parsing template [%s], key does not exist [%s]", name, key)
		}
		templatePaths = append(templatePaths, path)
	}

	tmpl, err := tmpl.ParseFS(s.FileSystem, templatePaths...)
	if err != nil {
		err = fmt.Errorf("Error building template name='%s': %w", name, err)
		log.Fatal(err)
	}
	return tmpl
}

func (s *Server) ListenAndServe() error {
	log.Println("[ 💿 Spinning up server on http://localhost:" + s.Port + " ]")
	if err := http.ListenAndServe(":"+s.Port, s.MuxRouter); err != nil {
		return fmt.Errorf("Error starting server: %w", err)
	}
	return nil
}

//
// Utils
//
//

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
