package router

import (
	"html/template"
	"io"

	"github.com/kirsle/go-website-template/webapp/config"
)

// LoadTemplate processes and returns a template. Filename is relative
// to the template directory, e.g. "index.html"
func LoadTemplate(filename string) *template.Template {
	files := templates(config.TemplatePath + "/" + filename)
	tmpl := template.Must(template.New("page").ParseFiles(files...))
	return tmpl
}

// Default template funcs.
var defaultFuncs = template.FuncMap{}

// Base template layout.
var baseTemplates = []string{
	config.TemplatePath + "/base.html",
}

// templates returns a template chain with the base templates preceding yours.
// Files given are expected to be full paths (config.TemplatePath + file)
func templates(files ...string) []string {
	return append(baseTemplates, files...)
}

// RenderTemplate executes a template. Filename is relative to the templates
// root, e.g. "index.html"
func RenderTemplate(w io.Writer, filename string) error {
	files := templates(config.TemplatePath + "/" + filename)
	tmpl := template.Must(
		template.New("index").ParseFiles(files...),
	)

	err := tmpl.ExecuteTemplate(w, "base", map[string]interface{}{
		"Title": config.Title,
	})
	if err != nil {
		return err
	}

	return nil
}
