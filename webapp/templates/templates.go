package templates

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/kirsle/go-website-template/webapp/config"
	"github.com/kirsle/go-website-template/webapp/log"
	"github.com/kirsle/go-website-template/webapp/session"
)

// Template is a logical HTML template for the app with ability to wrap around an html/template
// and provide middlewares, hooks or live reloading capability in debug mode.
type Template struct {
	filename string    // Filename on disk (index.html)
	filepath string    // Full path on disk (./web/templates/index.html)
	modified time.Time // Modification date of the file at init time
	tmpl     *template.Template
}

// LoadTemplate processes and returns a template. Filename is relative
// to the template directory, e.g. "index.html". Call this at the initialization
// of your endpoint controller; in debug mode the template HTML from disk may be
// reloaded if modified after initial load.
func LoadTemplate(filename string) (*Template, error) {
	filepath := config.TemplatePath + "/" + filename
	stat, err := os.Stat(filepath)
	if err != nil {
		return nil, fmt.Errorf("LoadTemplate(%s): %s", filename, err)
	}

	files := templates(config.TemplatePath + "/" + filename)
	tmpl := template.New("page")
	tmpl.Funcs(TemplateFuncs(nil))
	if _, err := tmpl.ParseFiles(files...); err != nil {
		return nil, err
	}

	return &Template{
		filename: filename,
		filepath: filepath,
		modified: stat.ModTime(),
		tmpl:     tmpl,
	}, nil
}

// Must LoadTemplate or panic.
func Must(filename string) *Template {
	tmpl, err := LoadTemplate(filename)
	if err != nil {
		panic(err)
	}
	return tmpl
}

// Execute a loaded template. In debug mode, the template file may be reloaded
// from disk if the file on disk has been modified.
func (t *Template) Execute(w http.ResponseWriter, r *http.Request, vars map[string]interface{}) error {
	if vars == nil {
		vars = map[string]interface{}{}
	}

	// Merge in global variables.
	MergeVars(r, vars)
	MergeUserVars(r, vars)

	// Merge the flashed messsage variables in.
	if r != nil {
		sess := session.Get(r)
		flashes, errors := sess.ReadFlashes(w)
		vars["Flashes"] = flashes
		vars["Errors"] = errors
	}

	// Reload the template from disk?
	if stat, err := os.Stat(t.filepath); err == nil {
		if stat.ModTime().After(t.modified) {
			log.Info("Template(%s).Execute: file updated on disk, reloading", t.filename)
			err = t.Reload()
			if err != nil {
				log.Error("Reloading error: %s", err)
			}
		}
	}

	// Install the function map.
	tmpl := t.tmpl
	if r != nil {
		tmpl = t.tmpl.Funcs(TemplateFuncs(r))
	}

	if err := tmpl.ExecuteTemplate(w, "base", vars); err != nil {
		log.Error("Template error: %s", err)
		return err
	}

	return nil
}

// Reload the template from disk.
func (t *Template) Reload() error {
	stat, err := os.Stat(t.filepath)
	if err != nil {
		return fmt.Errorf("Reload(%s): %s", t.filename, err)
	}

	files := templates(t.filepath)
	tmpl := template.New("page")
	tmpl.Funcs(TemplateFuncs(nil))
	tmpl.ParseFiles(files...)

	t.tmpl = tmpl
	t.modified = stat.ModTime()
	return nil
}

// Base template layout.
var baseTemplates = []string{
	config.TemplatePath + "/base.html",
	// config.TemplatePath + "/partials/user_avatar.html",
	// mix in other partials here
}

// templates returns a template chain with the base templates preceding yours.
// Files given are expected to be full paths (config.TemplatePath + file)
func templates(files ...string) []string {
	return append(baseTemplates, files...)
}

// RenderTemplate executes a template. Filename is relative to the templates
// root, e.g. "index.html"
func RenderTemplate(w io.Writer, r *http.Request, filename string, vars map[string]interface{}) error {
	if vars == nil {
		vars = map[string]interface{}{}
	}

	// Merge in user vars.
	MergeVars(r, vars)
	MergeUserVars(r, vars)

	files := templates(config.TemplatePath + "/" + filename)
	tmpl := template.Must(
		template.New("index").ParseFiles(files...),
	)

	err := tmpl.ExecuteTemplate(w, "base", vars)
	if err != nil {
		return err
	}

	return nil
}
