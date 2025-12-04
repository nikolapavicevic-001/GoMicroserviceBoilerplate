package templating

import (
	"html/template"
	"os"
	"path/filepath"
	"strings"
)

type LoadedTemplates struct {
	Pages    map[string]*template.Template
	Partials map[string]*template.Template
}

func LoadTemplates(funcMap template.FuncMap) (*LoadedTemplates, error) {
	pages := make(map[string]*template.Template)
	partials := make(map[string]*template.Template)

	// Load layouts (shared for all pages)
	layoutFiles, err := getHTMLFiles("templates/layouts")
	if err != nil {
		return nil, err
	}

	// Load pages with layouts included
	pageFiles, err := getHTMLFiles("templates/pages")
	if err != nil {
		return nil, err
	}

	for _, page := range pageFiles {
		name := filepath.Base(page)

		files := append(layoutFiles, page)

		tmpl := template.Must(
			template.New(name).Funcs(funcMap).ParseFiles(files...),
		)

		pages[name] = tmpl
	}

	// Load partials (standalone)
	partialFiles, err := getHTMLFiles("templates/partials")
	if err != nil {
		return nil, err
	}

	for _, partial := range partialFiles {
		// Key: partials/user-stats.html
		key := partial[len("templates/"):]

		tmpl := template.Must(
			template.New(filepath.Base(partial)).Funcs(funcMap).ParseFiles(partial),
		)

		partials[key] = tmpl
	}

	return &LoadedTemplates{
		Pages:    pages,
		Partials: partials,
	}, nil
}

// Utility: recursively collect HTML files
func getHTMLFiles(dir string) ([]string, error) {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".html") {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

