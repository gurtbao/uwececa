package templates

import (
	"embed"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"strings"
	"sync"
)

type Templates struct {
	mu      sync.RWMutex
	t       map[string]*template.Template
	embedFS embed.FS
}

func NewTemplates(fs embed.FS) *Templates {
	p := &Templates{
		mu:      sync.RWMutex{},
		t:       make(map[string]*template.Template),
		embedFS: fs,
	}

	p.loadAllTemplates()

	return p
}

func (p *Templates) loadAllTemplates() {
	templates := make(map[string]*template.Template)
	var fragmentPaths []string

	err := fs.WalkDir(p.embedFS, "templates", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".html") {
			return nil
		}
		if !strings.Contains(path, "fragments/") {
			return nil
		}
		name := strings.TrimPrefix(path, "templates/")
		name = strings.TrimSuffix(name, ".html")
		tmpl, err := template.New(name).ParseFS(p.embedFS, path)
		if err != nil {
			slog.Error("failed setting up template fragment", "error", err)
			os.Exit(1)
		}
		templates[name] = tmpl
		fragmentPaths = append(fragmentPaths, path)
		slog.Debug("loaded fragment", "name", name)
		return nil
	})
	if err != nil {
		slog.Error("failed walking template dir for fragments", "error", err)
		os.Exit(1)
	}

	err = fs.WalkDir(p.embedFS, "templates", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".html") {
			return nil
		}
		if strings.Contains(path, "fragments/") {
			return nil
		}
		if strings.Contains(path, "layouts/") {
			return nil
		}
		name := strings.TrimPrefix(path, "templates/")
		name = strings.TrimSuffix(name, ".html")

		templatePaths := []string{}
		templatePaths = append(templatePaths, "templates/layouts/*.html")
		templatePaths = append(templatePaths, fragmentPaths...)
		templatePaths = append(templatePaths, path)

		tmpl, err := template.New(name).ParseFS(p.embedFS, templatePaths...)
		if err != nil {
			slog.Error("failed setting up template", "error", err)
			os.Exit(1)
		}
		templates[name] = tmpl
		slog.Debug("loaded template", "name", "name")
		return nil
	})
	if err != nil {
		slog.Error("failed walking template dir", "error", err)
		os.Exit(1)
	}

	slog.Info("templates loaded", "count", len(templates))
	p.mu.Lock()
	defer p.mu.Unlock()
	p.t = templates
}

func (p *Templates) ExecuteBase(name string, w io.Writer, params any) error {
	return p.executeReload(name, w, "layouts/base", params)
}

func (p *Templates) ExecutePlain(name string, w io.Writer, params any) error {
	return p.executeReload(name, w, "", params)
}

func (p *Templates) Execute(name string, w io.Writer, base string, params any) error {
	return p.executeReload(name, w, base, params)
}

func (p *Templates) executeReload(name string, w io.Writer, base string, params any) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	tmpl, ok := p.t[name]
	if !ok {
		return fmt.Errorf("template not found: %s", name)
	}

	if base == "" {
		return tmpl.Execute(w, params)
	} else {
		return tmpl.ExecuteTemplate(w, base, params)
	}
}
