package web

import (
	"html/template"
	"net/http"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/egggo/inbucket/log"
)

var cachedMutex sync.Mutex
var cachedTemplates = map[string]*template.Template{}
var cachedPartials = map[string]*template.Template{}

// RenderTemplate fetches the named template and renders it to the provided
// ResponseWriter.
func RenderTemplate(name string, w http.ResponseWriter, data interface{}) error {
	t, err := ParseTemplate(name, false)
	if err != nil {
		log.LogError("Error in template '%v': %v", name, err)
		return err
	}
	w.Header().Set("Expires", "-1")
	return t.Execute(w, data)
}

// RenderPartial fetches the named template and renders it to the provided
// ResponseWriter.
func RenderPartial(name string, w http.ResponseWriter, data interface{}) error {
	t, err := ParseTemplate(name, true)
	if err != nil {
		log.LogError("Error in template '%v': %v", name, err)
		return err
	}
	w.Header().Set("Expires", "-1")
	return t.Execute(w, data)
}

// ParseTemplate loads the requested template along with _base.html, caching
// the result (if configured to do so)
func ParseTemplate(name string, partial bool) (*template.Template, error) {
	cachedMutex.Lock()
	defer cachedMutex.Unlock()

	if t, ok := cachedTemplates[name]; ok {
		return t, nil
	}

	tempPath := strings.Replace(name, "/", string(filepath.Separator), -1)
	tempFile := filepath.Join(webConfig.TemplateDir, tempPath)
	log.LogTrace("Parsing template %v", tempFile)

	var err error
	var t *template.Template
	if partial {
		// Need to get basename of file to make it root template w/ funcs
		base := path.Base(name)
		t = template.New(base).Funcs(TemplateFuncs)
		t, err = t.ParseFiles(tempFile)
	} else {
		t = template.New("_base.html").Funcs(TemplateFuncs)
		t, err = t.ParseFiles(filepath.Join(webConfig.TemplateDir, "_base.html"), tempFile)
	}
	if err != nil {
		return nil, err
	}

	// Allows us to disable caching for theme development
	if webConfig.TemplateCache {
		if partial {
			log.LogTrace("Caching partial %v", name)
			cachedTemplates[name] = t
		} else {
			log.LogTrace("Caching template %v", name)
			cachedTemplates[name] = t
		}
	}

	return t, nil
}
